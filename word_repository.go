package main

import "gorm.io/gorm"

type WordRepository struct {
	DB *gorm.DB
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
