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
	var eachSave bool
	var outputDir string
	flag.IntVar(&w, "wait", 10, "wait n seconds for each download")
	flag.StringVar(&cat, "cat", "", "category math.GT, cs.AI")
	flag.IntVar(&step, "step", 1000, "number of entries for each request")
	flag.BoolVar(&eachSave, "each-save", true, "use DownloadWithEachSave")
	flag.StringVar(&outputDir, "outputDir", "data", "output directory")
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
	if eachSave {
		fnameFormat := fmt.Sprintf("%s/data_%s_%%d_%%d.json", outputDir, strings.Replace(cat, ".", "_", 1))
		err := DownloadWithEachSave(q, fnameFormat, wait)
		if err != nil {
			panic(err)
		}
	} else {
		fname := fmt.Sprintf("%s/data_%s_%d.json", outputDir, strings.Replace(cat, ".", "_", 1), time.Now().Unix())
		err := Download(q, fname, wait)
		if err != nil {
			panic(err)
		}
	}
}
