package globalcontrollers

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// Global usage variables
var (
	GlobalTemplate   *template.Template
	GlobalWorkingDir string
)

// RetrieveAllTemplate = Get all .html templates
func RetrieveAllTemplate() *template.Template {
	var htmlListing []string
	err := filepath.Walk(GlobalWorkingDir+"templates/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch checkDir := info.Mode(); {
		case checkDir.IsRegular():
			htmlListing = append(htmlListing, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	GlobalTemplate = template.Must(template.ParseFiles(htmlListing...))
	return GlobalTemplate
}

// GetWorkingDir = Get the current working directory
func GetWorkingDir() string {
	GlobalWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return GlobalWorkingDir
}
