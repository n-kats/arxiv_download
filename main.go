package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

// http://arxiv.org/help/prep
// http://arxiv.org/help/api/user-manual

func main() {
	var cat string
	var w int
	var step int
	flag.IntVar(&w, "wait", 10, "wait n seconds for each download")
	flag.StringVar(&cat, "cat", "", "category math.GT, cs.AI")
	flag.IntVar(&step, "step", 1000, "number of entries for each request")
	flag.Parse()
	if cat == "" {
		fmt.Println("set category")
		return
	}
	q := QueryParams{
		SearchQuery: fmt.Sprintf("cat:%s", cat),
		Start:       0,
		MaxResults:  step,
		SortBy:      SubmittedDate,
		SortOrder:   Ascending,
	}
	wait := time.Duration(w) * time.Second
	fname := fmt.Sprintf("data_%s_%d.json", strings.Replace(cat, ".", "_", 1), time.Now().Unix())
	err := Download(q, fname, wait)
	if err != nil {
		panic(err)
	}
}
