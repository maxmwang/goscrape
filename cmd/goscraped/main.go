package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/maxmwang/goscrape/internal/scrape"
	"github.com/maxmwang/goscrape/internal/sqlc"
)

func main() {
	db, err := sqlc.Connect(false)
	if err != nil {
		panic(err)
	}
	q := sqlc.New(db)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scrapeDaemon(q, true)
	}()

	wg.Wait()
}

func scrapeDaemon(q *sqlc.Queries, sort bool) {
	companies, err := q.GetCompanies(context.Background())
	if err != nil {
		panic(err)
	}

	c := make(chan scrape.Job)

	wgPrint := sync.WaitGroup{}
	wgPrint.Add(1)

	go func() {
		defer wgPrint.Done()

		if sort {
			scrape.SortThenPrint(c)
		} else {
			scrape.Print(c)
		}
	}()

	wgScrape := sync.WaitGroup{}

	for _, company := range companies {
		wgScrape.Add(1)

		go func(company sqlc.Company) {
			defer wgScrape.Done()

			jobs, scrapeErr := scrape.Scrape(company)
			if scrapeErr != nil {
				fmt.Println(scrapeErr)
			}
			for _, j := range jobs {
				c <- j
			}
		}(company)
	}

	wgScrape.Wait()
	close(c)
	wgPrint.Wait()
}

// fromList migrates a .csv list to the sqlite table.
func fromList(filename string, q *sqlc.Queries) {
	if filepath.Ext(filename) != ".csv" {
		panic(fmt.Errorf("open %s: not a .csv file", filename))
	}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	list := make([]sqlc.AddCompanyParams, 0)
	r := csv.NewReader(f)
	r.ReuseRecord = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		list = append(list, sqlc.AddCompanyParams{
			Name: record[0],
			Site: record[1],
		})
	}

	for _, comp := range list {
		err := q.AddCompany(context.Background(), comp)
		if err != nil {
			fmt.Printf("failed to add company=%s site=%s: %s\n", comp.Name, comp.Site, err.Error())
		}
	}
}
