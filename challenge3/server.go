package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var database = "BucketHackathon"
var user = "test"
var password = "test"

type item struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
	Shared string `json:"shared"`
}
type bucket struct {
	Title string `json:"title"`
	Items []item `json:"items"`
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("password")

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		fmt.Fprintf(w, "error")
		return
	}
	var name string
	defer db.Close()
	row1 := db.QueryRow("SELECT uname FROM user WHERE uname=?", username)
	e1 := row1.Scan(&name)
	if e1 == nil {
		fmt.Println("registered already", username)
		fmt.Fprintf(w, "registered already")
		return
	}
	_, e := db.Exec("insert into user values(?,?)", username, pass)
	if e != nil {
		log.Println(e)
		fmt.Fprintf(w, "error")
		return
	}
	fmt.Fprintf(w, "success")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("password")
	isValid := isLoginValid(username, pass)
	if isValid {
		fmt.Println("logged in", username)
		fmt.Fprintf(w, "Logged in")
	} else {
		fmt.Println(" Not logged in", username)
		fmt.Fprintf(w, "INVALID")
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	title := r.FormValue("title")
	msg := r.FormValue("msg")

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		fmt.Fprintf(w, "error")
		return
	}
	defer db.Close()
	if username != "" {
		_, e := db.Exec("insert into userTitle values(?,?,'')", username, title)
		if e != nil {
			log.Println(e)
			fmt.Fprintf(w, "error")
			return
		}
		fmt.Fprintf(w, "success")
		return
	}

	_, e := db.Exec("insert into bucket values(?,?,?,?)", title, msg, "false", "false")
	if e != nil {
		log.Println(e)
		fmt.Fprintf(w, "error")
		return
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()

	var title string

	list := make([]string, 0, 100)
	rows, errs := db.Query("select title from userTitle where uname = ?", username)
	if errs != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&title)
		if err != nil {
			log.Fatal(err)
		}

		list = append(list, title)
	}
	bjson, _ := json.Marshal(list)
	fmt.Fprintf(w, string(bjson))
	//fmt.Fprintf(w, getSeq(username))
}

func changeHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	title := r.FormValue("title")
	msg := r.FormValue("msg")
	change := r.FormValue("change")
	to := r.FormValue("to")

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()
	_, _ = db.Exec("UPDATE userTitle SET url = ? WHERE title = ?", username+"007"+title)

	if change == "shareAll" {
		_, e := db.Exec("UPDATE bucket SET shared = 'true' WHERE title = ?", title)

		if e != nil {
			fmt.Fprintf(w, "error")
			return
		}
		fmt.Fprintf(w, "shared")
		return
	}
	_, e := db.Exec("UPDATE bucket SET "+change+" = ? WHERE title = ? and msg = ?", to, title, msg)

	if e != nil {
		fmt.Fprintf(w, "error")
		return
	}
	fmt.Fprintf(w, "shared the selected item")
}
func getHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	title := r.FormValue("title")

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()
	listItems := make([]item, 0, 100)
	var msg, status, shared string
	rows, errs := db.Query("select msg, status, shared from bucket where title = ?", title)
	if errs != nil {
		if errs == sql.ErrNoRows {
			fmt.Fprintf(w, "nil")
			fmt.Println(username, "--> no records")
			return
		}
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&msg, &status, &shared)
		if err != nil {
			log.Fatal(err)
		}
		obj := &item{Msg: msg, Status: status, Shared: shared}
		listItems = append(listItems, *obj)
	}
	bucketobj := &bucket{Title: title, Items: listItems}
	bjson, _ := json.Marshal(bucketobj)
	fmt.Fprintf(w, string(bjson))
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	a := strings.Split(url, "007")
	username := a[0]
	title := a[1]
	fmt.Println("Decoded", username+" & "+title)
	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()
	listItems := make([]item, 0, 100)
	var msg, status, shared string
	rows, errs := db.Query("select msg, status, shared from bucket where title = ? and shared = 'true'", title)
	if errs != nil {
		if errs == sql.ErrNoRows {
			fmt.Print("nil")
			return
		}
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&msg, &status, &shared)
		if err != nil {
			log.Fatal(err)
		}
		obj := &item{Msg: msg, Status: status, Shared: shared}
		listItems = append(listItems, *obj)
	}
	bucketobj := &bucket{Title: title, Items: listItems}
	bjson, _ := json.Marshal(bucketobj)
	fmt.Fprintf(w, string(bjson))
}

/*
func seqHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	fmt.Fprintf(w, getSeq(username))
}

func susHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	level := r.FormValue("level")
	timePts := r.FormValue("pts")
	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()
	_, e := db.Exec("UPDATE user SET level = ? , time = ? WHERE uname = ?", level, timePts, username)

	if e != nil {
		fmt.Fprintf(w, "some error occured")
		return
	}
	fmt.Fprintf(w, getSeq(username))
}
func lbHandler(w http.ResponseWriter, r *http.Request) {

		db, err := sql.Open("mysql", user+":"+password+"@/"+database)
		if err = db.Ping(); err != nil {
			log.Print(err)
			return
		}
		defer db.Close()

		var uname, level, time string

		lb := make([]leader, 0, 100)
		rows, errs := db.Query("select uname, level , time from user order by level DESC, time ASC")
		if errs != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&uname, &level, &time)
			if err != nil {
				log.Fatal(err)
			}

			obj := &leader{UName: uname, Level: level, Points: time}
			lb = append(lb, *obj)
		}
		bjson, _ := json.Marshal(lb)
		fmt.Fprintf(w, string(bjson))

}
*/
/////////////////////////////////////////////

func isLoginValid(username string, pass string) bool {
	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return false
	}
	var name string
	defer db.Close()
	row := db.QueryRow("SELECT uname FROM user WHERE password=? AND uname=?", pass, username)
	e := row.Scan(&name)
	if e != nil {
		log.Println(e)
		return false
	}
	if name == username {
		return true
	}
	return false
}

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/reg", regHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/change", changeHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/url", urlHandler)

	log.Fatal(http.ListenAndServe(":8082", nil))
}
