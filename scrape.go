package main

import (
	"fmt"
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

func main() {
	detailLinks := getDetailLinks()
	schoolData := getSchoolDetailPages(detailLinks)
	schoolDirectoryLinks := getSchoolDirectoryLinks(schoolData)
	fmt.Printf("\n\n\n %+v\n", schoolDirectoryLinks)
}

func getDetailLinks() []string {
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
		fmt.Println("\n\nFinished", r.Request.URL)
	})

	c.Visit("https://www.cde.ca.gov/SchoolDirectory/Results?Title=California%20School%20Directory&search=1&city=San%20Francisco&status=1%2C2&types=66&nps=0&multilingual=0&charter=0&magnet=0&yearround=0&qdc=0&qsc=0&Tab=1&Order=0&Page=0&Items=0&HideCriteria=False&isStaticReport=False")

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
		fmt.Println("\nFinished \n\n", r.Request.URL)
	})

	for _, link := range detailLinks {
		c.Visit(link)
	}

	return SchoolList
}

func getSchoolDirectoryLinks(schoolData []School) []School {
	var updatedSchools = schoolData
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
			// Add to list of possible directories
			updatedSchools[currentSchool].DirectoryLinks = append(updatedSchools[currentSchool].DirectoryLinks, link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("\nFinished \n\n", r.Request.URL)
	})

	for index, school := range schoolData {
		currentSchool = index
		c.Visit(school.Website)
	}

	return updatedSchools
}
