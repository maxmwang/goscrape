package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/maxmwang/goscrape/internal/proto"
	"github.com/maxmwang/goscrape/internal/scrape"
	"github.com/maxmwang/goscrape/internal/server"
	"github.com/maxmwang/goscrape/internal/sqlc"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
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
		startDaemon(q, true)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startServer(q)
	}()

	wg.Wait()
}

func startServer(q *sqlc.Queries) {
	s := server.New(q)

	lis, err := net.Listen("tcp", "localhost:5001")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterScraperServer(grpcServer, s)
	grpcServer.Serve(lis)
}

func startDaemon(q *sqlc.Queries, sort bool) {
	r := rate.NewLimiter(rate.Every(30*time.Minute), 1)
	for {
		if err := r.Wait(context.Background()); err != nil {
			panic(err)
		}

		fmt.Printf("[%s] scraping\n", time.Now().Format(time.TimeOnly))
		scrapeOnce(q, sort)
	}
}

func scrapeOnce(q *sqlc.Queries, sort bool) {
	companies, err := q.ListCompanies(context.Background())
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
