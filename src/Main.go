package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type User struct {
	Id                                 int
	Username, Surname, Password, Email string
}

type Item struct {
	ItemId    int
	ItemName  string
	ItemPrice float64
	ItemImage string
}

var store = sessions.NewCookieStore([]byte(os.Getenv(generateSessionKey())))
var db *sql.DB
var err error

func main() {
	db, err = sql.Open("mysql", "root:9Wyk%L7nUvm4@/sdu_store")
	if err != nil {
		panic(err.Error())
	}
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/catalog", catalogHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", nil)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "home", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.Values["registered"] != nil {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "register", nil)
	} else if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		name := r.FormValue("username")
		surname := r.FormValue("surname")
		password := r.FormValue("password")
		email := r.FormValue("email")
		var user User = User{id, name, surname, password, email}
		query := "INSERT INTO users VALUES ('" + strconv.Itoa(user.Id) + "','" + user.Username +
			"','" + user.Surname + "','" + user.Password + "','" + user.Email + "');"
		insert, err := db.Query(query)
		if err != nil {
			fmt.Println(query)
			panic(err.Error())
		}
		defer insert.Close()
		t, err := template.ParseFiles("templates/register-success.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["registered"] = true
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "register-success", nil)
	}
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.Values["registered"] != nil {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "login", nil)
	} else if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		password := r.FormValue("password")

		row := db.QueryRow("SELECT * FROM users WHERE id = ? AND password = ?", id, password)

		var userid uint
		var name string
		var surname string
		var userEmail string
		var userPassword string
		err = row.Scan(&userid, &name, &surname, &userPassword, &userEmail)
		if err == sql.ErrNoRows {
			fmt.Println("No user found with the given email and password.")
		} else if err != nil {
			panic(err)
		} else {
			t, err := template.ParseFiles("templates/login-success.html", "templates/header.html", "templates/footer.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t.ExecuteTemplate(w, "login-success", nil)
			return
		}

		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		name := r.FormValue("search")
		rows, err := db.Query("SELECT * FROM items WHERE item_name LIKE ?", "%"+name+"%")
		if err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		items := make([]Item, 0)

		for rows.Next() {
			var item Item
			err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage)
			if err != nil {
				log.Fatal(err)
			}
			items = append(items, item)
		}

		tmpl := template.Must(template.ParseFiles("templates/catalog.html", "templates/header.html", "templates/footer.html"))
		err = tmpl.ExecuteTemplate(w, "catalog", items)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		rows, err := db.Query("SELECT * FROM items")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		items := make([]Item, 0)

		for rows.Next() {
			var item Item
			err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage)
			if err != nil {
				log.Fatal(err)
			}
			items = append(items, item)
		}

		tmpl := template.Must(template.ParseFiles("templates/catalog.html", "templates/header.html", "templates/footer.html"))
		err = tmpl.ExecuteTemplate(w, "catalog", items)
		if err != nil {
			log.Fatal(err)
		}

	}

}

func generateSessionKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(key)
}
