package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// database connection
var db *sql.DB

func main() {
	var err error
	dataSourceString := "root:Thousand@90@tcp(127.0.0.1:3306)/go_new"
	db, err = sql.Open("mysql", dataSourceString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("database connected..")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", apiHealthHandler)
	r.Route("/users", func(r chi.Router) {
		r.Get("/", usersHandler)
		r.Get("/{userId}", userHandler)
		r.Post("/", createUserHandler)
		r.Put("/{userId}", updateHandler)
		r.Delete("/{userId}", deleteUser)
	})

	log.Println("server starting on port: 4000")
	log.Fatal(http.ListenAndServe(":4000", r))
}

func apiHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The api is healthy!"))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id,name,email FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	userId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var u User
	errR := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userId).Scan(&u.ID, &u.Name, &u.Email)
	if errR != sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if errR != nil {
		http.Error(w, errR.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := result.LastInsertId()
	u.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, errU := db.Exec("UPDATE users SET name = ? email = ? WHERE id = ?", u.Name, u.Email, id)
	if errU != nil {
		http.Error(w, errU.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, errU := db.Exec("DELETE FROM users WHERE id = ?", id)
	if errU != nil {
		http.Error(w, errU.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
