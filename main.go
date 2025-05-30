package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WordDto struct {
	Id          uint     `json:"id"`
	Original    string   `json:"original"`
	Translation string   `json:"translation"`
	Tags        []string `json:"tags"`
}

// WordRepository is now in word_repository.go

var db *sql.DB
var gormDB *gorm.DB

func init() {
	var err error
	_ = godotenv.Load()
	connStr := os.Getenv("DB_CONN_STRING")
	if connStr == "" {
		panic("DB_CONN_STRING environment variable not set")
	}
	// Initialize GORM
	gormDB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// Auto-migrate the Word struct
	err = gormDB.AutoMigrate(&Word{})
	if err != nil {
		panic(err)
	}
	// Optionally keep the old sql.DB for legacy code
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
}

func saveWordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var word WordDto
	if err := json.NewDecoder(r.Body).Decode(&word); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	dbWord := Word{
		Original:    word.Original,
		Translation: word.Translation,
	}
	// Convert string tags to Tag objects
	for _, tagName := range word.Tags {
		tag := Tag{Name: tagName}
		dbWord.Tags = append(dbWord.Tags, tag)
	}
	if err := gormDB.Create(&dbWord).Error; err != nil {
		http.Error(w, "Failed to save word", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (w Word) ConvertToDto() WordDto {
	tags := make([]string, len(w.Tags))
	for i, t := range w.Tags {
		tags[i] = t.Name
	}
	return WordDto{
		Id:          w.Model.ID,
		Original:    w.Original,
		Translation: w.Translation,
		Tags:        tags,
	}
}

func getWordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var words []Word
	err := gormDB.Preload("Tags").Find(&words).Error
	if err != nil {
		http.Error(w, "Failed to fetch words", http.StatusInternalServerError)
		return
	}
	var result []WordDto
	for _, w := range words {
		result = append(result, w.ConvertToDto())
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func wordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		saveWordHandler(w, r)
		return
	} else if r.Method == http.MethodGet {
		getWordsHandler(w, r)
		return
	} else if r.Method == http.MethodDelete {
		deleteAllWords(w)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func deleteAllWords(w http.ResponseWriter) bool {
	if err := gormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Word{}).Error; err != nil {
		http.Error(w, "Failed to delete words", http.StatusInternalServerError)
		return true
	}
	w.WriteHeader(http.StatusNoContent)
	return false
}

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	http.HandleFunc("/words", wordsHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
