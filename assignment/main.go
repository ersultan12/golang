package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	_"github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Ersers123"
	dbname   = "booksford"
)


func getDb() (*sql.DB, error) {
	psqlInfo := "host=localhost port=5432 user=postgres password=Ersers123 dbname=booksford sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	return db, err
}

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type User struct {
	Id                 int
	Username, Password string
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
	}
	t.ExecuteTemplate(w, "index", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/template.html")
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
	}
	t.ExecuteTemplate(w, "register", nil)
}

func save_user(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		fmt.Fprintf(w, "not all fields filled")
	} else {
		db, err := getDb()
		if err != nil {
			panic(err)
		}
		res, err := db.Query("select * from users;")
		if err != nil {
			panic(err)
		}

		for res.Next() {
			var user User
			err = res.Scan(&user.Id, &user.Username, &user.Password)
			if err != nil {
				panic(err)
			}
			if user.Username == username {
				fmt.Fprintf(w, "username already exists")
				return
			}
		}
		insert, err := db.Query(fmt.Sprintf("INSERT INTO users (username, password) VALUES('%s', '%s')", username, password))
		if err != nil {
			panic(err)
		}
		defer db.Close()
		defer res.Close()
		defer insert.Close()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
	}
	t.ExecuteTemplate(w, "login", nil)
}

func check_login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		fmt.Fprintf(w, "not all fields filled")
	} else {
		db, err := getDb()
		if err != nil {
			fmt.Fprint(w, err.Error())
		}

		res, err := db.Query("select * from users;")
		if err != nil {
			panic(err)
		}

		ok := false
		for res.Next() {
			var user User
			err = res.Scan(&user.Id, &user.Username, &user.Password)
			if err != nil {
				panic(err)
			}
			if user.Username == username {
				if user.Password == password {
					ok = true
				} else {
					fmt.Fprintf(w, "wrong password")
				}
			}
		}

		if !ok {
			fmt.Fprintf(w, "user not found")
		}

		defer db.Close()
		defer res.Close()
		http.Redirect(w, r, "books", http.StatusSeeOther)
	}
}

type Book struct {
	Id, Year      int
	Title, Author string
}

func books(w http.ResponseWriter, r *http.Request) {
	var books = []Book{}
	t, err := template.ParseFiles("templates/books.html")
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
	}
	db, err := getDb()
	if err != nil {
		panic(err)
	}
	res, err := db.Query("select * from books;")
	if err != nil {
		panic(err)
	}

	for res.Next() {
		var book Book
		err = res.Scan(&book.Id, &book.Title, &book.Year, &book.Author)
		if err != nil {
			panic(err)
		}
		books = append(books, book)
	}
	defer db.Close()
	defer res.Close()
	t.ExecuteTemplate(w, "books", books)
}

func addbook(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/addbook.html")
	if err != nil {
		fmt.Fprintf(w, "error %s", err.Error())
	}
	t.ExecuteTemplate(w, "addbook", nil)
}

func save_book(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	author := r.FormValue("author")
	year := r.FormValue("year")

	if title == "" || author == "" || year == "" {
		fmt.Fprintf(w, "not all fields filled")
	} else {
		db, err := getDb()
		if err != nil {
			panic(err)
		}

		res, err := db.Query("select * from books;")
		if err != nil {
			panic(err)
		}

		for res.Next() {
			var book Book
			err = res.Scan(&book.Id, &book.Title, &book.Year, &book.Author)
			if err != nil {
				panic(err)
			}
			if book.Title == title {
				fmt.Fprintf(w, "this book already exists")
				return
			}
		}
		defer db.Close()
		defer res.Close()
		insert, err := db.Query(fmt.Sprintf("INSERT INTO books (title, year, author) VALUES('%s', %s, '%s')", title, year, author))
		if err != nil {
			panic(err)
		}

		defer insert.Close()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func search(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("search")

	if title == "" {
		fmt.Fprintf(w, "empty form")
	} else {
		db, err := getDb()
		if err != nil {
			panic(err)
		}
		ok := true
		res, err := db.Query(fmt.Sprintf("select * from books where title = '%s';", title))
		if err != nil {
			ok = false
			panic(err)
		}
		var book Book
		var books = []Book{}
		for res.Next() {
			err = res.Scan(&book.Id, &book.Title, &book.Year, &book.Author)
			if err != nil {
				panic(err)
			}
		}

		t, err := template.ParseFiles("templates/books.html")
		if err != nil {
			panic(err)
		}
		books = append(books, book)
		t.ExecuteTemplate(w, "books", books)

		if !ok {
			fmt.Fprintf(w, "Book not found")
		}
		defer db.Close()
		defer res.Close()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func sortt(w http.ResponseWriter, r *http.Request) {
	db, err := getDb()
	if err != nil {
		panic(err)
	}

	res, err := db.Query("select * from books;")
	if err != nil {
		panic(err)
	}

	var books = []Book{}
	for res.Next() {
		var book Book

		err = res.Scan(&book.Id, &book.Title, &book.Year, &book.Author)
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		books = append(books, book)

	}
	t, err := template.ParseFiles("templates/books.html")
	sort.Slice(books, func(i, j int) bool {
		return books[i].Title < books[j].Title
	})
	t.ExecuteTemplate(w, "books", books)

	defer db.Close()
	defer res.Close()
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func HandleFunc() {
	fs := http.FileServer(http.Dir("./templates/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", index)
	http.HandleFunc("/register", register)
	http.HandleFunc("/save_user", save_user)
	http.HandleFunc("/login", login)
	http.HandleFunc("/check_login", check_login)
	http.HandleFunc("/books", books)
	http.HandleFunc("/addbook", addbook)
	http.HandleFunc("/save_book", save_book)
	http.HandleFunc("/search", search)
	http.HandleFunc("/sort", sortt)
	http.ListenAndServe(":8800", nil)
}

func main() {
	HandleFunc()
}
