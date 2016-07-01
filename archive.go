package inept

import (
	"path"

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
	debPath, obj, err := repo.Archive.Pool.IncludeDeb(debFile)
	if err != nil {
		return err
	}

	fd, err := repo.Archive.Open(*obj)
	if err != nil {
		return err
	}
	defer fd.Close()
	debFile, err = deb.Load(fd, path.Join(repo.Archive.Path(), debPath))
	if err != nil {
		return err
	}

	return IndexDeb(repo.DB, *repo.Archive, []string{
		"sha1", "sha256", "sha512",
	}, debFile)
}
