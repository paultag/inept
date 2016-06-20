package inept

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

// Suite Iterator {{{

func NewSuiteIterator(query *gorm.DB) (*SuiteIterator, error) {
	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	return &SuiteIterator{
		db:    query,
		state: rows,
	}, nil
}

type SuiteIterator struct {
	db    *gorm.DB
	state *sql.Rows
}

func (b SuiteIterator) Next() (*Suite, bool, error) {
	suite := Suite{}
	if !b.state.Next() {
		return nil, false, nil
	}
	return &suite, true, b.db.ScanRows(b.state, &suite)
}

// }}}

// vim: foldmethod=marker
