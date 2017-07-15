package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

var dbmap *gorp.DbMap
var dbName = "docSet.dsidx"

func InitDb(buildDir string) *gorp.DbMap {
	dbPath := filepath.Join(buildDir, dbName)
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	ExitIfError(err)

	db, err := sql.Open("sqlite3", dbPath)
	ExitIfError(err)

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.AddTableWithName(SearchIndex{}, "searchIndex").SetKeys(true, "ID")

	err = dbmap.CreateTablesIfNotExists()
	ExitIfError(err)

	err = dbmap.TruncateTables()
	ExitIfError(err)

	return dbmap
}

func SaveSearchIndex(dbmap *gorp.DbMap, entry TOCEntry, entryType *SupportedType, toc *AtlasTOC) {
	if entry.LinkAttr.Href == "" || entryType == nil {
		return
	}

	relLink := entry.GetContentFilepath(toc, false)
	name := entry.CleanTitle(*entryType)
	if entryType.ShowNamespace && len(entryHierarchy) > 0 {
		// Show namespace for methods
		name = entryHierarchy[len(entryHierarchy)-1] + "." + name
	}

	si := SearchIndex{
		Name: name,
		Type: entryType.TypeName,
		Path: relLink,
	}

	dbmap.Insert(&si)
	LogDebug("%s is indexed as a %s", entry.Text, entryType.TypeName)
}
