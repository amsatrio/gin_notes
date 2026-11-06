package main

import (
	"log"

	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/model"
)

func init() {
	initializer.LoadEnvironmentVariables()
	initializer.ConnectToDB()
}

func main() {
	err := initializer.DB.AutoMigrate(
		&model.MRole{},
		&model.MBiodata{},
		&model.MUser{},
		&model.MNotes{},
		&model.TToken{},
		&model.TResetPassword{},
	)
	if err != nil {
		log.Fatal("migration failed: ", err)
	}

	log.Println("migration completed successfully")
}
