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
	"fmt"
	"os"

	"github.com/urfave/cli"

	"pault.ag/go/archive"
	"pault.ag/go/debian/deb"
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

func getOpenPGPKey(keyring string, keyid uint64) (*openpgp.Entity, error) {
	fd, err := os.Open(keyring)
	if err != nil {
		return nil, err
	}
	el, err := openpgp.ReadKeyRing(fd)
	if err != nil {
		return nil, err
	}
	keys := el.KeysById(keyid)
	if len(keys) == 0 {
		return nil, fmt.Errorf(
			"No key with the given KeyID (%x) in the Keyring",
			keyid,
		)
	}
	key := keys[0].Entity
	return key, nil
}

func getSQlDatabase(path string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", path)

}

func IncludeDebPath(repo inept.Repository, path string) (*inept.Binary, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	debFile, err := deb.Load(fd, path)
	if err != nil {
		return nil, err
	}
	return repo.IncludeDeb(debFile)
}

func Associate(repo inept.Repository, c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		return cli.ShowCommandHelp(c, "associate")
	}

	component, err := utils.ComponentStringToComponent(repo.DB, args[1])
	if err != nil {
		return err
	}

	binaries, err := utils.BinaryStringToIterator(repo.DB, args[0])
	if err != nil {
		return err
	}
	defer binaries.Close()
	existingDebs := []*inept.Binary{}

	for {
		binary, next, err := binaries.Next()
		if err != nil {
			return err
		}

		if !next {
			break
		}

		existingDebs = append(existingDebs, binary)

	}
	binaries.Close()
	if len(existingDebs) == 0 {
		return fmt.Errorf("No binaries returned")
	}

	for _, deb := range existingDebs {
		fmt.Printf("Associating %s with %s\n", deb.Name, component.Name)
		if _, err := repo.AssociateBinary(deb, component); err != nil {
			return err
		}
	}

	return nil
}

func IncludeDeb(repo inept.Repository, c *cli.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return cli.ShowCommandHelp(c, "includedeb")
	}
	for _, filePath := range args {
		_, err := IncludeDebPath(repo, filePath)
		if err != nil {
			panic(err)
		}
	}
	return nil
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

	var configPath string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "Load configuration from `FILE`",
			Destination: &configPath,
		},
	}

	var db *gorm.DB
	var targetArchive *archive.Archive
	var repo inept.Repository

	app.Before = func(c *cli.Context) error {
		fd, err := os.Open(configPath)
		ohshit(err)
		config, err := utils.ParseConfiguration(fd)
		ohshit(err)

		db, err = getSQlDatabase(config.Global.Database)
		ohshit(err)
		keyID, err := config.Global.KeyIDInt()
		ohshit(err)
		key, err := getOpenPGPKey(config.Global.Keyring, keyID)
		ohshit(err)
		targetArchive, err = archive.New(config.Global.Archive, key)
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
				ohshit(Init(repo))
				return nil
			},
		},
		cli.Command{
			Name: "write",
			Action: func(c *cli.Context) error {
				ohshit(Write(repo))
				return nil
			},
		},
		cli.Command{
			Name:      "includedeb",
			ArgsUsage: "[debpath]",
			Action: func(c *cli.Context) error {
				ohshit(IncludeDeb(repo, c))
				return nil
			},
		},
		cli.Command{
			Name:      "associate",
			ArgsUsage: "[binary component]",
			Action: func(c *cli.Context) error {
				ohshit(Associate(repo, c))
				return nil
			},
		},
	}

	app.Run(os.Args)
}

// vim: foldmethod=marker
