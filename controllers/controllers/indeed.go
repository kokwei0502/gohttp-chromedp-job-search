package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// IndeedDetail = indeed.com detail structure
type IndeedDetail struct {
	Title       string
	Company     string
	Location    string
	Salary      string
	DateCreated string
	Link        string
}

var (
	indeedURL                                                                               = "https://sg.indeed.com/jobs?q={jobsearch}&fromage=7"
	indeedBaseURL                                                                           = "https://sg.indeed.com"
	indeedDataList                                                                          []*IndeedDetail
	indeedMsg                                                                               string
	totalIndeedfound                                                                        int
	nodeIndeedMain, nodeIndeedSalary, nodeIndeedNext                                        []*cdp.Node
	indeedTitle, indeedCompany, indeedLocation, indeedSalary, indeedDateCreated, indeedLink string
	boolTitle, boolLocation, boolLink, boolNext                                             bool
	indeedNextPage                                                                          string
	indeedNextPageMap                                                                       map[string]string
	startTime                                                                               time.Time
	endTime                                                                                 time.Duration
)

// GetIndeedData = Scrape datas from indeed.com
func GetIndeedData(search string) (IndeedResult []*IndeedDetail, Message string, TotalIndeed int) {
	startTime = time.Now()

	// Replace search keywords to correct format for indeed.com
	searchString := strings.ReplaceAll(search, " ", "+")
	indeedURL = strings.Replace(indeedURL, "{jobsearch}", searchString, 1)

	// Create new context background
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Set the timeout limit
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// Start to navigate to indeed.com URL
	if err := chromedp.Run(ctx,
		chromedp.Navigate(indeedURL),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		indeedMsg = fmt.Sprintf("Navigation Error, Please Try Again Later!\n%v", err)
		return indeedDataList, indeedMsg, 0
	}

	// Looping to get data until timeout limit reached / total data limit reached / pages limit reached
	for {
		// Initial URL, is the base URL, will change if next page navigation needs
		indeedNextPage = indeedBaseURL

		// Get the main section data
		if err := chromedp.Run(ctx,
			chromedp.Nodes(`div.jobsearch-SerpJobCard`, &nodeIndeedMain, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			indeedMsg = fmt.Sprintf("Error to Get Main Division, Please Try Again Later\n%v", err)
			return indeedDataList, indeedMsg, totalIndeedfound
		}

		// If main section found, sum the total found
		totalIndeedfound += len(nodeIndeedMain)

		// If main section data > 0, get the detail data
		for i := 0; i < len(nodeIndeedMain); i++ {
			if err := chromedp.Run(ctx,
				chromedp.AttributeValue(`h2.title > a`, "title", &indeedTitle, &boolTitle, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Text(`span.company`, &indeedCompany, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.AttributeValue(`div.recJobLoc`, "data-rc-loc", &indeedLocation, &boolLocation, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Nodes(`span.salaryText`, &nodeIndeedSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Text(`span.date`, &indeedDateCreated, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.AttributeValue("h2.title > a", "href", &indeedLink, &boolLink, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Get Informations, Please Try Again Later\n%v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
			if len(nodeIndeedSalary) == 0 {
				indeedSalary = "No Stated"
			} else {
				if err := chromedp.Run(ctx,
					chromedp.Text(`span.salaryText`, &indeedSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				); err != nil {
					indeedMsg = "Error to Get Salary Division, Please Try Again!"
					return indeedDataList, indeedMsg, totalIndeedfound
				}
			}
			indeedLink = indeedBaseURL + indeedLink

			// Append data get to list
			indeedDataList = append(indeedDataList, &IndeedDetail{
				Title:       indeedTitle,
				Company:     indeedCompany,
				Location:    indeedLocation,
				Salary:      indeedSalary,
				DateCreated: indeedDateCreated,
				Link:        indeedLink,
			})

		}

		// If data limit reached will break the loop, else continue
		if totalIndeedfound > 500 {
			break
		} else {
			// Check the next page division availability
			if err := chromedp.Run(ctx,
				chromedp.Nodes(`ul.pagination-list li`, &nodeIndeedNext, chromedp.ByQueryAll, chromedp.AtLeast(0)),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Navigate to Next Page, Please Try Again!\n%v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
		}

		// Create temparory data list for next page navigation data
		var nodeIndeedTemp []*cdp.Node
		if len(nodeIndeedNext) > 1 {
			if err := chromedp.Run(ctx,
				chromedp.Nodes("a", &nodeIndeedTemp, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedNext[len(nodeIndeedNext)-1])),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Navigate to Next Page, Please Try Again!\n %v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
		}

		// If next page navigation data is found, will get the href link and return to initial URL to navigate
		// and continue to loop for getting data
		if len(nodeIndeedTemp) == 1 {
			if err := chromedp.Run(ctx,
				chromedp.Attributes("a", &indeedNextPageMap, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedNext[len(nodeIndeedNext)-1])),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Navigate to Next Page, Please Try Again!\n %v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
			if val, ok := indeedNextPageMap["aria-label"]; ok {
				if val == "Next" {
					indeedNextPage = indeedBaseURL + indeedNextPageMap["href"]
					if err := chromedp.Run(ctx,
						chromedp.Navigate(indeedNextPage),
						chromedp.Sleep(2*time.Second),
					); err != nil {
						indeedMsg = fmt.Sprintf("Failed to Get Next Page Navigation Button, Please Retry!\n%v", err)
						return indeedDataList, indeedMsg, totalIndeedfound
					}
				}
			}
		} else {
			break
		}
	}
	endTime = time.Since(startTime)

	// Return all results
	indeedMsg = fmt.Sprintf("Successfully Scrape Datas From indeed.com")
	fmt.Println(endTime)
	return indeedDataList, indeedMsg, totalIndeedfound
}
