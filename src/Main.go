package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	ID                                 int
	Username, Surname, Password, Email string
}

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/maindb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/register", showRegisterForm).Methods("GET")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/login", showLoginForm).Methods("GET")
	router.HandleFunc("/welcome", welcome)
	http.ListenAndServe(":8080", router)
}

func home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/home.html")
	t.Execute(w, nil)
}

func showRegisterForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/register.html")
	t.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err = db.QueryRow("SELECT id FROM users WHERE username=?", username).Scan(&user.ID)
	if err != sql.ErrNoRows {
		t, _ := template.ParseFiles("templates/register.html")
		t.Execute(w, "Username already taken.")
		return
	}

	_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, password)
	if err != nil {
		panic(err.Error())
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func showLoginForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/login.html")
	t.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err = db.QueryRow("SELECT id, username, password FROM users WHERE username=? AND password=?", username, password).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		t, _ := template.ParseFiles("templates/login.html")
		t.Execute(w, "Invalid username or password.")
		return
	}

	t, _ := template.ParseFiles("templates/welcome.html")
	t.Execute(w, user)

}

func welcome(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/welcome.html")
	t.Execute(w, nil)
}
