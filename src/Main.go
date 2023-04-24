package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"
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

type Comment struct {
	CommentId     uint16
	ItemId        uint16
	CommentAuthor string
	CommentText   string
	CommentDate   string
}

var id int

var (
	db    *sql.DB
	err   error
	store = sessions.NewCookieStore([]byte(os.Getenv(generateSessionKey())))
)

func main() {
	store = sessions.NewCookieStore([]byte(os.Getenv(generateSessionKey())))
	db, err = sql.Open("mysql", "root:9Wyk%L7nUvm4@/sdu_store")
	if err != nil {
		panic(err.Error())
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/home", homeHandler)
	rtr.HandleFunc("/register", registerHandler)
	rtr.HandleFunc("/login", loginHandler)
	rtr.HandleFunc("/catalog", catalogHandler)
	rtr.HandleFunc("/about", aboutHandler)
	rtr.HandleFunc("/contact", contactHandler)
	rtr.HandleFunc("/logout", logout)
	rtr.HandleFunc("/save_comment", addComment)
	rtr.HandleFunc("/product/{id:[0-9]+}", productHandler)

	http.Handle("/", rtr)
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
	session, _ := store.Get(r, "user-session")
	logged, _ := session.Values["loggedIn"]
	if logged == true {
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
		session.Values["username"] = name
		session.Values["surname"] = surname
		session.Values["loggedIn"] = true
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logg, _ := store.Get(r, "logged")
		err = logg.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "register-success", nil)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")

	logged, _ := session.Values["loggedIn"]
	if logged == true {
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
			session.Values["username"] = name
			session.Values["surname"] = surname
			session.Values["loggedIn"] = true
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logg, _ := store.Get(r, "logged")
			err = logg.Save(r, w)
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

func logout(w http.ResponseWriter, r *http.Request) {
	logg, err := store.Get(r, "logged")
	if err != nil {
	}
	logg.Options.MaxAge = -1
	err = logg.Save(r, w)
	if err != nil {
	}
	session, _ := store.Get(r, "user-session")
	session.Values["loggedIn"] = false
	err = session.Save(r, w)
	if err != nil {
	}
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	logged, _ := session.Values["loggedIn"]
	if logged == false {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == "POST" {
		priceBy := r.FormValue("priceby")
		category := r.FormValue("category")
		if category != "" {
			if category == "All" {
				query := "SELECT * FROM items"
				if priceBy == "asc" {
					query += " ORDER BY item_price ASC"
				} else if priceBy == "desc" {
					query += " ORDER BY item_price DESC"
				} else if priceBy == "rating_asc" {
					query += " ORDER BY rating ASC"
				} else if priceBy == "rating_desc" {
					query += " ORDER BY rating DESC"
				}
				fmt.Println("Query" + query + " Priceby:" + priceBy)
				rows, err := db.Query(query)
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
				rows, err := db.Query("SELECT * FROM items WHERE item_name LIKE ?", "%"+category+"%")
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
		} else {
			name := r.FormValue("search")
			rows, err := db.Query("SELECT * FROM items WHERE item_name LIKE ?", "%"+name+"%")
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

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/about.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "about", nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/contact.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "contact", nil)
}

func generateSessionKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(key)
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/product.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	res, err := db.Query(fmt.Sprintf("SELECT * FROM `items` WHERE `item_id` = '%s'", vars["id"]))
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	m := map[string]interface{}{}

	for res.Next() {
		var item Item
		err := res.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		m["items"] = item
	}

	com, err := db.Query(fmt.Sprintf("SELECT * FROM `comments` WHERE `item_id` = '%s'", vars["id"]))
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	id, _ = strconv.Atoi(vars["id"])

	comments := make([]Comment, 0)

	for com.Next() {
		var comment Comment
		err := com.Scan(&comment.CommentId, &comment.ItemId, &comment.CommentAuthor, &comment.CommentText, &comment.CommentDate)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		comments = append(comments, comment)
	}
	m["comments"] = comments
	t.ExecuteTemplate(w, "product", m)
}

func addComment(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	comment := r.FormValue("comment_text")

	date := time.Now().Format("2006-01-02 15:04 Monday")

	if comment == "" {
		fmt.Fprintf(w, "Please, write something")
	} else {
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `comments` (`comment_text`, `item_id`, `comment_author`, `comment_date`) VALUES ('%s', '%d', '%s', '%s')", comment, id, session.Values["username"], date))
		if err != nil {
			panic(err)
		}
		defer insert.Close()

		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)
	}
}

func init() {
	_ = mime.AddExtensionType(".js", "text/javascript")
}
