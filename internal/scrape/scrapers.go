package scrape

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/maxmwang/goscrape/internal/sqlc"
)

type site string

const (
	siteAshby      site = "ashby"
	siteGreenhouse site = "greenhouse"
	siteLever      site = "lever"
	siteCustom     site = "_custom"
)

func Scrape(company sqlc.Company) (jobs []Job, err error) {
	var f func(string) ([]Job, error)

	switch company.Site {
	case string(siteAshby):
		f = scrapeAshby
	case string(siteGreenhouse):
		f = scrapeGreenhouse
	case string(siteLever):
		f = scrapeLever
	}

	jobs, err = f(company.Name)

	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func scrapeGreenhouse(company string) (jobs []Job, err error) {
	res, err := http.Get(fmt.Sprintf("https://api.greenhouse.io/v1/boards/%s/jobs", company))

	body, err := checkThenDecode[struct {
		Jobs []struct {
			Title     string `json:"title"`
			UpdatedAt string `json:"updated_at"`
		} `json:"jobs"`
	}](company, siteGreenhouse, res, err)
	if err != nil {
		return nil, err
	}

	for _, resJob := range body.Jobs {
		j := Job{
			title:   resJob.Title,
			company: company,
			site:    siteGreenhouse,
		}

		parsedTime, err := time.Parse(time.RFC3339, resJob.UpdatedAt)
		if err == nil {
			j.updatedAt = parsedTime
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func scrapeAshby(company string) (jobs []Job, err error) {
	query := strings.NewReader(fmt.Sprintf(`{"operationName":"ApiJobBoardWithTeams","variables":{"organizationHostedJobsPageName":"%s"},"query":"query ApiJobBoardWithTeams($organizationHostedJobsPageName: String!) {\n  jobBoard: jobBoardWithTeams(\n    organizationHostedJobsPageName: $organizationHostedJobsPageName\n  ) {\n    jobPostings {\n      title\n     }\n  }\n}"}`, company))
	res, err := http.Post("https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobBoardWithTeams", "application/json", query)

	body, err := checkThenDecode[struct {
		Data struct {
			JobBoard struct {
				JobPostings []struct {
					Title string `json:"title"`
				} `json:"jobPostings"`
			} `json:"jobBoard"`
		} `json:"data"`
	}](company, siteAshby, res, err)
	if err != nil {
		return nil, err
	}

	for _, j := range body.Data.JobBoard.JobPostings {
		jobs = append(jobs, Job{
			company: company,
			title:   j.Title,
			site:    siteAshby,
		})
	}

	return jobs, nil
}

func scrapeLever(company string) (jobs []Job, err error) {
	res, err := http.Get(fmt.Sprintf("https://api.lever.co/v0/postings/%s?limit=999", company))

	body, err := checkThenDecode[[]struct {
		Title     string `json:"text"`
		UpdatedAt int    `json:"createdAt"`
	}](company, siteLever, res, err)
	if err != nil {
		return nil, err
	}

	for _, j := range body {
		jobs = append(jobs, Job{
			company:   company,
			title:     j.Title,
			site:      siteLever,
			updatedAt: time.UnixMilli(int64(j.UpdatedAt)),
		})
	}

	return jobs, nil
}

func checkThenDecode[T any](company string, site site, res *http.Response, reqError error) (body T, err error) {
	if reqError != nil {
		return body, fmt.Errorf("failed to request site=%s for company=%s: %w", site, company, err)
	}
	if res.StatusCode != http.StatusOK {
		return body, fmt.Errorf("invalid site=%s for company=%s: code=%d", site, company, res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		return body, fmt.Errorf("failed to decode response from site=%s for company=%s: %w", site, company, err)
	}

	return body, nil
}
