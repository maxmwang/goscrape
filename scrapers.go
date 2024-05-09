package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func requestError(siteType siteType, company string, err error) error {
	return fmt.Errorf("could not decode request from %s listing for %s: %s", siteType, company, err.Error())
}

func requestNotOKError(siteType siteType, company string, status int) error {
	return fmt.Errorf("could not get %s listing for %s: code=%d", siteType, company, status)
}

func parseResponse[T any](siteType siteType, company string, res *http.Response) (body T, err error) {
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		return body, fmt.Errorf("could not decode request from %s listing for %s: %s", siteType, company, err.Error())
	}
	return body, nil
}

func scrapeGreenhouse(company string) (jobs []job, err error) {
	res, err := http.Get(fmt.Sprintf("https://api.greenhouse.io/v1/boards/%s/jobs", company))
	if err != nil {
		return nil, requestError(siteTypeGreenhouse, company, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, requestNotOKError(siteTypeGreenhouse, company, res.StatusCode)
	}

	body, err := parseResponse[struct {
		Jobs []struct {
			Title     string `json:"title"`
			UpdatedAt string `json:"updated_at"`
		} `json:"jobs"`
	}](siteTypeGreenhouse, company, res)
	if err != nil {
		return nil, err
	}

	for _, j := range body.Jobs {
		t, err := time.Parse(time.RFC3339, j.UpdatedAt)
		if err != nil {
			jobs = append(jobs, job{
				title: j.Title,
			})
		} else {
			jobs = append(jobs, job{
				company:   company,
				title:     j.Title,
				updatedAt: t,
			})
		}
	}

	return jobs, nil
}

func scrapeAshby(company string) (jobs []job, err error) {
	query := strings.NewReader(fmt.Sprintf(`{"operationName":"ApiJobBoardWithTeams","variables":{"organizationHostedJobsPageName":"%s"},"query":"query ApiJobBoardWithTeams($organizationHostedJobsPageName: String!) {\n  jobBoard: jobBoardWithTeams(\n    organizationHostedJobsPageName: $organizationHostedJobsPageName\n  ) {\n    jobPostings {\n      title\n     }\n  }\n}"}`, company))
	res, err := http.Post("https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobBoardWithTeams", "application/json", query)
	if err != nil {
		return nil, requestError(siteTypeAshby, company, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, requestNotOKError(siteTypeAshby, company, res.StatusCode)
	}

	body, err := parseResponse[struct {
		Data struct {
			JobBoard struct {
				JobPostings []struct {
					Title string `json:"title"`
				} `json:"jobPostings"`
			} `json:"jobBoard"`
		} `json:"data"`
	}](siteTypeAshby, company, res)
	if err != nil {
		return nil, err
	}

	for _, j := range body.Data.JobBoard.JobPostings {
		jobs = append(jobs, job{
			company: company,
			title:   j.Title,
		})
	}

	return jobs, nil
}

func scrapeLever(company string) (jobs []job, err error) {
	res, err := http.Get(fmt.Sprintf("https://api.lever.co/v0/postings/%s?limit=999", company))
	if err != nil {
		return nil, requestError(siteTypeLever, company, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, requestNotOKError(siteTypeLever, company, res.StatusCode)
	}

	body, err := parseResponse[[]struct {
		Title     string `json:"text"`
		UpdatedAt int    `json:"createdAt"`
	}](siteTypeLever, company, res)
	if err != nil {
		return nil, err
	}

	for _, j := range body {
		jobs = append(jobs, job{
			company:   company,
			title:     j.Title,
			updatedAt: time.UnixMilli(int64(j.UpdatedAt)),
		})
	}

	return jobs, nil
}
