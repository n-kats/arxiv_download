package main

import (
	"fmt"
	"strconv"
	"strings"
)

const MaxMaxResults = 2000
const Relevance = "relevance"
const LastUpdatedDate = "lastUpdatedDate"
const SubmittedDate = "submittedDate"
const Ascending = "ascending"
const Descending = "descending"
const MaxRetry = 10

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
