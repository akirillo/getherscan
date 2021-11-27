package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) Initialize(connectionString string) error {
	var err error
	db.DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.InitializeModels()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) InitializeModels() error {
	return db.AutoMigrate(
		&Block{},
		&OrphanedBlock{},
		&Transaction{},
		&OrphanedTransaction{},
		&Balance{},
	)
}
