package main

import (
	"fmt"
	"pault.ag/go/archive"
	"pault.ag/go/inept"
	"pault.ag/go/inept/indexer"

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

	ohshitdb(db.DropTableIfExists(
		&inept.Suite{},
		&inept.Component{},
		&inept.Arch{},
		&inept.Binary{},
		&inept.MetadataKey{},
		&inept.BinaryMetadata{},
		&inept.BinaryAssociation{},
	))

	ohshitdb(db.CreateTable(
		&inept.Suite{},
		&inept.Component{},
		&inept.Arch{},
		&inept.Binary{},
		&inept.MetadataKey{},
		&inept.BinaryMetadata{},
		&inept.BinaryAssociation{},
	))

	arch, err := archive.New("/home/paultag/tmp/infra", nil)
	ohshit(err)

	// Key metadata {{{
	for name, order := range map[string]int{
		"Package":               -2600,
		"Source":                -2500,
		"Binary":                -2400,
		"Version":               -2300,
		"Essential":             -2250,
		"Installed-Size":        -2200,
		"Maintainer":            -2100,
		"Uploaders":             -2090,
		"Original-Maintainer":   -2080,
		"Build-Depends":         -2000,
		"Build-Depends-Indep":   -1990,
		"Build-Conflicts":       -1980,
		"Build-Conflicts-Indep": -1970,
		"Architecture":          -1800,
		"Standards-Version":     -1700,
		"Format":                -1600,
		"Files":                 -1500,
		"Dm-Upload-Allowed":     -1400,
		"Vcs-Mtn":               -1300,
		"Vcs-Browse":            -1300,
		"Vcs-Browser":           -1300,
		"Vcs-Git":               -1300,
		"Vcs-Svn":               -1300,
		"Vcs-Cvs":               -1300,
		"Vcs-Darcs":             -1300,
		"Vcs-Bzr":               -1300,
		"Vcs-Arch":              -1300,
		"Vcs-Hg":                -1300,
		"Checksums-Sha1":        -1200,
		"Checksums-Sha256":      -1200,
		"Checksums-Sha512":      -1200,
		"Replaces":              -1100,
		"Provides":              -1000,
		"Depends":               -900,
		"Pre-Depends":           -850,
		"Recommends":            -800,
		"Suggests":              -700,
		"Enhances":              -650,
		"Conflicts":             -600,
		"Breaks":                -500,
		"Description":           -400,
		"Origin":                -300,
		"Bugs":                  -200,
		"Multi-Arch":            -150,
		"Homepage":              -100,
	} {
		mKey := inept.MetadataKey{Name: name, Order: order}
		ohshitdb(db.Create(&mKey))
	}
	// }}}

	fmt.Printf("Indexing debs\n")
	ohshit(indexer.IndexDebs(db, *arch, []string{"sha256", "sha512", "sha1"}))
	fmt.Printf("Indexing suites\n")
	ohshit(indexer.IndexSuites(db, *arch))
}

// vim: foldmethod=marker
