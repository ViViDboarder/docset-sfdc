package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// JSON Structs

// AtlasTOC represents the meta documenation from Salesforce
type AtlasTOC struct {
	AvailableVersions []VersionInfo `json:"available_versions"`
	Content           string
	ContentDocumentID string `json:"content_document_id"`
	Deliverable       string
	DocTitle          string `json:"doc_title"`
	Locale            string
	Language          LanguageInfo
	PDFUrl            string     `json:"pdf_url"`
	TOCEntries        []TOCEntry `json:"toc"`
	Title             string
	Version           VersionInfo
}

// LanguageInfo contains information for linking and displaying the language
type LanguageInfo struct {
	Label  string
	Locale string
	URL    string
}

// VersionInfo representes a Salesforce documentation version
type VersionInfo struct {
	DocVersion     string `json:"doc_version"`
	ReleaseVersion string `json:"release_version"`
	VersionText    string `json:"version_text"`
	VersionURL     string `json:"version_url"`
}

// TOCEntry represents a single Table of Contents item
type TOCEntry struct {
	Text                    string
	ID                      string
	LinkAttr                LinkAttr `json:"a_attr,omitempty"`
	Children                []TOCEntry
	ComputedFirstTopic      bool
	ComputedResetPageLayout bool
}

// LinkAttr represents all attributes bound to a link
type LinkAttr struct {
	Href string
}

// TOCContent contains content information for a piece of documenation
type TOCContent struct {
	ID      string
	Title   string
	Content string
}

// Sqlite Struct

// SearchIndex is the database table that indexes the docs
type SearchIndex struct {
	ID   int64  `db:id`
	Name string `db:name`
	Type string `db:type`
	Path string `db:path`
}

// IsType indicates that the TOCEntry is of a given SupportedType
// This is done by checking the suffix of the entry text
func (entry TOCEntry) IsType(t SupportedType) bool {
	return strings.HasSuffix(entry.Text, t.TitleSuffix)
}

// CleanTitle trims known suffix from TOCEntry titles
func (entry TOCEntry) CleanTitle(t SupportedType) string {
	if t.NoTrim {
		return entry.Text
	}
	return strings.TrimSuffix(entry.Text, " "+t.TitleSuffix)
}

// GetRelLink extracts only the relative link from the Link Href
func (entry TOCEntry) GetRelLink(removeAnchor bool) (relLink string) {
	if entry.LinkAttr.Href == "" {
		return
	}

	// Get the JSON file
	relLink = entry.LinkAttr.Href
	if removeAnchor {
		anchorIndex := strings.LastIndex(relLink, "#")
		if anchorIndex > 0 {
			relLink = relLink[0:anchorIndex]
		}
	}
	return
}

// GetContent retrieves Content for this TOCEntry from the API
func (entry TOCEntry) GetContent(toc *AtlasTOC) (content *TOCContent, err error) {
	relLink := entry.GetRelLink(true)
	if relLink == "" {
		return
	}

	url := fmt.Sprintf(
		"https://developer.salesforce.com/docs/get_document_content/%s/%s/%s/%s",
		toc.Deliverable,
		relLink,
		toc.Locale,
		toc.Version.DocVersion,
	)

	// fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	// Read the downloaded JSON
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Load into Struct
	content = new(TOCContent)
	err = json.Unmarshal([]byte(contents), content)
	if err != nil {
		fmt.Println("Error reading JSON")
		fmt.Println(resp.Status)
		fmt.Println(url)
		fmt.Println(string(contents))
		return
	}
	return
}

// GetContentFilepath returns the filepath that should be used for the content
func (entry TOCEntry) GetContentFilepath(toc *AtlasTOC, removeAnchor bool) string {
	relLink := entry.GetRelLink(removeAnchor)
	if relLink == "" {
		ExitIfError(NewFormatedError("Link not found for %s", entry.ID))
	}

	return fmt.Sprintf("atlas.%s.%s.meta/%s/%s", toc.Locale, toc.Deliverable, toc.Deliverable, relLink)
}
