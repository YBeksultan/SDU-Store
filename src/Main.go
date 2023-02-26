package main

import (
	_ "fmt"
	"html/template"
	"net/http"
	"strconv"
)

type User struct {
	Id                                 int
	Username, Surname, Password, Email string
}

var users []User

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		name := r.FormValue("username")
		surname := r.FormValue("surname")
		password := r.FormValue("password")
		email := r.FormValue("email")
		for _, user := range users {
			if user.Id == id {
				http.Error(w, "Username already taken", http.StatusBadRequest)
				return
			}
		}

		users = append(users, User{Id: id, Username: name, Surname: surname, Password: password, Email: email})

		t, err := template.ParseFiles("templates/register-success.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	}
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		password := r.FormValue("password")

		for _, user := range users {
			if user.Id == id && user.Password == password {
				t, err := template.ParseFiles("templates/login-success.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				t.Execute(w, nil)
				return
			}
		}

		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
}
