package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

func newCollector(goquerySelector string, f func(e *colly.HTMLElement) string) *colly.Collector {
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("n", fmt.Sprintf("%d", 0))
	})

	c.OnHTML(goquerySelector, func(e *colly.HTMLElement) {
		jobTitle := f(e)
		if len(jobTitle) > 5 {
			i, _ := strconv.Atoi(e.Request.Ctx.Get("n"))
			e.Request.Ctx.Put("n", fmt.Sprintf("%d", i+1))
		}
		if len(jobTitle) > 100 && !strings.Contains(jobTitle, "Job") {
			return
		}
		if strings.Contains(jobTitle, "Software") && strings.Contains(jobTitle, "Intern") {
			fmt.Printf("%20s:\t %v\n", e.Request.Ctx.Get("name"), jobTitle)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		n, _ := strconv.Atoi(r.Ctx.Get("n"))
		if n < 5 {
			fmt.Printf("%20s: only has %d postings\n", r.Ctx.Get("name"), n)
		}
	})

	return c
}

func cockroachlabs() {
	res, err := http.Get("https://api.greenhouse.io/v1/boards/cockroachlabs/departments/")
	if err != nil {
		fmt.Printf("could not query cockroachlabs listings: %s", err.Error())
		return
	}

	var body struct {
		Departments []struct {
			Name string `json:"name"`
			Jobs []struct {
				Title string `json:"title"`
			} `json:"jobs"`
		} `json:"departments"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		fmt.Printf("could not decode request from cockroachlabs listing: %s", err.Error())
		return
	}

	for _, department := range body.Departments {
		if department.Name == "Engineering" {
			for _, job := range department.Jobs {
				if strings.Contains(job.Title, "Intern") {
					fmt.Printf("%20s:\t %v\n", "cockroach labs", job.Title)
				}
			}
		}
	}
}

func main() {
	links, err := openList("list.csv")
	if err != nil {
		panic(err)
	}

	greenhouseCollector := newCollector("div.opening", func(e *colly.HTMLElement) string {
		return e.ChildText("a")
	})
	generalCollector := newCollector("a[href]", func(e *colly.HTMLElement) string {
		return e.Text
	})

	wg := sync.WaitGroup{}
	for _, l := range links {
		wg.Add(1)
		go func(l link) {
			defer wg.Done()

			ctx := colly.NewContext()
			ctx.Put("name", l.companyName)
			ctx.Put("type", l.siteType)

			switch l.siteType {
			case siteTypeGreenhouse:
				greenhouseCollector.Request(http.MethodGet, l.url, nil, ctx, nil)
			// case siteTypeAshby:
			// 	ashbyCollector.Request(http.MethodGet, l.url, nil, ctx, nil)
			default:
				if l.companyName == "cockroach labs" {
					cockroachlabs()
				} else {
					generalCollector.Request(http.MethodGet, l.url, nil, ctx, nil)
				}
			}
		}(l)
	}

	wg.Wait()
}

// use https://api.greenhouse.io/v1/boards/[company name]/jobs
