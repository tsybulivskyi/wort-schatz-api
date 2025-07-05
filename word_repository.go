package main

import (
	"gorm.io/gorm"
)

type WordRepository struct {
	DB *gorm.DB
}

// NewWordRepository initializes WordRepository with SQL Server
func NewWordRepository(db *gorm.DB) (*WordRepository, error) {
	return &WordRepository{DB: db}, nil
}

func (r *WordRepository) Create(word *Word) error {
	return r.DB.Create(word).Error
}

func (r *WordRepository) GetAll() ([]Word, error) {
	var words []Word
	err := r.DB.Preload("Tags").Find(&words).Error
	return words, err
}

func (r *WordRepository) DeleteAll() error {
	return r.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Word{}).Error
}
