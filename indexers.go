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
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"

	"pault.ag/go/archive"
	"pault.ag/go/debian/control"
	"pault.ag/go/debian/deb"
	"pault.ag/go/debian/transput"
)

type Indexer func(dir, file string) error

// Index helper {{{

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

// }}}

// Index Debian .deb files {{{

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

		err, _ = InsertDeb(
			db,
			*debFile,
			debPath,
			uint64(fileInfo.Size()),
			hashers,
		)
		return err
	})
}

// }}}

// Index Suites {{{

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

		suite := Suite{}

		for _, err := range db.FirstOrCreate(&suite, Suite{
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

			comp := Component{}
			for _, err := range db.FirstOrCreate(&comp, Component{
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

				bin := Binary{}
				for _, err := range db.First(&bin, "location = ?", pkg.Filename).GetErrors() {
					return err
				}

				if err != nil {
					return err
				}

				assn := BinaryAssociation{}

				for _, err := range db.FirstOrCreate(&assn, BinaryAssociation{
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

// }}}

// vim: foldmethod=marker
