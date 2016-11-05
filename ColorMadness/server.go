package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var database = "hackathon"
var user = "test"
var password = "test"

type leader struct {
	UName  string `json:"uname"`
	Level  string `json:"level"`
	Points string `json:"points"`
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
	_, e := db.Exec("insert into user values(?,?,'0','N/A')", username, pass)
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
		getLevel(username)
		fmt.Fprintf(w, strconv.Itoa(getLevel(username)))
	} else {
		fmt.Println(" Not logged in", username)
		fmt.Fprintf(w, "INVALID")
	}
}

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
func getLevel(username string) int {
	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return 0
	}
	var level string
	defer db.Close()
	row := db.QueryRow("SELECT level FROM user WHERE uname=?", username)
	e := row.Scan(&level)
	if e != nil {
		log.Println(e)
		return 0
	}
	if l, err := strconv.Atoi(level); err == nil {
		return l
	}
	return 0
}
func getSeq(username string) string {
	l := getLevel(username)
	seq := make([]int, 0, 100)
	i := 0
	var ran1, ran int
	ran1 = -1
	for i < l+1 {
		ran = rand.Intn(9)
		if ran1 == ran {
			continue
		}
		ran1 = ran
		seq = append(seq, ran)
		i = i + 1
	}
	bjson, _ := json.Marshal(seq)
	return string(bjson)

}
func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/reg", regHandler)
	http.HandleFunc("/seq", seqHandler)
	http.HandleFunc("/sus", susHandler)
	http.HandleFunc("/lb", lbHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
