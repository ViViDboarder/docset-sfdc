package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// CSS Paths
var cssBaseURL = "https://developer.salesforce.com/resource/stylesheets"
var cssFiles = []string{"holygrail.min.css", "docs.min.css", "syntax-highlighter.min.css"}
var buildDir = "build"

var wg sync.WaitGroup
var throttle = make(chan int, maxConcurrency)

const maxConcurrency = 16

func parseFlags() (locale string, deliverables []string, debug bool) {
	flag.StringVar(
		&locale, "locale", "en-us",
		"locale to use for documentation (default: en-us)",
	)
	flag.BoolVar(
		&debug, "debug", false, "this flag supresses warning messages",
	)
	flag.Parse()

	// All other args are for deliverables
	// apexcode, pages, or lightening
	deliverables = flag.Args()
	return
}

// getTOC Retrieves the TOC JSON and Unmarshals it
func getTOC(locale string, deliverable string) (toc *AtlasTOC, err error) {
	var tocURL = fmt.Sprintf("https://developer.salesforce.com/docs/get_document/atlas.%s.%s.meta", locale, deliverable)
	LogDebug("TOC URL: %s", tocURL)
	resp, err := http.Get(tocURL)
	ExitIfError(err)

	// Read the downloaded JSON
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	ExitIfError(err)

	// Load into Struct
	toc = new(AtlasTOC)
	LogDebug("TOC JSON: %s", string(contents))
	err = json.Unmarshal([]byte(contents), toc)
	return
}

// verifyVersion ensures that the version retrieved is the latest
func verifyVersion(toc *AtlasTOC) error {
	// jsonVersion, _ := json.Marshal(toc.Version)
	// LogDebug("toc.Version" + string(jsonVersion))
	currentVersion := toc.Version.DocVersion
	// jsonAvailVersions, _ := json.Marshal(toc.AvailableVersions)
	// LogDebug("toc.AvailableVersions" + string(jsonAvailVersions))
	topVersion := toc.AvailableVersions[0].DocVersion
	if currentVersion != topVersion {
		return NewFormatedError("verifyVersion: retrieved version is not the latest. Found %s, latest is %s", currentVersion, topVersion)
	}
	return nil
}

func printSuccess(toc *AtlasTOC) {
	LogInfo("Success: %s - %s - %s", toc.DocTitle, toc.Version.VersionText, toc.Version.DocVersion)
}

func saveMainContent(toc *AtlasTOC) {
	filePath := fmt.Sprintf("%s.html", toc.Deliverable)
	// Prepend build dir
	filePath = filepath.Join(buildDir, filePath)
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

func saveContentVersion(toc *AtlasTOC) {
	filePath := fmt.Sprintf("%s-version.txt", toc.Deliverable)
	// Prepend build dir
	filePath = filepath.Join(buildDir, filePath)
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	ExitIfError(err)

	ofile, err := os.Create(filePath)
	ExitIfError(err)

	defer ofile.Close()
	_, err = ofile.WriteString(toc.Version.DocVersion)
	ExitIfError(err)
}

func downloadCSS(fileName string, wg *sync.WaitGroup) {
	downloadFile(cssBaseURL+"/"+fileName, fileName, wg)
}

func downloadFile(url string, fileName string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	filePath := filepath.Join(buildDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		ExitIfError(err)

		ofile, err := os.Create(filePath)
		ExitIfError(err)
		defer ofile.Close()

		response, err := http.Get(url)
		ExitIfError(err)
		defer response.Body.Close()

		_, err = io.Copy(ofile, response.Body)
		ExitIfError(err)
	}

	if wg != nil {
		<-throttle
	}
}

func getEntryType(entry TOCEntry) (*SupportedType, error) {
	for _, t := range SupportedTypes {
		if entry.IsType(t) {
			return &t, nil
		}
	}
	return nil, NewTypeNotFoundError(entry)
}

// processEntryReference downloads html and indexes a toc item
func processEntryReference(entry TOCEntry, entryType *SupportedType, toc *AtlasTOC) {
	LogDebug("Processing: %s", entry.Text)
	throttle <- 1
	wg.Add(1)

	go downloadContent(entry, toc, &wg)

	if entryType == nil {
		LogDebug("No entry type for %s. Cannot index", entry.Text)
	} else if entryType.IsContainer || entryType.IsHidden {
		LogDebug("%s is a container or is hidden. Do not index", entry.Text)
	} else {
		SaveSearchIndex(dbmap, entry, entryType, toc)
	}
}

// entryHierarchy allows breadcrumb naming
var entryHierarchy []string

// processChildReferences iterates through all child toc items, cascading types, and indexes them
func processChildReferences(entry TOCEntry, entryType *SupportedType, toc *AtlasTOC) {
	if entryType != nil && entryType.PushName {
		entryHierarchy = append(entryHierarchy, entry.CleanTitle(*entryType))
	}

	for _, child := range entry.Children {
		LogDebug("Reading child: %s", child.Text)
		var err error
		var childType *SupportedType
		// Skip anything without an HTML page
		if child.LinkAttr.Href != "" {
			childType, err = getEntryType(child)
			if childType == nil && entryType != nil && (entryType.IsContainer || entryType.CascadeType) {
				// No child type, and parent is set to cascade
				LogDebug("Parent was container or cascade, using parent type of %s", entryType.TypeName)
				childType = entryType
				childType.IsContainer = false
			} else if childType != nil && entryType != nil {
				// We didn't cascade in full, but some features are still hereditary
				if entryType.IsHidden {
					childType.IsHidden = true
				}
			}
			if childType == nil && err != nil {
				WarnIfError(err)
			}
			processEntryReference(child, childType, toc)
		} else {
			LogDebug("%s has no link. Skipping", child.Text)
		}
		if len(child.Children) > 0 {
			processChildReferences(child, childType, toc)
		}
	}
	LogDebug("Done processing children for %s", entry.Text)

	if entryType != nil && entryType.PushName {
		entryHierarchy = entryHierarchy[:len(entryHierarchy)-1]
	}
}

// downloadContent will download the html file for a given entry
func downloadContent(entry TOCEntry, toc *AtlasTOC, wg *sync.WaitGroup) {
	defer wg.Done()

	filePath := entry.GetContentFilepath(toc, true)
	// Prepend build dir
	filePath = filepath.Join(buildDir, filePath)
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

func main() {
	LogInfo("Starting...")
	locale, deliverables, debug := parseFlags()
	if debug {
		SetLogLevel(DEBUG)
	}

	// Download CSS
	for _, cssFile := range cssFiles {
		throttle <- 1
		wg.Add(1)
		go downloadCSS(cssFile, &wg)
	}

	// Download icon
	go downloadFile("https://developer.salesforce.com/resources2/favicon.ico", "icon.ico", nil)

	// Init the Sqlite db
	dbmap = InitDb(buildDir)
	err := dbmap.TruncateTables()
	ExitIfError(err)

	for _, deliverable := range deliverables {
		toc, err := getTOC(locale, deliverable)

		err = verifyVersion(toc)
		WarnIfError(err)

		saveMainContent(toc)
		saveContentVersion(toc)

		// Download each entry
		for _, entry := range toc.TOCEntries {
			entryType, err := getEntryType(entry)
			if entryType != nil && err == nil {
				processEntryReference(entry, entryType, toc)
			}
			processChildReferences(entry, entryType, toc)
		}

		printSuccess(toc)
	}

	wg.Wait()
}
