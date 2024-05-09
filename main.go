package main

import (
	"fmt"
	"strings"
	"sync"
)

func scrape() {
	links, err := openList("list.csv")
	if err != nil {
		panic(err)
	}

	c := make(chan job)

	wg0 := sync.WaitGroup{}
	wg0.Add(1)
	go func() {
		printJobs(c)
		wg0.Done()
	}()

	wg1 := sync.WaitGroup{}

	for _, l := range links {
		wg1.Add(1)
		go func(l link) {
			defer wg1.Done()

			switch l.siteType {
			case siteTypeGreenhouse:
				jobs, err := scrapeGreenhouse(l.companyName)
				if err != nil {
					fmt.Println(err)
				}

				for _, j := range jobs {
					c <- j
				}
			case siteTypeAshby:
				jobs, err := scrapeAshby(l.companyName)
				if err != nil {
					fmt.Println(err)
				}

				for _, j := range jobs {
					c <- j
				}
			case siteTypeLever:
				jobs, err := scrapeLever(l.companyName)
				if err != nil {
					fmt.Println(err)
				}

				for _, j := range jobs {
					c <- j
				}
			default:
			}

			c <- job{}
		}(l)
	}

	wg1.Wait()
	close(c)
	wg0.Wait()
}

func populate() {
	links, err := openList("list.csv")
	if err != nil {
		panic(err)
	}
	set := make(map[string]struct{})
	for _, l := range links {
		set[l.companyName] = struct{}{}
	}

	names := openSummer()

	c := 0
	i := 0
	n := 0
	l := sync.Mutex{}

	wg := sync.WaitGroup{}
	for _, name := range names {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			var jobs []job
			var err error

			name = strings.ToLower(name)
			name = strings.Split(name, "-")[0]

			jobs, err = scrapeGreenhouse(name)
			if err == nil && len(jobs) > 0 {
				if _, ok := set[name]; !ok {
					set[name] = struct{}{}
					fmt.Printf("%s,greenhouse\n", name)

					l.Lock()
					c++
					l.Unlock()
					return
				} else {
					l.Lock()
					i++
					l.Unlock()
					return
				}
			}

			jobs, err = scrapeAshby(name)
			if err == nil && len(jobs) > 0 {
				if _, ok := set[name]; !ok {
					set[name] = struct{}{}
					fmt.Printf("%s,ashby\n", name)

					l.Lock()
					c++
					l.Unlock()
					return
				} else {
					l.Lock()
					i++
					l.Unlock()
					return
				}
			}

			jobs, err = scrapeLever(name)
			if err == nil && len(jobs) > 0 {
				if _, ok := set[name]; !ok {
					set[name] = struct{}{}
					fmt.Printf("%s,lever\n", name)

					l.Lock()
					c++
					l.Unlock()
					return
				} else {
					l.Lock()
					i++
					l.Unlock()
					return
				}
			}

			l.Lock()
			n++
			l.Unlock()
		}(name)
	}

	wg.Wait()
	fmt.Println("")
	fmt.Printf("new: %d/%d\n", c, len(names))
	fmt.Printf("old: %d/%d\n", i, len(names))
	fmt.Printf("no:  %d/%d\n", n, len(names))
}

func main() {
	scrape()
}

/*
in:  110/159
out: 49/159
coverage: 110/254
*/
