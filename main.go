package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type WorkLog struct {
	Date    string
	Title   string
	Minutes int
}

func main() {

	dsn := "root:Srivenkat@67@tcp(localhost:3306)/work_tracker"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("❌ Database error:", err)
	}
	fmt.Println("✅ MySQL Connected")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/add", addWorkHandler)

	fmt.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))

	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err := db.Exec("INSERT INTO users(username,password) VALUES(?,?)", username, hash)
	if err != nil {
		tmpl.Execute(w, "Username already exists")
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&hash)
	if err != nil {
		tmpl.Execute(w, "Invalid username or password")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		tmpl.Execute(w, "Invalid username or password")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT work_date, title, duration_minutes FROM work_logs ORDER BY work_date DESC`)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()

	var logs []WorkLog
	for rows.Next() {
		var l WorkLog
		rows.Scan(&l.Date, &l.Title, &l.Minutes)
		logs = append(logs, l)
	}

	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	tmpl.Execute(w, logs)
}

func addWorkHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/add.html"))

	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	date := r.FormValue("work_date")
	title := r.FormValue("title")
	minutes := r.FormValue("minutes")

	_, err := db.Exec("INSERT INTO work_logs(work_date,title,duration_minutes) VALUES(?,?,?)",
		date, title, minutes)

	if err != nil {
		http.Error(w, "Insert failed", 500)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
