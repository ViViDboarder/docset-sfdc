package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/*
TODO:
 - Move structs to own file
 - Move db stuff to own file
 - Stylesheets
*/

// CSS Paths
var cssBasePath = "https://developer.salesforce.com/resource/stylesheets"
var cssFiles = []string{"holygrail.min.css", "docs.min.css", "syntax-highlighter.min.css"}

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

var dbmap *gorp.DbMap
var wg sync.WaitGroup

const maxConcurrency = 16

var throttle = make(chan int, maxConcurrency)

func parseFlags() (locale string, deliverables []string, silent bool) {
	flag.StringVar(
		&locale, "locale", "en-us",
		"locale to use for documentation (default: en-us)",
	)
	flag.BoolVar(
		&silent, "silent", false, "this flag supresses warning messages",
	)
	flag.Parse()

	// All other args are for deliverables
	// apexcode or pages
	deliverables = flag.Args()
	return
}

// getTOC Retrieves the TOC JSON and Unmarshals it
func getTOC(locale string, deliverable string) (toc *AtlasTOC, err error) {
	var tocURL = fmt.Sprintf("https://developer.salesforce.com/docs/get_document/atlas.%s.%s.meta", locale, deliverable)
	resp, err := http.Get(tocURL)
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
	toc = new(AtlasTOC)
	err = json.Unmarshal([]byte(contents), toc)
	return
}

// verifyVersion ensures that the version retrieved is the latest
func verifyVersion(toc *AtlasTOC) error {
	currentVersion := toc.Version.DocVersion
	topVersion := toc.AvailableVersions[0].DocVersion
	if currentVersion != topVersion {
		return NewFormatedError("verifyVersion : retrieved version is not the latest. Found %s, latest is %s", currentVersion, topVersion)
	}
	return nil
}

func printSuccess(toc *AtlasTOC) {
	fmt.Println("Success:", toc.DocTitle, "-", toc.Version.VersionText)
}

func saveMainContent(toc *AtlasTOC) {
	filePath := fmt.Sprintf("%s.html", toc.Deliverable)
	// Make sure file doesn't exist first
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		content := toc.Content

		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		ExitIfError(err)

		// TODO: Do something to format full page here

		ofile, err := os.Create(filePath)
		ExitIfError(err)

		defer ofile.Close()
		_, err = ofile.WriteString(
			"<meta http-equiv='Content-Type' content='text/html; charset=UTF-8' />" +
				content,
		)
		ExitIfError(err)
	}
}

func main() {
	locale, deliverables, silent := parseFlags()
	if silent {
		WithoutWarning()
	}

	// Download CSS
	for _, cssFile := range cssFiles {
		throttle <- 1
		wg.Add(1)
		go downloadCSS(cssFile, &wg)
	}

	// Init the Sqlite db
	dbmap = initDb()
	err := dbmap.TruncateTables()
	ExitIfError(err)

	for _, deliverable := range deliverables {
		toc, err := getTOC(locale, deliverable)
		ExitIfError(err)

		saveMainContent(toc)

		err = verifyVersion(toc)
		WarnIfError(err)

		// Download each entry
		for _, entry := range toc.TOCEntries {
			if entry.ID == "apex_reference" || entry.ID == "pages_compref" {
				processChildReferences(entry, nil, toc)
			}
		}

		printSuccess(toc)
	}

	wg.Wait()
}

// SupportedType contains information for generating indexes for types we care about
type SupportedType struct {
	TypeName, TitleSuffix                                                     string
	PushName, AppendParents, IsContainer, NoTrim, ShowNamespace, ParseContent bool
}

var supportedTypes = []SupportedType{
	SupportedType{
		TypeName:      "Method",
		TitleSuffix:   "Methods",
		AppendParents: true,
		IsContainer:   true,
		ShowNamespace: true,
	},
	SupportedType{
		TypeName:      "Constructor",
		TitleSuffix:   "Constructors",
		AppendParents: true,
		IsContainer:   true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Class",
		TitleSuffix:   "Class",
		PushName:      true,
		AppendParents: true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Namespace",
		TitleSuffix:   "Namespace",
		PushName:      true,
		AppendParents: true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Interface",
		TitleSuffix:   "Interface",
		PushName:      true,
		AppendParents: true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Statement",
		TitleSuffix:   "Statement",
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Enum",
		TitleSuffix:   "Enum",
		AppendParents: true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Property",
		TitleSuffix:   "Properties",
		AppendParents: true,
		IsContainer:   true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Guide",
		TitleSuffix:   "Example Implementation",
		NoTrim:        true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Statement",
		TitleSuffix:   "Statements",
		NoTrim:        true,
		AppendParents: false,
		IsContainer:   true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Field",
		TitleSuffix:   "Fields",
		AppendParents: true,
		PushName:      true,
		IsContainer:   true,
		ShowNamespace: false,
	},
	SupportedType{
		TypeName:      "Exception",
		TitleSuffix:   "Exceptions",
		NoTrim:        true,
		AppendParents: true,
		ShowNamespace: false,
		ParseContent:  true,
	},
	SupportedType{
		TypeName:      "Constant",
		TitleSuffix:   "Constants",
		NoTrim:        true,
		AppendParents: true,
		ShowNamespace: false,
		ParseContent:  true,
	},
	SupportedType{
		TypeName:      "Class",
		TitleSuffix:   "Class (Base Email Methods)",
		PushName:      true,
		AppendParents: true,
		ShowNamespace: false,
	},
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

func getEntryType(entry TOCEntry) (*SupportedType, error) {
	if strings.HasPrefix(entry.ID, "pages_compref_") {
		return &SupportedType{
			TypeName: "Tag",
			NoTrim:   true,
		}, nil
	}
	for _, t := range supportedTypes {
		if entry.IsType(t) {
			return &t, nil
		}
	}
	return nil, NewTypeNotFoundError(entry)
}

var entryHierarchy []string

func processChildReferences(entry TOCEntry, entryType *SupportedType, toc *AtlasTOC) {
	if entryType != nil && entryType.PushName {
		entryHierarchy = append(entryHierarchy, entry.CleanTitle(*entryType))
	}

	for _, child := range entry.Children {
		// fmt.Println("Processing: " + child.Text)
		var err error
		var childType *SupportedType
		if child.LinkAttr.Href != "" {
			throttle <- 1
			wg.Add(1)

			go downloadContent(child, toc, &wg)

			childType, err = getEntryType(child)
			if childType == nil && (entryType != nil && entryType.IsContainer) {
				saveSearchIndex(dbmap, child, entryType, toc)
			} else if childType != nil && !childType.IsContainer {
				saveSearchIndex(dbmap, child, childType, toc)
			} else {
				WarnIfError(err)
			}
		}
		if len(child.Children) > 0 {
			processChildReferences(child, childType, toc)
		}
	}
	// fmt.Println("Done processing children for " + entry.Text)

	if entryType != nil && entryType.PushName {
		entryHierarchy = entryHierarchy[:len(entryHierarchy)-1]
	}
}

// GetContentFilepath returns the filepath that should be used for the content
func (entry TOCEntry) GetContentFilepath(toc *AtlasTOC, removeAnchor bool) string {
	relLink := entry.GetRelLink(removeAnchor)
	if relLink == "" {
		ExitIfError(NewFormatedError("Link not found for %s", entry.ID))
	}

	return fmt.Sprintf("atlas.%s.%s.meta/%s/%s", toc.Locale, toc.Deliverable, toc.Deliverable, relLink)
}

func downloadContent(entry TOCEntry, toc *AtlasTOC, wg *sync.WaitGroup) {
	defer wg.Done()

	filePath := entry.GetContentFilepath(toc, true)
	// Make sure file doesn't exist first
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		content, err := entry.GetContent(toc)
		ExitIfError(err)

		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		ExitIfError(err)

		// TODO: Do something to format full page here

		ofile, err := os.Create(filePath)
		ExitIfError(err)

		header := "<meta http-equiv='Content-Type' content='text/html; charset=UTF-8' />" +
			"<base href=\"../../\"/>\n"
		for _, cssFile := range cssFiles {
			header += fmt.Sprintf("<link rel=\"stylesheet\" type=\"text/css\" href=\"%s\">", cssFile)
		}
		header += "<style>body { padding: 15px; }</style>"

		defer ofile.Close()
		_, err = ofile.WriteString(
			header + content.Content,
		)
		ExitIfError(err)
	}
	<-throttle
}

func downloadCSS(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(fileName), 0755)
		ExitIfError(err)

		ofile, err := os.Create(fileName)
		ExitIfError(err)
		defer ofile.Close()

		cssURL := cssBasePath + "/" + fileName
		response, err := http.Get(cssURL)
		ExitIfError(err)
		defer response.Body.Close()

		_, err = io.Copy(ofile, response.Body)
		ExitIfError(err)
	}

	<-throttle
}

/**********************
    Database
**********************/

func saveSearchIndex(dbmap *gorp.DbMap, entry TOCEntry, entryType *SupportedType, toc *AtlasTOC) {
	if entry.LinkAttr.Href == "" || entryType == nil {
		return
	}

	relLink := entry.GetContentFilepath(toc, false)
	name := entry.CleanTitle(*entryType)
	if entryType.ShowNamespace && len(entryHierarchy) > 0 {
		// Show namespace for methods
		name = entryHierarchy[len(entryHierarchy)-1] + "." + name
	}

	// fmt.Println("Storing: " + name)

	si := SearchIndex{
		Name: name,
		Type: entryType.TypeName,
		Path: relLink,
	}

	dbmap.Insert(&si)
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "docSet.dsidx")
	ExitIfError(err)

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.AddTableWithName(SearchIndex{}, "searchIndex").SetKeys(true, "ID")

	err = dbmap.CreateTablesIfNotExists()
	ExitIfError(err)

	err = dbmap.TruncateTables()
	ExitIfError(err)

	return dbmap
}
