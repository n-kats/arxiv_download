package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// http://arxiv.org/help/prep
// http://arxiv.org/help/api/user-manual

const MaxMaxResults = 2000
const WaitForNextRequest = 10 * time.Second
const WaitForRetryRequest = 60 * time.Second
const Relevance = "relevance"
const LastUpdatedDate = "lastUpdatedDate"
const SubmittedDate = "submittedDate"
const Ascending = "ascending"
const Descending = "descending"
const MaxRetry = 10

// ArXivInfo carries infomation from arXiv API
type ArXivInfo struct {
	XMLName      xml.Name
	Name         string
	Entries      []Entry `xml:"entry" json:"entry"`
	Updated      string  `xml:"updated" json:"updated"`
	TotalResults int     `xml:"totalResults" json:"totalResults"`
	ItemPerPage  int     `xml:"itemsPerPage"`
	StartIndex   int     `xml:"startIndex"`
}

// Entry is for each arXiv entry/paper
type Entry struct {
	ID              string     `xml:"id" json:"id"`
	Updated         string     `xml:"updated" json:"updated"`
	Published       string     `xml:"published" json:"published"`
	Title           string     `xml:"title" json:"title"`
	Summary         string     `xml:"summary" json:"summary"`
	Authors         []Author   `xml:"author" json:"authors"`
	Comment         string     `xml:"comment" json:"comment"`
	Doi             string     `xml:"doi" json:"doi"`
	JournalRef      string     `xml:"journal_ref" json:"journal_ref"`
	PrimaryCategory Category   `xml:"primary_category" json:"primary_category"`
	Categories      []Category `xml:"category" json:"categories"`
	Links           []Link     `xml:"link" json:"link"`
}

// MSC-class, ACM-class

// Author is
type Author struct {
	Name        string `xml:"name" json:"name"`
	Affiliation string `xml:"affiliation" json:"affiliation"`
}

type Category struct {
	Name string `xml:"term,attr"`
}
type Link struct {
	Title string `xml:"title,attr"`
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
}

func GetXML(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// this returns (Start+1)-MaxResults
type QueryParams struct {
	SearchQuery string
	IDList      []string
	Start       int
	MaxResults  int
	SortBy      string
	SortOrder   string
}

func (q *QueryParams) HasError() error {
	if (q.SearchQuery == "") && (len(q.IDList) == 0) {
		return fmt.Errorf("no request")
	}
	if MaxMaxResults < q.MaxResults {
		return fmt.Errorf("Too large max_results")
	}

	if q.SortBy != "" && q.SortBy != Relevance && q.SortBy != LastUpdatedDate && q.SortBy != SubmittedDate {
		return fmt.Errorf("Illeagal sortBy")
	}
	if q.SortOrder != "" && q.SortOrder != Ascending && q.SortOrder != Descending {
		return fmt.Errorf("Illeagal sortOrder")
	}
	return nil
}

func (q *QueryParams) URL() string {
	s := "http://export.arxiv.org/api/query?"
	needAnd := false
	if q.SearchQuery != "" {
		s += "search_query=" + q.SearchQuery
		needAnd = true
	}

	if len(q.IDList) != 0 {
		if needAnd {
			s += "&"
		}
		s += "id_list=" + strings.Join(q.IDList, ",")
		needAnd = true
	}

	if q.Start != 0 {
		if needAnd {
			s += "&"
		}
		s += "start=" + strconv.Itoa(q.Start)
		needAnd = true
	}
	if q.MaxResults != 0 {
		if needAnd {
			s += "&"
		}
		s += "max_results=" + strconv.Itoa(q.MaxResults)
		needAnd = true
	}

	if q.SortBy != "" {
		if needAnd {
			s += "&"
		}
		s += "sortBy=" + q.SortBy
		needAnd = true
	}

	if q.SortOrder != "" {
		if needAnd {
			s += "&"
		}
		s += "sortOrder=" + q.SortOrder
		needAnd = true
	}

	return s
}
func (q *QueryParams) Next() {
	q.Start += q.MaxResults
}

func (a *ArXivInfo) IsEmpty() bool {
	return len(a.Entries) == 0
}
func ReadXML(data []byte) (ArXivInfo, error) {
	info := new(ArXivInfo)
	err := xml.Unmarshal(data, info)
	return *info, err
}

func DownloadOnce(q *QueryParams) (*ArXivInfo, error) {
	// q -> url
	if err := q.HasError(); err != nil {
		return nil, err
	}
	url := q.URL()
	// url -> data_xml
	fmt.Printf("downloading %s\n", url)
	dataXML, err := GetXML(url)
	if err != nil {
		return nil, err
	}
	// dataXML -> data
	data := new(ArXivInfo)
	err = xml.Unmarshal(dataXML, data)
	if err != nil {
		return nil, err
	}
	return data, err
}

func DownloadAll(q *QueryParams, wait time.Duration) ([]Entry, error) {
	entries := []Entry{}
	retryN := 0
	var err error
	// TODO
	for {
		data, e := DownloadOnce(q)
		err = e
		retry := (data.ItemPerPage == 0) // && data.TotalResults != q.Start+1)
		print(data.ItemPerPage)
		if retry {
			if retryN >= MaxRetry {
				break
			}
			time.Sleep(WaitForRetryRequest)
			retryN++
			continue
		}
		if err != nil || data.IsEmpty() {
			break
		}
		entries = append(entries, data.Entries...)
		q.Next()
		time.Sleep(wait)
	}
	return entries, err
}
func Download(q QueryParams, fname string, wait time.Duration) error {
	data, err := DownloadAll(&q, wait)
	if err != nil {
		return err
	}
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fname, j, os.ModePerm)
	return err
}
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
