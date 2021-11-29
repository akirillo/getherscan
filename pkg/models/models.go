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

func (db *DB) ClearDB() error {
	tempDB := db.Session(&gorm.Session{AllowGlobalUpdate: true})

	// Delete transactions
	err := tempDB.Unscoped().Delete(&Transaction{}).Error
	if err != nil {
		return err
	}

	// Delete balances
	err = tempDB.Unscoped().Delete(&Balance{}).Error
	if err != nil {
		return err
	}

	// Delete blocks
	err = tempDB.Unscoped().Delete(&Block{}).Error
	if err != nil {
		return err
	}

	// Delete orphaned transactions
	err = tempDB.Unscoped().Delete(&OrphanedTransaction{}).Error
	if err != nil {
		return err
	}

	// Delete orphaned blocks
	err = tempDB.Unscoped().Delete(&OrphanedBlock{}).Error
	if err != nil {
		return err
	}

	return nil
}
