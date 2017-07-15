package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var dbmap *gorp.DbMap

func InitDb() *gorp.DbMap {
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
