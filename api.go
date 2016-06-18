package inept

import (
	"github.com/jinzhu/gorm"
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

// vim: foldmethod=marker
