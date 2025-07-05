package main

import (
	"gorm.io/gorm"
)

type Word struct {
	gorm.Model
	Original    string
	Translation string
	Tags        []Tag `gorm:"many2many:word_tags;"`
}

type Tag struct {
	gorm.Model
	Name  string `json:"name"`
	Color string `json:"color,omitempty"` // Optional field for color
}
