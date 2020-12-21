package main

import (
	"log"
	"net/http"

	"github.com/kokwei0502/gohttp-chromedp-job-search/controllers/globalcontrollers"
	"github.com/kokwei0502/gohttp-chromedp-job-search/controllers/webcontrollers"
)

// var global *globalcontrollers.GlobalUsageStructure

func init() {
	globalcontrollers.GetWorkingDir()
	globalcontrollers.RetrieveAllTemplate()
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", webcontrollers.WebIndexPage)
	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

	// _, b, c := controllers.GetIndeedData("admin assistant")
	// for _, i := range a {
	// 	fmt.Println(i.Company)
	// 	fmt.Println(i.Title)
	// 	fmt.Println(i.Location)
	// 	fmt.Println(i.Salary)
	// 	fmt.Println(i.DateCreated)
	// 	fmt.Println(i.Link)
	// 	fmt.Println("------------------------------")
	// }
	// fmt.Println(b)
	// fmt.Println(c)

	// a, b, _ := controllers.GetJobstreetData("admin assistant")
	// fmt.Println(b)
	// for _, i := range a {
	// 	fmt.Println(i.Title)
	// 	fmt.Println(i.Company)
	// 	fmt.Println(i.Location)
	// 	fmt.Println(i.Salary)
	// 	fmt.Println(i.DateCreated)
	// 	fmt.Println(i.Link)

	// }
}
