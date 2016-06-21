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

func main() {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}

	fd, err := os.Open("/home/paultag/keyring")
	ohshit(err)
	el, err := openpgp.ReadKeyRing(fd)
	ohshit(err)
	keys := el.KeysById(0xEE07B8B6CB89FDDB)
	key := keys[0].Entity

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
