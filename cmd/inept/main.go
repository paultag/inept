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

package main

import (
	"os"

	"github.com/urfave/cli"

	"pault.ag/go/archive"
	"pault.ag/go/inept"
	"pault.ag/go/inept/utils"

	"golang.org/x/crypto/openpgp"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func ohshit(err error) {
	if err != nil {
		panic(err)
	}
}

func ohshitdb(db *gorm.DB) {
	for _, err := range db.GetErrors() {
		panic(err)
	}
}

func getOpenPGPKey(keyid uint64) (*openpgp.Entity, error) {
	fd, err := os.Open("/home/paultag/keyring")
	if err != nil {
		return nil, err
	}
	el, err := openpgp.ReadKeyRing(fd)
	if err != nil {
		return nil, err
	}
	keys := el.KeysById(keyid)
	key := keys[0].Entity
	return key, nil
}

func getSQlDatabase(path string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", path)

}

func Init(repo inept.Repository) error {
	ohshitdb(utils.DropTables(repo.DB))
	ohshit(utils.Bootstrap(repo))
	return nil
}

func Write(repo inept.Repository) error {
	suites, err := utils.WriteSuites(repo, repo.DB.Table("suites"))
	ohshit(err)
	for _, suite := range suites {
		blobs, err := repo.Archive.Engross(*suite)
		ohshit(err)
		ohshit(repo.Archive.Link(blobs))
	}
	ohshit(repo.Archive.Decruft())
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "inept"
	app.Usage = "be the opposite of apt"
	app.Version = "0.0.1~alpha1"

	var archivePath string
	var keyringPath string
	var databasePath string
	var keyid uint = 0x00000000

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "archive",
			Usage:       "Load Archive from `FILE`",
			Destination: &archivePath,
		},
		cli.StringFlag{
			Name:        "database",
			Usage:       "sqlite database `FILE` to use",
			Destination: &databasePath,
		},
		cli.StringFlag{
			Name:        "keyring",
			Usage:       "OpenPGP Keyring `FILE` to use",
			Destination: &keyringPath,
		},
		cli.UintFlag{
			Name:        "keyid",
			Usage:       "OpenPGP `Key` ID to use",
			Destination: &keyid,
		},
	}

	var db *gorm.DB
	var targetArchive *archive.Archive
	var repo inept.Repository

	app.Before = func(c *cli.Context) error {
		var err error
		db, err = getSQlDatabase(databasePath)
		ohshit(err)
		key, err := getOpenPGPKey(uint64(keyid))
		ohshit(err)
		targetArchive, err = archive.New(archivePath, key)
		ohshit(err)
		mRepo, err := inept.NewRepository(db, targetArchive)
		ohshit(err)
		repo = *mRepo
		return nil
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "init",
			Action: func(c *cli.Context) error {
				return Init(repo)
			},
		},
		cli.Command{
			Name: "write",
			Action: func(c *cli.Context) error {
				return Write(repo)
			},
		},
	}

	app.Run(os.Args)
}

// vim: foldmethod=marker
