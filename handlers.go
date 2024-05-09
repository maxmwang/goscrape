package main

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

func printJobs(c chan job) {
	jobs := make([]job, 0)

	for j := range c {
		jobs = append(jobs, j)
	}

	slices.SortFunc(jobs, func(a, b job) int {
		return strings.Compare(a.updatedAt.String(), b.updatedAt.String())
	})

	for _, j := range jobs {
		if j.condition() {
			if j.updatedAt.IsZero() {
				fmt.Printf("%46s:\t %v\n", j.company, j.title)
			} else {
				loc, _ := time.LoadLocation("America/Los_Angeles")
				fmt.Printf("%24s: %20s:\t %v\n", j.updatedAt.In(loc).Format(time.DateTime+" MST"), j.company, j.title)
			}
		}
	}
}
