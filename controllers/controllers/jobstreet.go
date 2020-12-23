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

var (
	nodeArticles, nodePage, nodeSpan                                                 []*cdp.Node
	jobstrTitle, jobstrCompany, jobstrLocation, jobstrSalary, jobstrLink, jobstrTime string
	totalJobstr                                                                      = 0
	jobStreetURL                                                                     = "https://www.jobstreet.com/en/job-search/{search-content}-jobs/?createdAt=7d"
	jobstrErrMsg                                                                     string
)

// JobStreetDetail = jobstreet.com data structure
type JobStreetDetail struct {
	Title       string
	Company     string
	Location    string
	Salary      string
	DateCreated string
	Link        string
}

// GetJobstreetData = Scrape data from jobstreet.com
func GetJobstreetData(search string) (JobStreetResult []*JobStreetDetail, Message string, TotalJobStreet int) {
	var jobstreetList []*JobStreetDetail
	// Replace search keywords to correct format
	searchContent := strings.ReplaceAll(search, " ", "-")
	jobStreetURL = strings.Replace(jobStreetURL, "{search-content}", searchContent, 1)

	// Create new context background
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Set the timeout limit
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// Start to navigate to jobstreet URL
	if err := chromedp.Run(ctx,
		chromedp.Navigate(jobStreetURL),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		jobstrErrMsg = fmt.Sprintf("Navigation Error, Please Try Again Later!\n%v", err)
		return jobstreetList, jobstrErrMsg, 0
	}

	// Looping to get data until timeout limit reached / total data limit reached / pages limit reached
	for {
		// Check the main section availability
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`article`),
			chromedp.Nodes(`article`, &nodeArticles, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			jobstrErrMsg = fmt.Sprintf("Error to Get Main Division, Please Try Again Later\n%v", err)
			return jobstreetList, jobstrErrMsg, 0
		}

		// Sum the total found
		totalJobstr += len(nodeArticles)

		// If main section found at least 1, start to get the detail data
		var ok bool
		for i := 0; i < len(nodeArticles); i++ {
			if err := chromedp.Run(ctx,
				chromedp.Text(`h1`, &jobstrTitle, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
				chromedp.Nodes(`span`, &nodeSpan, chromedp.ByQueryAll, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
				chromedp.AttributeValue("h1 > a", "href", &jobstrLink, &ok, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
			); err != nil {
				jobstrErrMsg = fmt.Sprintf("Error to Get Informations, Please Try Again Later\n%v", err)
				return jobstreetList, jobstrErrMsg, 0
			}
			switch len(nodeSpan[0].Children) {
			case 1:
				jobstrCompany = nodeSpan[0].Children[0].NodeValue
			case 0:
				jobstrCompany = ""
			}
			switch len(nodeSpan[2].Children) {
			case 1:
				jobstrLocation = nodeSpan[2].Children[0].NodeValue
			case 0:
				jobstrLocation = ""
			}
			switch len(nodeSpan) {
			case 5:
				switch len(nodeSpan[3].Children) {
				case 1:
					jobstrTime = nodeSpan[3].Children[0].NodeValue
					jobstrSalary = "No Stated"
				case 0:
					jobstrTime = ""
					jobstrSalary = "No Stated"
				}
			case 6:
				switch len(nodeSpan[3].Children) {
				case 1:
					jobstrSalary = nodeSpan[3].Children[0].NodeValue
				case 0:
					jobstrSalary = "No Stated"
				}
				switch len(nodeSpan[4].Children) {
				case 1:
					jobstrTime = nodeSpan[4].Children[0].NodeValue
				case 0:
					jobstrTime = ""
				}
			}
			jobstrLink = "https://www.jobstreet.com.sg" + jobstrLink

			// Append data get to list
			jobstreetList = append(jobstreetList, &JobStreetDetail{
				Title:       jobstrTitle,
				Company:     jobstrCompany,
				Location:    jobstrLocation,
				Salary:      jobstrSalary,
				DateCreated: jobstrTime,
				Link:        jobstrLink,
			})
		}

		// If less than data limit, get the next page navigation
		if totalJobstr >= 500 {
			break
		} else {
			// Wait the next page division
			if err := chromedp.Run(ctx,
				chromedp.WaitVisible(`div[data-automation="pagination"]`, chromedp.ByQuery),
				chromedp.Nodes(`div[data-automation="pagination"]>a:nth-last-child(1)`, &nodePage, chromedp.ByQuery, chromedp.AtLeast(0)),
			); err != nil {
				jobstrErrMsg = fmt.Sprintf("Failed to Get Next Page Navigation Button, Please Retry!\n%v", err)
				return jobstreetList, jobstrErrMsg, 0
			}

			// If the next page navigation available, click the division to navigate to next page
			if len(nodePage) == 1 {
				if err := chromedp.Run(ctx,
					chromedp.WaitVisible(`div[data-automation="pagination"]`, chromedp.ByQuery),
					chromedp.Click(`div[data-automation="pagination"]>a:nth-last-child(1)`, chromedp.ByQuery),
					chromedp.Sleep(2*time.Second)); err != nil {
					jobstrErrMsg = fmt.Sprintf("Failed to Get Next Page Navigation Button, Please Retry!\n%v", err)
					return jobstreetList, jobstrErrMsg, 0
				}
			} else {
				break
			}
		}
	}

	// Return results if < timeout limit / data limit / page limit
	jobstrErrMsg = fmt.Sprintf("Successfully Scrape Datas From Jobstreet.com......")
	return jobstreetList, jobstrErrMsg, totalJobstr
}
