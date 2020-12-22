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
	searchString := strings.ReplaceAll(search, " ", "+")
	indeedURL = strings.Replace(indeedURL, "{jobsearch}", searchString, 1)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(indeedURL),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		indeedMsg = "Navigation Error, Please Try Again Later!"
		return indeedDataList, indeedMsg, 0
	}
	for {
		indeedNextPage = indeedBaseURL
		if err := chromedp.Run(ctx,
			chromedp.Nodes(`div.jobsearch-SerpJobCard`, &nodeIndeedMain, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			indeedMsg = "Error to Get Main Division, Please Try Again Later"
			return indeedDataList, indeedMsg, 0
		}
		totalIndeedfound += len(nodeIndeedMain)
		for i := 0; i < len(nodeIndeedMain); i++ {
			if err := chromedp.Run(ctx,
				chromedp.AttributeValue(`h2.title > a`, "title", &indeedTitle, &boolTitle, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Text(`span.company`, &indeedCompany, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.AttributeValue(`div.recJobLoc`, "data-rc-loc", &indeedLocation, &boolLocation, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Nodes(`span.salaryText`, &nodeIndeedSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.Text(`span.date`, &indeedDateCreated, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				chromedp.AttributeValue("h2.title > a", "href", &indeedLink, &boolLink, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
			); err != nil {
				indeedMsg = "Error to Get Informations, Please Try Again Later"
				return indeedDataList, indeedMsg, 0
			}
			if len(nodeIndeedSalary) == 0 {
				indeedSalary = "No Stated"
			} else {
				if err := chromedp.Run(ctx,
					chromedp.Text(`span.salaryText`, &indeedSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedMain[i])),
				); err != nil {
					indeedMsg = "Error to Get Salary Division, Please Try Again!"
					return indeedDataList, indeedMsg, 0
				}
			}
			indeedLink = indeedBaseURL + indeedLink
			indeedDataList = append(indeedDataList, &IndeedDetail{
				Title:       indeedTitle,
				Company:     indeedCompany,
				Location:    indeedLocation,
				Salary:      indeedSalary,
				DateCreated: indeedDateCreated,
				Link:        indeedLink,
			})

		}

		if totalIndeedfound > 500 {
			break
		} else {
			if err := chromedp.Run(ctx,
				chromedp.Nodes(`ul.pagination-list li`, &nodeIndeedNext, chromedp.ByQueryAll, chromedp.AtLeast(0)),
			); err != nil {
				indeedMsg = "Error to Navigate to Next Page, Please Try Again!"
				return indeedDataList, indeedMsg, totalIndeedfound
			}
		}
		fmt.Println(len(nodeIndeedNext))
		var nodeIndeedTemp []*cdp.Node
		if len(nodeIndeedNext) > 1 {
			if err := chromedp.Run(ctx,
				chromedp.Nodes("a", &nodeIndeedTemp, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedNext[len(nodeIndeedNext)-1])),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Navigate to Next Page, Please Try Again!\n %v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
		}
		fmt.Println(len(nodeIndeedTemp))
		if len(nodeIndeedTemp) == 1 {
			if err := chromedp.Run(ctx,
				chromedp.Attributes("a", &indeedNextPageMap, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeIndeedNext[len(nodeIndeedNext)-1])),
			); err != nil {
				indeedMsg = fmt.Sprintf("Error to Navigate to Next Page, Please Try Again!\n %v", err)
				return indeedDataList, indeedMsg, totalIndeedfound
			}
			if val, ok := indeedNextPageMap["aria-label"]; ok {
				fmt.Println(val)
				if val == "Next" {
					indeedNextPage = indeedBaseURL + indeedNextPageMap["href"]
					if err := chromedp.Run(ctx,
						chromedp.Navigate(indeedNextPage),
						chromedp.Sleep(2*time.Second),
					); err != nil {
						indeedMsg = "Failed to Get Next Page Navigation Button, Please Retry!"
						return indeedDataList, indeedMsg, totalIndeedfound
					}
				} else {
					indeedMsg = "Successfully Scrape Datas From indeed.com"
					return indeedDataList, indeedMsg, totalIndeedfound
				}
			} else {
				indeedMsg = "Successfully Scrape Datas From indeed.com"
				return indeedDataList, indeedMsg, totalIndeedfound
			}
		} else {
			break
		}
	}
	endTime = time.Since(startTime)
	indeedMsg = "Successfully Scrape Datas From indeed.com"
	fmt.Println(endTime)
	return indeedDataList, indeedMsg, totalIndeedfound
}

// ctx, _ := context.WithTimeout(ctx, time.Second)
// ctx, _ = chromedp.NewContext(ctx)
// var cookies string
// if err := chromedp.Run(ctx,
// 	chromedp.Navigate(ts.URL),
// 	chromedp.Text("#cookies", &cookies),
// ); err != nil {
// 	panic(err)
// }
