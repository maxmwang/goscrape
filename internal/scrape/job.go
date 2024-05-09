package scrape

import (
	"fmt"
	"strings"
	"time"
)

type Job struct {
	company   string
	title     string
	site      site
	updatedAt time.Time
}

func (j Job) String() string {
	if j.updatedAt.IsZero() {
		return fmt.Sprintf("%46s:\t %v", j.company, j.title)
	} else {
		loc, _ := time.LoadLocation("America/Los_Angeles")
		return fmt.Sprintf("%24s: %20s:\t %v", j.updatedAt.In(loc).Format(time.DateTime+" MST"), j.company, j.title)
	}
}

func (j Job) isTarget() bool {
	if strings.Index(j.title, "Intern") == strings.Index(j.title, "Internal") && strings.Count(j.title, "Intern") == 1 {
		return false
	}
	if strings.Index(j.title, "Intern") == strings.Index(j.title, "International") && strings.Count(j.title, "Intern") == 1 {
		return false
	}
	if strings.Contains(j.title, "Software") && strings.Contains(j.title, "Intern") {
		return true
	}

	return false
}
