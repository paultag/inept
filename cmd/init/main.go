package main

import (
	"fmt"
	"pault.ag/go/archive"
	"pault.ag/go/inept/indexer"
	"pault.ag/go/inept/utils"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func ohshit(err error) {
	if err != nil {
		panic(err)
	}
}

func ohshitdb(db *gorm.DB) {
	for _, err := range db.GetErrors() {
		panic(err)
	}
}

func main() {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}

	ohshitdb(utils.DropTables(db))

	arch, err := archive.New("/home/paultag/tmp/infra", nil)
	ohshit(err)

	ohshitdb(utils.Bootstrap(db, arch))
}

// vim: foldmethod=marker
