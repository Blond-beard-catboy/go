package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "alex"
	password = "123321"
	dbname   = "test_db"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	flagPort = flag.String("port", "9000", "Port to listen on")
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	createProductTable(db)
	fmt.Println("Successfully connected to the database!")

	mux := http.NewServeMux()
	mux.HandleFunc("/", GetHandler)
	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		PostHandler(w, r, db)
	})

	log.Printf("Listening on port %s", *flagPort)
	log.Fatal(http.ListenAndServe(":"+*flagPort, mux))
}

var results []string

func GetHandler(w http.ResponseWriter, r *http.Request) {
	jsonBody, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error converting results to JSON", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}

// func PostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Error reading request body", http.StatusInternalServerError)
// 		return
// 	}
// 	results = append(results, string(body))

// 	var person Person
// 	if err := json.Unmarshal(body, &person); err != nil {
// 		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
// 		return
// 	}

// 	pk, err := insertProduct(db, person)
// 	if err != nil {
// 		http.Error(w, "Error inserting into database", http.StatusInternalServerError)
// 		return
// 	}

// 	fmt.Fprintf(w, "POST done. Inserted ID: %d", pk)
// }

func PostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var person Person
	if err := json.Unmarshal(body, &person); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	pk, err := insertPerson(db, person)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		http.Error(w, "Error inserting into database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "POST done. Inserted ID: %d", pk)
}

func createProductTable(db *sql.DB) {
	query := `
 CREATE TABLE IF NOT EXISTS product (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  age INTEGER NOT NULL
 )`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func insertPerson(db *sql.DB, person Person) (int, error) {
	query := `INSERT INTO person (name, age) VALUES ($1, $2) RETURNING id`
	var pk int
	err := db.QueryRow(query, person.Name, person.Age).Scan(&pk)
	if err != nil {
		return 0, fmt.Errorf("error inserting person: %w", err)
	}
	return pk, nil
}
