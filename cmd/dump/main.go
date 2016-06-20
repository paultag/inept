package main

import (
	"log"
	"os"

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

	suites, err := inept.NewSuiteIterator(db.Table("suites"))
	ohshit(err)

	for {
		suite, next, err := suites.Next()
		ohshit(err)
		if !next {
			break
		}

		archiveSuite, err := infra.Suite(suite.Name)
		ohshit(err)

		archiveSuite.Description = suite.Description
		archiveSuite.Origin = suite.Origin
		archiveSuite.Version = suite.Version

		components, err := suite.Components(db)
		ohshit(err)

		for _, component := range components {
			comp, err := archiveSuite.Component(component.Name)
			ohshit(err)

			if err := utils.WritePackages(
				comp,
				db,
				db.Raw(`
			SELECT * FROM
			binaries
			LEFT JOIN binary_associations ON
				binary_associations.binary_id = binaries.id
			WHERE binary_associations.component_id = ?
			ORDER BY binaries.name`, component.ID),
			); err != nil {
				panic(err)
			}

		}

		log.Println("Engrossing")
		blobs, err := infra.Engross(*archiveSuite)
		ohshit(err)
		log.Println("Linking")
		ohshit(infra.Link(blobs))
	}

	log.Println("Decrufting")
	ohshit(infra.Decruft())
}

// vim: foldmethod=marker
