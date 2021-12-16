package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Person struct {
	gorm.Model // เก็บ id createdAt updatedAt deletedAt

	Name  string
	Email string `gorm:"typevarchar(100);unique_index"` // 1 คนตัองมี email ได้แค่เมลเดียว(unique) และห้ามเกิน 100 อักษร
	Books []Book
}

type Book struct {
	gorm.Model

	Title      string
	Author     string
	CallNumber int `gorm:"unique_index"` // Id of book for identify books (เก็บเป็น string ก็ได้)
	PersonID   int // relation between Person struct
}

// var (
// 	person = &Person{Name: "Asun", Email: "Asun@email.com"}
// 	books  = []Book{
// 		{Title: "Titanic 2077", Author: "IDK", CallNumber: 123456, PersonID: 1},
// 		{Title: "KIJK", Author: "PPP", CallNumber: 234512, PersonID: 1},
// 		{Title: "ww3 ", Author: "ME", CallNumber: 234213, PersonID: 1},
// 	}
// )

var db *gorm.DB
var err error

func main() {
	//Loading env
	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := os.Getenv("PASSWORD")

	//db connecting
	dbURI := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbName, dbPort)

	//Openong connection to database
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("connect to postgres DB")
	}

	//Close connection to db when main func finished (defer = ทำเมื่อการทำงานจบ)
	defer db.Close() // ทำเพื่อความชัวร์

	//Make migrations to database if they have not already been created ถ้าdatabase ยังไม่มีจะทำการสร้่าง table โดยอัตโนมัติ
	db.AutoMigrate(&Person{})
	db.AutoMigrate(&Book{})

	// API routes
	router := mux.NewRouter()

	router.HandleFunc("/people", getPeople).Methods("GET")
	router.HandleFunc("/person/{id}", getPerson).Methods("GET") //get single person and their books

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/book/{id}", getBook).Methods("GET")

	router.HandleFunc("/create/person", createPerson).Methods("POST")
	router.HandleFunc("/create/book", createBook).Methods("POST")

	router.HandleFunc("/delete/person/{id}", deletePerson).Methods("DELETE")
	router.HandleFunc("/delete/book/{id}", deleteBook).Methods("DELETE")

	http.ListenAndServe(":8080", router)

}

//API controller

//People Controller
func getPeople(w http.ResponseWriter, r *http.Request) {
	var people []Person
	db.Find(&people) //find all of people

	json.NewEncoder(w).Encode(&people)
}

func getPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r) //read request header and get id

	var person Person
	var books []Book

	db.First(&person, params["id"]) //find the first person when query
	db.Model(&person).Related(&books)

	person.Books = books

	json.NewEncoder(w).Encode(&person)
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	var person Person
	json.NewDecoder(r.Body).Decode(&person)

	createdPerson := db.Create(&person)
	err = createdPerson.Error
	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&person)
	}

}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var person Person

	db.First(&person, params["id"])
	db.Delete(&person)

	json.NewEncoder(w).Encode(&person)
}

//Book controller
func getBooks(w http.ResponseWriter, r *http.Request) {
	var books Book

	err := db.Find(&books)

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(&books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var book Book
	err := db.First(&book, params["id"])

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(&book)
}

func createBook(w http.ResponseWriter, r *http.Request) {

	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	createdBook := db.Create(&book)
	err = createdBook.Error
	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&book)
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var book Book
	db.First(&book, params["id"])
	err := db.Delete(&book)

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(&book)
}
