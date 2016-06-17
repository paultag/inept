package inept

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"pault.ag/go/archive"
	"pault.ag/go/debian/control"
	"pault.ag/go/debian/deb"
	"pault.ag/go/debian/transput"

	"github.com/jinzhu/gorm"
)

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

type Component struct {
	gorm.Model

	/* Main */
	Name    string `gorm:"unique_index:idx_component"`
	SuiteID uint   `gorm:"unique_index:idx_component"`
	Suite   Suite
}

type Arch struct {
	gorm.Model

	Name string
}

func InsertDeb(db *gorm.DB, debFile deb.Deb, location string, size uint64, hashers []*transput.Hasher) (error, *Binary) {
	binary := Binary{}
	for _, err := range db.FirstOrCreate(&binary, Binary{
		Name:    debFile.Control.Package,
		Version: debFile.Control.Version.String(),
		Arch:    debFile.Control.Architecture.String(),
	}).GetErrors() {
		return err, nil
	}

	binary.Location = location
	for _, err := range db.Save(&binary).GetErrors() {
		return err, nil
	}

	debControl := debFile.Control.Paragraph
	debControl.Set("Filename", location)
	debControl.Set("Size", strconv.FormatUint(size, 10))

	for _, hasher := range hashers {
		debControl.Set(strings.ToUpper(hasher.Name()), fmt.Sprintf("%x", hasher.Sum(nil)))
	}

	for key, value := range debControl.Values {
		mKey := MetadataKey{}

		for _, err := range db.FirstOrCreate(&mKey, MetadataKey{
			Name: key,
		}).GetErrors() {
			return err, nil
		}

		meta := BinaryMetadata{}
		for _, err := range db.FirstOrCreate(&meta, BinaryMetadata{
			BinaryID: binary.ID,
			KeyID:    mKey.ID,
		}).GetErrors() {
			return err, nil
		}
		meta.Value = value
		for _, err := range db.Save(&mKey).Save(&meta).GetErrors() {
			return err, nil
		}
	}
	return nil, &binary
}

type Binary struct {
	gorm.Model

	// Source   Source
	// SourceID int

	Name     string `gorm:"unique_index:idx_binary"`
	Version  string `gorm:"unique_index:idx_binary"`
	Arch     string `gorm:"unique_index:idx_binary"`
	Location string
}

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

type MetadataKey struct {
	gorm.Model

	Name  string `gorm:"unique"`
	Order int
}

type BinaryMetadata struct {
	gorm.Model

	Binary Binary
	Key    MetadataKey

	BinaryID uint `gorm:"unique_index:idx_binary_metadata"`
	KeyID    uint `gorm:"unique_index:idx_binary_metadata"`

	Value string
}

type BinaryAssociation struct {
	gorm.Model

	Component Component
	Binary    Binary

	ComponentID uint `gorm:"unique_index:idx_binary_associations"`
	BinaryID    uint `gorm:"unique_index:idx_binary_associations"`
}
