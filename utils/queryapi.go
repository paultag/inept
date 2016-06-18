package utils

import (
	"database/sql"
	"github.com/jinzhu/gorm"

	"pault.ag/go/archive"
)

// Binary Iterator {{{

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

// }}}

// Write Packages {{{

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

// }}}

// vim: foldmethod=marker
