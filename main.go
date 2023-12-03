package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

const (
	CHARACTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@£$%^*"
)

var db *sql.DB

func Connect() {
	var err error

	db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=urls sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWD")))

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	Connect()

	fmt.Println("Init")

	http.HandleFunc("/create", CreateUrl)
	http.HandleFunc("/", GetUrl)

	http.ListenAndServe(":8080", nil)
}

// POST /create?url=<url>
func CreateUrl(w http.ResponseWriter, r *http.Request) {
	var err error
	url := r.URL.Query().Get("url")

	if url == "" {
		w.WriteHeader(400)
		return
	}

	code := GenerateRandomCode()

	_, err = db.Exec("INSERT INTO urls VALUES ($1 , $2)", code, url)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	log.Println("Created url: " + url + " with code: " + code)

	w.WriteHeader(201)
	w.Write([]byte(code))
}

func GetUrl(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	fmt.Println(code)

	var url string

	if err := db.QueryRow("SELECT url FROM urls WHERE code = $1", code).Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			log.Println("Served Home page")

			url := strings.TrimPrefix(r.URL.Path, "/")

			if url == "" {
				http.ServeFile(w, r, "static/index.html")
				return
			}

			http.ServeFile(w, r, "static/"+url)
			return
		}
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	log.Println("Redirecting to: " + url + " with code: " + code)

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func GenerateRandomCode() string {
	var code string = ""

	for i := 0; i < 7; i++ {
		code += string(CHARACTERS[rand.Intn(len(CHARACTERS))])
	}

	return code
}
