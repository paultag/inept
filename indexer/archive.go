package indexer

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/jinzhu/gorm"

	"pault.ag/go/archive"
	"pault.ag/go/debian/control"
	"pault.ag/go/debian/deb"
	"pault.ag/go/debian/transput"
	"pault.ag/go/inept"
)

type Indexer func(dir, file string) error

func Index(archive archive.Archive, i Indexer) error {
	paths, err := archive.Paths()
	if err != nil {
		return err
	}

	for fpath, _ := range paths {
		dir, file := path.Split(fpath)
		if err := i(dir, file); err != nil {
			return err
		}
	}

	return nil
}

func IndexDebs(db *gorm.DB, a archive.Archive, hashes []string) error {
	return Index(a, func(dir, file string) error {
		if !strings.HasSuffix(file, ".deb") {
			return nil
		}

		debFile, closer, err := deb.LoadFile(path.Join(dir, file))
		if err != nil {
			return err
		}
		defer closer()

		/* So much SIGH about to go down here */
		fd, err := os.Open(debFile.Path)
		if err != nil {
			return err
		}
		fileInfo, err := fd.Stat()
		if err != nil {
			return err
		}

		writers := []io.Writer{}
		hashers := []*transput.Hasher{}

		for _, hash := range hashes {
			hasher, err := transput.NewHasher(hash)
			if err != nil {
				return err
			}
			writers = append(writers, hasher)
			hashers = append(hashers, hasher)
		}

		io.Copy(io.MultiWriter(writers...), fd)

		debPath := path.Clean(debFile.Path)
		archivePath := path.Clean(a.Path())

		if !strings.HasPrefix(debPath, archivePath) {
			return fmt.Errorf(".deb is outside the Archive root")
		}

		debPath = debPath[len(archivePath)+1:]

		err, _ = inept.InsertDeb(
			db,
			*debFile,
			debPath,
			uint64(fileInfo.Size()),
			hashers,
		)
		return err
	})
}

func IndexSuites(db *gorm.DB, a archive.Archive) error {
	return Index(a, func(dir, file string) error {
		if file != "Release" {
			return nil
		}
		fd, err := os.Open(path.Join(dir, file))
		if err != nil {
			return err
		}
		release := archive.Release{}
		control.Unmarshal(&release, fd)

		suite := inept.Suite{}

		for _, err := range db.FirstOrCreate(&suite, inept.Suite{
			Name: release.Suite,
		}).GetErrors() {
			return err
		}

		suite.Description = release.Description
		suite.Origin = release.Origin
		suite.Version = release.Version

		for _, err := range db.Save(&suite).GetErrors() {
			return err
		}

		for index, _ := range release.Indices() {
			dirs := strings.Split(index, "/")
			if len(dirs) <= 0 {
				continue
			}
			component := dirs[0]

			comp := inept.Component{}
			for _, err := range db.FirstOrCreate(&comp, inept.Component{
				Name:    component,
				SuiteID: suite.ID,
			}).GetErrors() {
				return err
			}

			packages, err := archive.LoadPackagesFile(path.Join(dir, index))
			if err != nil {
				return err
			}

			for {
				pkg, err := packages.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				bin := inept.Binary{}
				for _, err := range db.First(&bin, "location = ?", pkg.Filename).GetErrors() {
					return err
				}

				if err != nil {
					return err
				}

				assn := inept.BinaryAssociation{}

				for _, err := range db.FirstOrCreate(&assn, inept.BinaryAssociation{
					BinaryID:    bin.ID,
					ComponentID: comp.ID,
				}).GetErrors() {
					return err
				}
			}
		}

		return nil
	})
}
