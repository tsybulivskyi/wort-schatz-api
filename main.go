package main

import (
	"encoding/json"
	"fmt"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type WordDto struct {
	Id          uint     `json:"id"`
	Original    string   `json:"original"`
	Translation string   `json:"translation"`
	Tags        []string `json:"tags"`
}

// WordRepository is now in word_repository.go

var wordRepository *WordRepository

func init() {
	var err error
	_ = godotenv.Load()

	// Initialize GORM
	gormDB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// Auto-migrate the Word struct
	err = gormDB.AutoMigrate(&Word{})
	if err != nil {
		panic(err)
	}
	// Optionally keep the old sql.DB for legacy code

	wordRepository, _ = NewWordRepository(gormDB)
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

	if err := wordRepository.Create(&dbWord); err != nil {
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

	words, err := wordRepository.GetAll()
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
	wordRepository.DeleteAll()

	w.WriteHeader(http.StatusNoContent)
	return false
}

func jwtHandler(w http.ResponseWriter, r *http.Request) {

	// var tokenEncodeString string = "something"
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(100 * time.Minute).Unix()
	claims["authorized"] = true
	claims["user"] = "username"

	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		http.Error(w, "Error creating token"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"token": tokenString}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

}

func authenticationMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	var tokenString string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		tokenString = ""
	}
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		c.Abort()
		return
	}

	// token, err := verifyToken(tokenString)
	token, err := ValidateToken(c.Request.Context(), "https://your-issuer.com", "your-audience", tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	fmt.Printf("Token verified successfully: %v\n", token)

	c.Next()

}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return sampleSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil

}

func main() {
	router := gin.Default()

	router.GET("/hello", authenticationMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	router.POST("/words", func(c *gin.Context) {
		wordsHandler(c.Writer, c.Request)
	})
	router.GET("/jwt", func(c *gin.Context) {
		jwtHandler(c.Writer, c.Request)
	})

	router.Run(":8080")
	fmt.Println("Server running on :8080")

}
