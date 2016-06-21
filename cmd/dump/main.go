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
	"log"
	"os"

	"pault.ag/go/archive"
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

func getSQlDatabase() (*gorm.DB, error) {
	return gorm.Open("sqlite3", "test.db")

}

func main() {
	db, err := getSQlDatabase()
	ohshit(err)

	key, err := getOpenPGPKey(0xEE07B8B6CB89FDDB)
	ohshit(err)

	infra, err := archive.New("./infra/", key)
	ohshit(err)

	log.Println("Writing Suites")
	suites, err := utils.WriteSuites(infra, db, db.Table("suites"))
	ohshit(err)

	for _, suite := range suites {
		log.Println("Engrossing")
		blobs, err := infra.Engross(*suite)
		ohshit(err)
		log.Println("Linking")
		ohshit(infra.Link(blobs))
	}

	log.Println("Decrufting")
	ohshit(infra.Decruft())
}

// vim: foldmethod=marker
