package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type User struct {
	Id                             int
	name, surname, password, email string
}

var users []User

func main() {
	start()
}

func start() {
	fmt.Println("Choose a action:\n1) Sign up\n2) Log in")
	var choice int
	_, err := fmt.Scanf("%d", &choice)

	if err != nil {
		log.Fatal(err)
	}

	if choice == 1 {
		register()
	} else if choice == 2 {
		authorize()
	} else {
		start()
	}

}

func register() {
	newUser := User{}

	newUser.Id = generateID()

	fmt.Print("Name: ")
	fmt.Scan(&newUser.name)

	fmt.Print("Surname: ")
	fmt.Scan(&newUser.surname)

	fmt.Print("Password: ")
	fmt.Scan(&newUser.password)

	fmt.Print("Email: ")
	fmt.Scan(&newUser.email)

	users = append(users, newUser)

	fmt.Println()
	fmt.Println("Log In\n")
	authorize()
}

func generateID() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(1000000)
}

func authorize() {
	var loggedUser User
	var inputEmail string
	var pswrd string

	fmt.Printf("Enter your email: ")
	fmt.Scan(&inputEmail)

	for i := 0; i < len(users); i++ {
		if users[i].email == inputEmail {
			fmt.Printf("Enter your password: ")
			fmt.Scan(&pswrd)
			if pswrd == users[i].password {
				loggedUser = users[i]
			} else {
				fmt.Println()
				fmt.Println("Password is incorrect. Try again.")
				fmt.Println()
				start()
			}
		} else {
			fmt.Println()
			fmt.Println("Email is incorrect. Try again.")
			fmt.Println()
			start()
		}
	}
	fmt.Println(loggedUser)
}
