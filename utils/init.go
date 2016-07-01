/* {{{ Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2016
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE. }}} */

package utils

import (
	"pault.ag/go/inept"

	"github.com/jinzhu/gorm"
)

func DropTables(db *gorm.DB) *gorm.DB {
	return db.DropTableIfExists(
		&inept.Suite{},
		&inept.Component{},
		&inept.Arch{},
		&inept.Binary{},
		&inept.MetadataKey{},
		&inept.BinaryMetadata{},
		&inept.BinaryAssociation{},
	)
}

func CreateTables(db *gorm.DB) *gorm.DB {
	return db.CreateTable(
		&inept.Suite{},
		&inept.Component{},
		&inept.Arch{},
		&inept.Binary{},
		&inept.MetadataKey{},
		&inept.BinaryMetadata{},
		&inept.BinaryAssociation{},
	)
}

func CreateMetadataKeys(db *gorm.DB) *gorm.DB {
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
		db.Create(&mKey)
	}
	return db
}

func Bootstrap(repo inept.Repository) error {
	for _, err := range CreateTables(repo.DB).GetErrors() {
		return err
	}

	for _, err := range CreateMetadataKeys(repo.DB).GetErrors() {
		return err
	}

	if err := inept.IndexDebs(repo.DB, *repo.Archive, []string{
		"sha256",
		"sha512",
		"sha1",
	}); err != nil {
		return err
	}

	if err := inept.IndexSuites(repo.DB, *repo.Archive); err != nil {
		return err
	}

	return nil
}

// vim: foldmethod=marker
