package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type siteType string

func siteTypeFromString(str string) siteType {
	switch str {
	case string(siteTypeGreenhouse):
		return siteTypeGreenhouse
	case string(siteTypeAshby):
		return siteTypeAshby
	default:
		return siteTypeCustom
	}
}

const (
	siteTypeGreenhouse siteType = "greenhouse"
	siteTypeAshby      siteType = "ashby"
	siteTypeCustom     siteType = "_custom"
)

type link struct {
	companyName string
	siteType
	url string
}

func openList(filename string) (list []link, err error) {
	if filepath.Ext(filename) != ".csv" {
		return nil, fmt.Errorf("open %s: not a .csv file", filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		list = append(list, link{
			companyName: record[0],
			siteType:    siteTypeFromString(record[1]),
			url:         record[2],
		})
	}

	return list, nil
}

func openRaw(filename string) (list []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		list = append(list, record[0])
	}

	return list, nil
}
