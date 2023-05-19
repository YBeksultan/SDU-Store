package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

type User struct {
	Id, isAdmin                        int
	Username, Surname, Password, Email string
}

type Item struct {
	ItemId     int
	ItemName   string
	ItemPrice  float64
	ItemImage  string
	ItemRating float64
	ItemRated  int
}

type Comment struct {
	CommentId     uint16
	ItemId        uint16
	CommentAuthor string
	CommentText   string
	CommentDate   string
}

type Rating struct {
	ItemId uint16
	Rating float64
	author string
	Count  int
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
	rtr.HandleFunc("/submit", submitHandler)
	rtr.HandleFunc("/cart", cartHandler)
	rtr.HandleFunc("/add_to_cart", addToCart)
	rtr.HandleFunc("/add_item", addItem)
	rtr.HandleFunc("/remove_product", removeProduct)
	rtr.HandleFunc("/order", orderHandler)

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
		var user User = User{id, 0, name, surname, password, email}
		query := "INSERT INTO users VALUES ('" + strconv.Itoa(user.Id) + "','" + user.Username +
			"','" + user.Surname + "','" + user.Password + "','" + user.Email + "', 0);"
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
		session.Values["id"] = id
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
		var isAdmin int
		err = row.Scan(&userid, &name, &surname, &userPassword, &userEmail, &isAdmin)
		if isAdmin == 1 {
			session.Values["admin"] = true
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			session.Values["admin"] = false
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
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
			session.Values["id"] = id
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
		priceBy := r.FormValue("priceBy")
		category := r.FormValue("category")
		if category != "" {
			if category == "All" {
				query := "SELECT * FROM items"
				if priceBy == "asc" {
					query += " ORDER BY item_price ASC"
				} else if priceBy == "desc" {
					query += " ORDER BY item_price DESC"
				} else if priceBy == "rateAsc" {
					query += " ORDER BY rating ASC"
				} else if priceBy == "rateDesc" {
					query += " ORDER BY rating DESC"
				}
				rows, err := db.Query(query)
				if err != nil {
					log.Fatal(err)
				}
				defer rows.Close()

				items := make([]Item, 0)

				for rows.Next() {
					var item Item
					err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
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
				query := "SELECT * FROM items WHERE item_name LIKE '%" + category + "%'"
				if priceBy == "rateAsc" {
					query += " ORDER BY rating ASC"
				} else if priceBy == "rateDesc" {
					query += " ORDER BY rating DESC"
				}
				rows, err := db.Query(query)
				if err != nil {
					log.Fatal(err)
				}
				defer rows.Close()

				items := make([]Item, 0)

				for rows.Next() {
					var item Item
					err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
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
			query := "SELECT * FROM items WHERE item_name LIKE %" + name + "%"
			if priceBy == "rateAsc" {
				query += " ORDER BY rating ASC"
			} else if priceBy == "rateDesc" {
				query += " ORDER BY rating DESC"
			}
			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			items := make([]Item, 0)

			for rows.Next() {
				var item Item
				err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
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
			err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
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
		err := res.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		m["items"] = item
	}
	com, err := db.Query(fmt.Sprintf("SELECT * FROM `comments` WHERE `item_id` = '%s' ORDER BY comment_date DESC", vars["id"]))
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

	ratings, err := db.Query(fmt.Sprintf("SELECT * FROM `ratings` WHERE `item_id` = '%s'", vars["id"]))
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	n := 0
	sum := 0.0
	for ratings.Next() {
		var rating Rating
		err := ratings.Scan(&rating.ItemId, &rating.Rating, &rating.author)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
		sum += rating.Rating
		n++
	}
	var overall float64

	if n != 0 {
		overall = sum / float64(n)
	} else {
		overall = 0.0
	}
	var ratingOfItem Rating
	i, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	result := uint16(i)
	ratingOfItem.ItemId = result
	rate, err := strconv.ParseFloat(strconv.FormatFloat(overall, 'f', 1, 64), 64)
	ratingOfItem.Rating = rate
	ratingOfItem.Count = n

	m["rating"] = ratingOfItem
	if n != 0 {
		query := "UPDATE items SET rating=" + strconv.FormatFloat(overall, 'f', 1, 64) + ", rated = " + strconv.Itoa(n) + " WHERE item_id=" + strconv.Itoa(id) + ";"
		update, err := db.Query(fmt.Sprintf(query))
		if err != nil {
			panic(err)
		}
		defer update.Close()
	}

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

func submitHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	r.ParseForm()
	stars := r.Form.Get("rating")

	row := db.QueryRow("SELECT * FROM ratings WHERE author = ? and item_id = ?", session.Values["id"], id)

	var itemid uint
	var rating string
	var author string
	var author_id = (session.Values["id"]).(int)
	err = row.Scan(&itemid, &rating, &author)
	if err == sql.ErrNoRows {
		query := "INSERT INTO `ratings` (`item_id`, `rating`, `author`) VALUES ('" + strconv.Itoa(id) + "', '" + stars + "', '" + strconv.Itoa(author_id) + "')"
		insert, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		defer insert.Close()
		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)
	} else {
		query := "UPDATE ratings SET rating=" + stars + " WHERE author='" + strconv.Itoa(author_id) + "';"
		update, err := db.Query(fmt.Sprintf(query))
		if err != nil {
			panic(err)
		}
		defer update.Close()
		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)
	}

}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	logged, _ := session.Values["loggedIn"]
	if logged == false {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	var author_id = (session.Values["id"]).(int)

	query := "SELECT * FROM cart WHERE cart_author = " + strconv.Itoa(author_id)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	items := make([]Item, 0)

	for rows.Next() {
		var item Item
		var author string
		err := rows.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &author)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	//
	//if len(items) == 0 {
	//
	//}
	//else {
	//
	//}
	t, err := template.ParseFiles("templates/cart.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "cart", items)
}

func addToCart(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	row := db.QueryRow("SELECT * FROM cart WHERE cart_author = ? and item_id = ?", session.Values["id"], id)

	var itemId int
	var itemName string
	var itemPrice int
	var itemImage string
	var cartAuthor int
	err = row.Scan(&itemId, &itemName, &itemPrice, &itemImage, &cartAuthor)

	if err == sql.ErrNoRows {
		row2 := db.QueryRow("SELECT * from items WHERE item_id = ?", id)

		var item Item
		err := row2.Scan(&item.ItemId, &item.ItemName, &item.ItemPrice, &item.ItemImage, &item.ItemRating, &item.ItemRated)
		var author_id = (session.Values["id"]).(int)

		query := "INSERT INTO cart VALUES (" + strconv.Itoa(id) + ", '" + item.ItemName + "', " + strconv.FormatFloat(item.ItemPrice, 'f', -1, 64) + ", '" + item.ItemImage + "', " + strconv.Itoa(author_id) + ")"
		insert, err := db.Query(query)

		if err != nil {
			panic(err)
		}
		defer insert.Close()
		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)

	}
}

func addItem(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	admin, _ := session.Values["admin"]
	if admin == false {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/addItem.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		t.ExecuteTemplate(w, "add_item", nil)
	} else if r.Method == "POST" {
		item_name := r.FormValue("item_name")
		price, _ := strconv.Atoi(r.FormValue("price"))
		image := r.FormValue("image")
		query := "INSERT INTO items(item_name, item_price, item_image, rating, rated) VALUES ('" + item_name + "'," + strconv.Itoa(price) +
			",'" + image + "',0.0 ,0);"
		insert, err := db.Query(query)
		if err != nil {
			fmt.Println(query)
			panic(err.Error())
		}
		defer insert.Close()
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
	}

}
func removeProduct(w http.ResponseWriter, r *http.Request) {
	itemID := r.FormValue("item_id")
	fmt.Println(itemID)
	query := "DELETE from cart WHERE item_id = " + itemID
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}
func orderHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user-session")
	var author_id = (session.Values["id"]).(int)
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/order.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		t.ExecuteTemplate(w, "order", nil)
	} else if r.Method == "POST" {
		email := r.FormValue("email")
		sendEmail(email)
		query := "DELETE from cart WHERE cart_author = " + strconv.Itoa(author_id)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
	}
}
func sendEmail(email string) {
	auth := smtp.PlainAuth("", "sdutalkreset@gmail.com", "mfvinfuohpjeuhxi", "smtp.gmail.com")
	randNum := generateRandomNumber()
	err := smtp.SendMail("smtp.gmail.com:587", auth, "sdutalkreset@gmail.com", []string{email}, []byte("Subject: SDU Store\nThank You for purchase! \nYour order id is: "+strconv.Itoa(randNum)+"\nCome again, we will be glad to see you in our store!"))
	if err != nil {
		fmt.Println(err)
	}
}

func generateRandomNumber() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(901) + 100
}
