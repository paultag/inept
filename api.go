package inept

import (
	"github.com/jinzhu/gorm"

	"pault.ag/go/archive"
)

func Suites(db *gorm.DB, query *gorm.DB) ([]Suite, error) {
	suites := []Suite{}
	rows, err := query.Rows()
	if err != nil {
		return suites, err
	}

	for rows.Next() {
		suite := Suite{}
		if err := db.ScanRows(rows, &suite); err != nil {
			return []Suite{}, err
		}
		suites = append(suites, suite)
	}
	return suites, nil
}

func WritePackages(component *archive.Component, db *gorm.DB, query *gorm.DB) error {
	binaries, err := NewBinaryIterator(query)
	if err != nil {
		return err
	}

	for {
		binary, next, err := binaries.Next()
		if err != nil {
			return err
		}

		if !next {
			break
		}

		pkg, err := binary.Package(db)
		if err != nil {
			return err
		}

		if err := component.AddPackage(*pkg); err != nil {
			return err
		}
	}

	return nil
}
