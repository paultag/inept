package utils

import (
	"database/sql"
	"github.com/jinzhu/gorm"

	"pault.ag/go/archive"
	"pault.ag/go/inept"
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

func (b BinaryIterator) Next() (*inept.Binary, bool, error) {
	binary := inept.Binary{}
	if !b.state.Next() {
		return nil, false, nil
	}
	return &binary, true, b.db.ScanRows(b.state, &binary)
}

// }}}

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

func (b SuiteIterator) Next() (*inept.Suite, bool, error) {
	suite := inept.Suite{}
	if !b.state.Next() {
		return nil, false, nil
	}
	return &suite, true, b.db.ScanRows(b.state, &suite)
}

// }}}

// Write Suites {{{

func WriteSuites(arch *archive.Archive, db *gorm.DB, query *gorm.DB) ([]*archive.Suite, error) {
	ret := []*archive.Suite{}

	suites, err := NewSuiteIterator(query)
	if err != nil {
		return []*archive.Suite{}, err
	}

	for {
		suite, next, err := suites.Next()
		if err != nil {
			return []*archive.Suite{}, err
		}

		if !next {
			break
		}

		archiveSuite, err := arch.Suite(suite.Name)
		if err != nil {
			return []*archive.Suite{}, err
		}

		archiveSuite.Description = suite.Description
		archiveSuite.Origin = suite.Origin
		archiveSuite.Version = suite.Version

		components, err := suite.Components(db)
		if err != nil {
			return []*archive.Suite{}, err
		}

		for _, component := range components {
			comp, err := archiveSuite.Component(component.Name)
			if err != nil {
				return []*archive.Suite{}, err
			}

			if err := WritePackages(
				comp,
				db,
				db.Raw(`
			SELECT * FROM
			binaries
			LEFT JOIN binary_associations ON
				binary_associations.binary_id = binaries.id
			WHERE binary_associations.component_id = ?
			ORDER BY binaries.name`, component.ID),
			); err != nil {
				return []*archive.Suite{}, err
			}
		}

		ret = append(ret, archiveSuite)
	}
	return ret, nil
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
