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

// JobsDBDetail = jobsdb.com data structure
type JobsDBDetail struct {
	Title       string
	Company     string
	Location    string
	Salary      string
	DateCreated string
	Link        string
}

const (
	jobsdbBaseURL = "https://sg.jobsdb.com"
)

var (
	jobsDBURL                                                     = "https://sg.jobsdb.com/j?a=7d&l=&q={search-content}&sp=facet_listed_date"
	jobsDBErrMsg                                                  string
	nodeJobsDBMain, nodeJobsDBSalary, nodeJobsDBNext              []*cdp.Node
	nodeJobsDBCompany                                             []*cdp.Node
	jobsdbTitle, jobsdbCompany, jobsdbLocation, jobsdbDateCreated string
	jobsdbLink, jobsdbSalary                                      string
	jobsdbNextPageURL                                             string
	jobsdbNextMap                                                 map[string]string
	boolJobsdbLink                                                bool
	totalJobsDB                                                   = 0
)

// GetJobsDBData = Scrape data from jobsdb.com
func GetJobsDBData(search string) (JobsDBdetail []*JobsDBDetail, Message string, TotalJobsDB int) {
	var jobsdblist []*JobsDBDetail
	searchContent := strings.ReplaceAll(search, " ", "+")
	jobsDBURL = strings.Replace(jobsDBURL, "{search-content}", searchContent, 1)
	fmt.Println(jobsDBURL)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(jobsDBURL),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		jobsDBErrMsg = fmt.Sprintf("Navigation Error, Please Try Again Later!\n%v", err)
		return jobsdblist, jobsDBErrMsg, 0
	}
	for {
		if err := chromedp.Run(ctx,
			chromedp.Nodes(`div[class="job-container result organic-job"]`, &nodeJobsDBMain, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			jobsDBErrMsg = fmt.Sprintf("Error to Get Main Division, Please Try Again Later\n%v", err)
			return jobsdblist, jobsDBErrMsg, 0
		}
		fmt.Println(len(nodeJobsDBMain))
		totalJobsDB += len(nodeJobsDBMain)
		if len(nodeJobsDBMain) > 0 {
			for i := 0; i < (len(nodeJobsDBMain)); i++ {
				if err := chromedp.Run(ctx,
					chromedp.Text(`div.job-item-top-container > h3`, &jobsdbTitle, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					chromedp.Nodes(`span.job-company`, &nodeJobsDBCompany, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					chromedp.Text(`span.job-location`, &jobsdbLocation, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					chromedp.Text(`span.job-listed-date`, &jobsdbDateCreated, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					chromedp.AttributeValue(`a.job-item`, "href", &jobsdbLink, &boolJobsdbLink, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					chromedp.Nodes(`div.job-salary-badge`, &nodeJobsDBSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
				); err != nil {
					jobsDBErrMsg = fmt.Sprintf("Error to Get Datas From JobsDB, Please Try Again Later\n%v", err)
					return jobsdblist, jobsDBErrMsg, totalJobsDB
				}
				if len(nodeJobsDBSalary) == 1 {
					if err := chromedp.Run(ctx,
						chromedp.Text(`div.job-salary-badge`, &jobsdbSalary, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					); err != nil {
						jobsDBErrMsg = fmt.Sprintf("Error to Get Salary Division From JobsDB, Please Try Again Later\n%v", err)
						return jobsdblist, jobsDBErrMsg, totalJobsDB
					}
				} else {
					jobsdbSalary = "No Stated"
				}
				if len(nodeJobsDBCompany) == 1 {
					if err := chromedp.Run(ctx,
						chromedp.Text(`span.job-company`, &jobsdbCompany, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(nodeJobsDBMain[i])),
					); err != nil {
						jobsDBErrMsg = fmt.Sprintf("Error to Get Company Division From JobsDB, Please Try Again Later\n%v", err)
						return jobsdblist, jobsDBErrMsg, totalJobsDB
					}
				} else {
					jobsdbCompany = "No Stated"
				}
				jobsdbLink = jobsdbBaseURL + jobsdbLink
				jobsdblist = append(jobsdblist, &JobsDBDetail{
					Title:       jobsdbTitle,
					Company:     jobsdbCompany,
					Location:    jobsdbLocation,
					Salary:      jobsdbSalary,
					DateCreated: jobsdbDateCreated,
					Link:        jobsdbLink,
				})
			}
			// Next page navigation
			// multi-pages-pagination pagination-container
			if err := chromedp.Run(ctx,
				chromedp.Nodes(`a.next-page-button`, &nodeJobsDBNext, chromedp.ByQuery, chromedp.AtLeast(0)),
			); err != nil {
				jobsDBErrMsg = fmt.Sprintf("Error to Get Next Page Division From JobsDB, Please Try Again Later\n%v", err)
				return jobsdblist, jobsDBErrMsg, totalJobsDB
			}
			fmt.Println(len(nodeJobsDBNext))
			if totalJobsDB < 200 {
				if len(nodeJobsDBNext) == 1 {
					if err := chromedp.Run(ctx,
						chromedp.Attributes(`a.next-page-button`, &jobsdbNextMap, chromedp.ByQuery),
					); err != nil {
						jobsDBErrMsg = fmt.Sprintf("Error to Get Next Page Navigate From JobsDB, Please Try Again Later\n%v", err)
						return jobsdblist, jobsDBErrMsg, totalJobsDB
					}
					jobsdbNextPageURL = jobsdbBaseURL + jobsdbNextMap["href"]
					fmt.Println(jobsdbNextPageURL)
					if err := chromedp.Run(ctx,
						chromedp.Navigate(jobsdbNextPageURL),
						chromedp.Sleep(2*time.Second),
					); err != nil {
						jobsDBErrMsg = fmt.Sprintf("Failed to Get Next Page Navigation Button, Please Retry!%v", err)
						return jobsdblist, jobsDBErrMsg, totalJobsDB
					}
				}
			} else {
				break
			}
		} else {
			break
		}
	}
	jobsDBErrMsg = fmt.Sprintf("Successfully Scrape Data From JobsDB.com")
	return jobsdblist, jobsDBErrMsg, totalJobsDB
}
