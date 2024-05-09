package main

import (
	"strings"
	"time"
)

type job struct {
	company   string
	title     string
	updatedAt time.Time
}

func (j job) condition() bool {
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
