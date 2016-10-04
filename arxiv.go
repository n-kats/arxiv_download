package main

import (
	"encoding/xml"
)

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

// IsEmpty checks the information is empty
func (a *ArXivInfo) IsEmpty() bool {
	return len(a.Entries) == 0
}

// Author is
type Author struct {
	Name        string `xml:"name" json:"name"`
	Affiliation string `xml:"affiliation" json:"affiliation"`
}

// Category is category in arXiv
// MSC-class, ACM-class
type Category struct {
	Name string `xml:"term,attr"`
}

// Link to pdf, abstract, ...
type Link struct {
	Title string `xml:"title,attr"`
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
}
