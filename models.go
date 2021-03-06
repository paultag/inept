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

package inept

import (
	"database/sql"

	"pault.ag/go/archive"
	"pault.ag/go/debian/control"

	"github.com/jinzhu/gorm"
)

// Suite {{{

type Suite struct {
	gorm.Model

	Name        string `gorm:"unique_index:idx_suite"`
	Description string
	Origin      string
	Version     string
}

func (s Suite) Components(db *gorm.DB) ([]Component, error) {
	ret := []Component{}

	rows, err := db.Table("components").Where("suite_id = ?", s.ID).Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		el := Component{}
		db.ScanRows(rows, &el)
		ret = append(ret, el)
	}

	return ret, nil
}

// }}}

// Component {{{

type Component struct {
	gorm.Model

	/* Main */
	Name    string `gorm:"unique_index:idx_component"`
	SuiteID uint   `gorm:"unique_index:idx_component"`
	Suite   Suite
}

// }}}

// Arch {{{

type Arch struct {
	gorm.Model

	Name string
}

// }}}

// Binary {{{

type Binary struct {
	gorm.Model

	// Source   Source
	// SourceID int

	Name     string `gorm:"unique_index:idx_binary"`
	Version  string `gorm:"unique_index:idx_binary"`
	Arch     string `gorm:"unique_index:idx_binary"`
	Location string
}

func (b Binary) Paragraph(db *gorm.DB) (*control.Paragraph, error) {
	rows, err := db.Raw(`
	SELECT metadata_keys.name, binary_metadata.value
	FROM binary_metadata
	LEFT JOIN metadata_keys ON
		binary_metadata.key_id = metadata_keys.id
	WHERE binary_id = ?
	ORDER BY "metadata_keys.order"`, b.ID).Rows()

	if err != nil {
		return nil, err
	}
	return keyValueToParagraph(rows)
}

func (b Binary) Package(db *gorm.DB) (*archive.Package, error) {
	para, err := b.Paragraph(db)
	if err != nil {
		return nil, err
	}
	pkg := archive.Package{}
	return &pkg, control.UnpackFromParagraph(*para, &pkg)
}

// }}}

// Metadata {{{

type MetadataKey struct {
	gorm.Model

	Name  string `gorm:"unique"`
	Order int
}

// }}}

// BinaryMetadata {{{

type BinaryMetadata struct {
	gorm.Model

	Binary Binary
	Key    MetadataKey

	BinaryID uint `gorm:"unique_index:idx_binary_metadata"`
	KeyID    uint `gorm:"unique_index:idx_binary_metadata"`

	Value string
}

// }}}

// Binary Association {{{

type BinaryAssociation struct {
	gorm.Model

	Component Component
	Binary    Binary

	ComponentID uint `gorm:"unique_index:idx_binary_associations"`
	BinaryID    uint `gorm:"unique_index:idx_binary_associations"`
}

// }}}

// Helpers {{{

// keyValueToParagraph helper {{{

func keyValueToParagraph(rows *sql.Rows) (*control.Paragraph, error) {
	result := control.Paragraph{Order: []string{}, Values: map[string]string{}}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result.Set(key, value)
	}
	return &result, nil
}

// }}}

// }}}

// vim: foldmethod=marker
