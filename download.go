package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const WaitForNextRequest = 10 * time.Second
const WaitForRetryRequest = 60 * time.Second

func GetXML(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
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

func DownloadOnceWithSave(q *QueryParams, fnameFormat string, wait time.Duration) (int, error) {
	data, err := DownloadOnce(q)
	if err != nil {
		return len(data.Entries), err
	}
	fname := fmt.Sprintf(fnameFormat, q.Start, q.Start+q.MaxResults-1)
	j, err := json.Marshal(data.Entries)
	if err != nil || data.IsEmpty() {
		return len(data.Entries), err
	}
	err = ioutil.WriteFile(fname, j, os.ModePerm)
	return len(data.Entries), err
}

// fnameFormat should be as "data_%d_%d"
func DownloadWithEachSave(q QueryParams, fnameFormat string, wait time.Duration) error {
	for {
		n, err := DownloadOnceWithSave(&q, fnameFormat, wait)
		if n == 0 || err != nil {
			return err
		}
		q.Next()
	}
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
