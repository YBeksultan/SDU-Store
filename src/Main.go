package main

import (
	_ "fmt"
	"html/template"
	"net/http"
)

type User struct {
	Username string
	Password string
}

var users []User

func main() {
	http.HandleFunc("/", homeHandler)
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
