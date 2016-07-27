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
	"strings"
	"sync"
)

// CSS Paths
var cssBasePath = "https://developer.salesforce.com/resource/stylesheets"
var cssFiles = []string{"holygrail.min.css", "docs.min.css", "syntax-highlighter.min.css"}

var wg sync.WaitGroup
var throttle = make(chan int, maxConcurrency)

const maxConcurrency = 16

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
	fmt.Println("Success:", toc.DocTitle, "-", toc.Version.VersionText, "-", toc.Version.DocVersion)
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
				SaveSearchIndex(dbmap, child, entryType, toc)
			} else if childType != nil && !childType.IsContainer {
				SaveSearchIndex(dbmap, child, childType, toc)
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
