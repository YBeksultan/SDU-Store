package main

import (
	"html/template"
	"net/http"
)

type User struct {
	Id                                 int
	Username, Surname, Password, Email string
}

var users []User

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("home.html")
	t.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		newUser := User{Username: username, Password: password}
		users = append(users, newUser)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("register.html")
		t.Execute(w, nil)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		for _, user := range users {
			if user.Username == username && user.Password == password {
				http.Redirect(w, r, "/welcome", http.StatusSeeOther)
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	}
}
