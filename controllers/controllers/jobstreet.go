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

type JobStreetDetail struct {
	Title       string
	Company     string
	Location    string
	Salary      string
	DateCreated string
	Link        string
}

func GetJobstreetData(search string) (JobStreetResult []*JobStreetDetail, Message string, TotalJobStreet int) {
	var jobstreetList []*JobStreetDetail
	searchContent := strings.ReplaceAll(search, " ", "-")
	jobStreetURL = strings.Replace(jobStreetURL, "{search-content}", searchContent, 1)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(jobStreetURL),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		jobstrErrMsg = "Navigation Error, Please Try Again Later!"
		return jobstreetList, jobstrErrMsg, 0
		// return dataListing, errMsg, 0
	}
	var ok bool
	for {
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`article`),
			chromedp.Nodes(`article`, &nodeArticles, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			jobstrErrMsg = "Error to Get Main Division, Please Try Again Later"
			return jobstreetList, jobstrErrMsg, 0
		}
		totalJobstr += len(nodeArticles)
		for i := 0; i < len(nodeArticles); i++ {
			if err := chromedp.Run(ctx,
				chromedp.Text(`h1`, &jobstrTitle, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
				chromedp.Nodes(`span`, &nodeSpan, chromedp.ByQueryAll, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
				chromedp.AttributeValue("h1 > a", "href", &jobstrLink, &ok, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeArticles[i])),
			); err != nil {
				jobstrErrMsg = "Error to Get Informations, Please Try Again Later"
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
			jobstreetList = append(jobstreetList, &JobStreetDetail{
				Title:       jobstrTitle,
				Company:     jobstrCompany,
				Location:    jobstrLocation,
				Salary:      jobstrSalary,
				DateCreated: jobstrTime,
				Link:        jobstrLink,
			})
		}
		fmt.Println(totalJobstr)
		if totalJobstr >= 500 {
			break
		} else {
			if err := chromedp.Run(ctx,
				chromedp.WaitVisible(`div[data-automation="pagination"]`, chromedp.ByQuery),
				chromedp.Nodes(`div[data-automation="pagination"]>a:nth-last-child(1)`, &nodePage, chromedp.ByQuery, chromedp.AtLeast(0)),
			); err != nil {
				jobstrErrMsg = "Failed to Get Next Page Navigation Button, Please Retry!"
				return jobstreetList, jobstrErrMsg, 0
			}

			if len(nodePage) == 1 {
				if err := chromedp.Run(ctx,
					chromedp.WaitVisible(`div[data-automation="pagination"]`, chromedp.ByQuery),
					chromedp.Click(`div[data-automation="pagination"]>a:nth-last-child(1)`, chromedp.ByQuery),
					chromedp.Sleep(2*time.Second)); err != nil {
					jobstrErrMsg = "Failed to Get Next Page Navigation Button, Please Retry!"
					return jobstreetList, jobstrErrMsg, 0
				}
			} else {
				break
			}
		}
	}
	jobstrErrMsg = "Successfully Scrape Datas From Jobstreet.com......"
	return jobstreetList, jobstrErrMsg, totalJobstr
}
