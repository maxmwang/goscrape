package scrape

import (
	"fmt"
	"slices"
	"strings"
)

func SortThenPrint(c <-chan Job) {
	jobs := make([]Job, 0)

	for j := range c {
		jobs = append(jobs, j)
	}

	slices.SortFunc(jobs, func(a, b Job) int {
		return strings.Compare(a.updatedAt.String(), b.updatedAt.String())
	})

	for _, j := range jobs {
		if j.isTarget() {
			fmt.Println(j)
		}
	}
}

func Print(c <-chan Job) {
	for j := range c {
		if j.isTarget() {
			fmt.Println(j)
		}
	}
}
