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
var cssBasePath = "https://developer.salesforce.com/resource/stylesheets"
var cssFiles = []string{"holygrail.min.css", "docs.min.css", "syntax-highlighter.min.css"}

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
	// apexcode or pages
	deliverables = flag.Args()
	return
}

// getTOC Retrieves the TOC JSON and Unmarshals it
func getTOC(locale string, deliverable string) (toc *AtlasTOC, err error) {
	var tocURL = fmt.Sprintf("https://developer.salesforce.com/docs/get_document/atlas.%s.%s.meta", locale, deliverable)
	resp, err := http.Get(tocURL)
	ExitIfError(err)

	// Read the downloaded JSON
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	ExitIfError(err)

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
		return NewFormatedError("verifyVersion: retrieved version is not the latest. Found %s, latest is %s", currentVersion, topVersion)
	}
	return nil
}

func printSuccess(toc *AtlasTOC) {
	LogInfo("Success: %s - %s - %s", toc.DocTitle, toc.Version.VersionText, toc.Version.DocVersion)
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

func saveContentVersion(toc *AtlasTOC) {
	filePath := fmt.Sprintf("%s-version.txt", toc.Deliverable)
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	ExitIfError(err)

	ofile, err := os.Create(filePath)
	ExitIfError(err)

	defer ofile.Close()
	_, err = ofile.WriteString(toc.Version.DocVersion)
	ExitIfError(err)
}

func main() {
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

	// Init the Sqlite db
	dbmap = InitDb()
	err := dbmap.TruncateTables()
	ExitIfError(err)

	for _, deliverable := range deliverables {
		toc, err := getTOC(locale, deliverable)

		err = verifyVersion(toc)
		WarnIfError(err)

		saveMainContent(toc)
		saveContentVersion(toc)

		// Download each entry
		/*
		 * topLevelEntryIDs := map[string]bool{
		 * 	"apex_dev_guide": true,
		 * 	"pages_compref":  true,
		 * }
		 */
		for _, entry := range toc.TOCEntries {
			processChildReferences(entry, nil, toc)
		}

		printSuccess(toc)
	}

	wg.Wait()
}

func getEntryType(entry TOCEntry) (*SupportedType, error) {
	for _, t := range SupportedTypes {
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
		LogDebug("Processing: %s", child.Text)
		var err error
		var childType *SupportedType
		if child.LinkAttr.Href != "" {
			throttle <- 1
			wg.Add(1)

			go downloadContent(child, toc, &wg)

			childType, err = getEntryType(child)
			if childType == nil && entryType != nil && (entryType.IsContainer || entryType.CascadeType) {
				LogDebug("Parent was container or cascade, using parent type of %s", entryType.TypeName)
				childType = entryType
			}

			if childType == nil {
				WarnIfError(err)
			} else if !childType.IsContainer {
				SaveSearchIndex(dbmap, child, childType, toc)
			} else {
				LogDebug("%s is a container. Do not index", child.Text)
			}
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
