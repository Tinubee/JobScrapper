package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id	   		string
	title 		string
	location 	string
	companyName string
	summary 	string
}

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

func main() {
	var jobs []extractedJob
	totalPages := getPages()
	for i := 0; i < totalPages; i++ {
		extractedjobs := getPage(i)
		jobs = append(jobs, extractedjobs...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs), "jobs")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	file.Write(utf8bom)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link","Title","Location","Company","Summary"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobline := []string{"http://kr.indeed.com/viewjob?jk="+job.id, job.title, job.location, job.companyName, job.summary}
		jwErr := w.Write(jobline)
		checkErr(jwErr)
	}
}

func getPage(page int) []extractedJob{
	//https://kr.indeed.com/jobs?q=python&limit=50&start=50
	var jobs []extractedJob
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res , err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")
	
	searchCards.Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobs = append(jobs, job)
	})

	return jobs
}

func extractJob(card*goquery.Selection) extractedJob {
	id, _:= card.Find(".jcs-JobTitle").Attr("data-jk")
	title := cleanString(card.Find("h2>a>span").Text())
	location := cleanString(card.Find(".companyLocation").Text())
	companyName :=cleanString(card.Find(".companyName").Text())
	summary := cleanString(card.Find(".job-snippet").Text())
	return extractedJob{
		id: id, 
		title: title, 
		location: location, 
		companyName: companyName, 
		summary: summary}
}

func cleanString(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str)) , " ")
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

