package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

type School struct {
	CdsCode        string
	County         string
	Name           string
	Website        string
	DirectoryLinks []string
	CdsLink        string
}

// 19 SF HS: https://www.cde.ca.gov/SchoolDirectory/Results?Title=California%20School%20Directory&search=1&city=San%20Francisco&status=1%2C2&types=66&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&Tab=1&Order=0&Page=0&Items=0&HideCriteria=False&isStaticReport=False
func main() {
	// Every HS in CA
	entryLinks := []string{
		"https://www.cde.ca.gov/SchoolDirectory/Results?title=California%20School%20Directory&search=0&status=1%2C2&types=80%2C66%2C67&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&tab=1&order=0&page=0&items=500&hidecriteria=False&isstaticreport=False",
		"https://www.cde.ca.gov/SchoolDirectory/Results?title=California%20School%20Directory&search=0&status=1%2C2&types=80%2C66%2C67&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&tab=1&order=0&page=1&items=500&hidecriteria=False&isstaticreport=False",
		"https://www.cde.ca.gov/SchoolDirectory/Results?title=California%20School%20Directory&search=0&status=1%2C2&types=80%2C66%2C67&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&tab=1&order=0&page=2&items=500&hidecriteria=False&isstaticreport=False",
		"https://www.cde.ca.gov/SchoolDirectory/Results?title=California%20School%20Directory&search=0&status=1%2C2&types=80%2C66%2C67&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&tab=1&order=0&page=3&items=500&hidecriteria=False&isstaticreport=False",
	}

	detailLinks := getDetailLinks(entryLinks)
	schoolData := getSchoolDetailPages(detailLinks)
	allSchools, schoolsWithDirectories := getSchoolDirectoryLinks(schoolData)

	fmt.Println(schoolsWithDirectories)

	writeResultsToJsonFile(allSchools, "all_results.json")
	writeResultsToJsonFile(schoolsWithDirectories, "directory_results.json")
}

func writeResultsToJsonFile(schoolData []School, fileName string) {
	dataJson, err := json.Marshal(schoolData)
	if err != nil {
		panic(err)
	}
	// We convert the dataJson to a string and then replace escaped & characters
	// json.Marshal auto escapes html characters like <, >, & so we have to replace them
	dataJsonString := strings.Replace(string(dataJson), "\\u0026", "&", -1)
	// Although this is considerably slow, json.Marshal doesnt offer a way to turn off htmlEncoding
	// Read More: https://github.com/golang/go/issues/8592

	ioutil.WriteFile(fileName, []byte(dataJsonString), 0644)
}

func getDetailLinks(linksToVisit []string) []string {
	var detailLinks = make([]string, 0)

	// Instantiate default collector
	c := colly.NewCollector()

	// Get the school's cde.ca.gov detail page
	c.OnHTML("td:nth-child(4) > a", func(e *colly.HTMLElement) {
		link := "https://www.cde.ca.gov" + e.Attr("href")

		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)

		detailLinks = append(detailLinks, link)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// After a request print "Finished ..."
	c.OnScraped(func(r *colly.Response) {
		fmt.Print("Finished ", r.Request.URL, "\n\n")
	})

	for _, link := range linksToVisit {
		c.Visit(link)
	}

	return detailLinks
}

func getSchoolDetailPages(detailLinks []string) []School {
	var SchoolList = make([]School, 0)
	var school = School{}

	// Instantiate default collector
	c := colly.NewCollector()

	// Get school County
	c.OnHTML("table.table.small tr:nth-child(1) > td", func(e *colly.HTMLElement) {
		county := strings.TrimSpace(e.Text)

		// Print link
		fmt.Printf("County found: %q\n", county)
		// Set school county
		school.County = county
	})

	// Get school Name
	c.OnHTML("tr:nth-child(3) > td", func(e *colly.HTMLElement) {
		name := strings.TrimSpace(e.Text)

		fmt.Printf("Name found: %q\n", name)
		// Set school name
		school.Name = name
		school.DirectoryLinks = make([]string, 0)
	})

	// Get school CDS code
	c.OnHTML("tr:nth-child(4) > td", func(e *colly.HTMLElement) {
		cdsCode := strings.TrimSpace(e.Text)

		fmt.Printf("CDS Code found: %q\n", cdsCode)
		// Set school cds code
		school.CdsCode = cdsCode
	})

	// Get school website link
	c.OnHTML("tr:nth-child(10) > td > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		fmt.Printf("Website found: %q -> %s\n", e.Text, link)
		// Set school website
		school.Website = link
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		// Set school cdeLink
		school.CdsLink = r.URL.String()
	})

	// After a request print "Finished ..."
	c.OnScraped(func(r *colly.Response) {
		// Append the new school data
		SchoolList = append(SchoolList, school)
		fmt.Print("Finished ", r.Request.URL, "\n\n")
	})

	for _, link := range detailLinks {
		c.Visit(link)
	}

	return SchoolList
}

func getSchoolDirectoryLinks(schoolData []School) ([]School, []School) {
	var allSchools = schoolData
	var schoolsWithDirectories = make([]School, 0)
	var currentSchool = 0
	c := colly.NewCollector()

	// Get school website link
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// Check if link text matches our expression
		re := regexp.MustCompile("(?i)directory")
		matched := re.MatchString(e.Text)

		if matched {
			// If matched then print link
			fmt.Printf("Possible directory found: %q -> %s\n", e.Text, link)

			// If the link doesnt start with http then prefix it with the base url
			// (directory links should not be partial: ex: /apps/staff -> foo.com/apps/staff)
			if len(link) >= 4 && string(link[0:4]) != "http" {
				// get the base website url
				website := allSchools[currentSchool].Website

				// remove website trailing slash
				if string(website[len(website)-1]) == "/" {
					website = string(website[0 : len(website)-1])
				}
				// remove link leading slash
				if string(link[0]) == "/" {
					link = string(link[1:])
				}

				link = website + "/" + link
			}

			// check if link already exists in dir links
			// (dont allow duplicate links)
			linkAlreadyAdded := false
			for _, dirLink := range allSchools[currentSchool].DirectoryLinks {
				if dirLink == link {
					linkAlreadyAdded = true
				}
			}

			// Add to list of possible directories
			if !linkAlreadyAdded {
				allSchools[currentSchool].DirectoryLinks = append(allSchools[currentSchool].DirectoryLinks, link)
				fmt.Printf("Link Saved\n\n")
			} else {
				fmt.Printf("Link Discarded (duplicate)\n\n")
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnScraped(func(r *colly.Response) {

		// get this school only if possible directory links were found
		if len(allSchools[currentSchool].DirectoryLinks) > 0 {
			schoolsWithDirectories = append(schoolsWithDirectories, allSchools[currentSchool])
		}

		fmt.Print("Finished ", r.Request.URL, "\n\n")
	})

	for index, school := range schoolData {
		currentSchool = index
		c.Visit(school.Website)
	}

	return allSchools, schoolsWithDirectories
}
