package inept

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

func NewBinaryIterator(query *gorm.DB) (*BinaryIterator, error) {
	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	return &BinaryIterator{
		db:    query,
		state: rows,
	}, nil
}

type BinaryIterator struct {
	db    *gorm.DB
	state *sql.Rows
}

func (b BinaryIterator) Next() (*Binary, bool, error) {
	binary := Binary{}
	if !b.state.Next() {
		return nil, false, nil
	}
	return &binary, true, b.db.ScanRows(b.state, &binary)
}
