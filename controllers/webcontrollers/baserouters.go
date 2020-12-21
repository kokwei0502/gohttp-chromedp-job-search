package webcontrollers

import (
	"net/http"

	"github.com/kokwei0502/gohttp-chromedp-job-search/controllers/controllers"
	"github.com/kokwei0502/gohttp-chromedp-job-search/controllers/globalcontrollers"
)

// PageData = index page data
type PageData struct {
	JobStreetSearchResults []*controllers.JobStreetDetail
	ErrMessage             string
	TotalResultsFound      int
}

var (
	inputSearch     string
	errMsg          string
	totalResult     int
	jobstreetSearch []*controllers.JobStreetDetail
)

// WebIndexPage = index page
func WebIndexPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		searchPlatform := r.FormValue("search-platform")
		inputSearch = r.FormValue("search-keywords")
		if inputSearch == "" {
			errMsg = "The Search Bar Can't Be Empty!"
		} else if inputSearch != "" {
			switch searchPlatform {
			case "indeed":
				jobstreetSearch = nil
			case "jobstreet":
				jobstreetSearch, errMsg, totalResult = controllers.GetJobstreetData(inputSearch)
			}
		}

	}
	pageData := &PageData{
		JobStreetSearchResults: jobstreetSearch,
		ErrMessage:             errMsg,
		TotalResultsFound:      totalResult,
	}

	globalcontrollers.GlobalTemplate.ExecuteTemplate(w, "index.html", &pageData)
}

// func serveTemplate(w http.ResponseWriter, r *http.Request) {
// 	lp := filepath.Join("templates", "layout.html")
// 	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

// 	// Return a 404 if the template doesn't exist
// 	info, err := os.Stat(fp)
// 	if err != nil {
// 	  if os.IsNotExist(err) {
// 		http.NotFound(w, r)
// 		return
// 	  }
// 	}

// 	// Return a 404 if the request is for a directory
// 	if info.IsDir() {
// 	  http.NotFound(w, r)
// 	  return
// 	}

// 	tmpl, err := template.ParseFiles(lp, fp)
// 	if err != nil {
// 	  // Log the detailed error
// 	  log.Println(err.Error())
// 	  // Return a generic "Internal Server Error" message
// 	  http.Error(w, http.StatusText(500), 500)
// 	  return
// 	}

// 	err = tmpl.ExecuteTemplate(w, "layout", nil)
// 	if err != nil {
// 	  log.Println(err.Error())
// 	  http.Error(w, http.StatusText(500), 500)
// 	}
//   }
