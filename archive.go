package inept

import (
	"pault.ag/go/archive"
	"pault.ag/go/debian/deb"

	"github.com/jinzhu/gorm"
)

type Repository struct {
	DB      *gorm.DB
	Archive *archive.Archive
}

func NewRepository(db *gorm.DB, arch *archive.Archive) (*Repository, error) {
	return &Repository{
		DB:      db,
		Archive: arch,
	}, nil
}

func (repo Repository) IncludeDeb(debFile *deb.Deb) error {
	_, _, err := repo.Archive.Pool.IncludeDeb(debFile)
	return err
}
