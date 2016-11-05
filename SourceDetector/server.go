package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var database = "hackathon"
var user = "test"
var password = "test"

func regHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("password")

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		fmt.Fprintf(w, "error")
		return
	}

	defer db.Close()
	isValid := isLoginValid(username, pass)
	if isValid {
		fmt.Println("registered already", username)
		fmt.Fprintf(w, "registered already")
		return
	}
	_, e := db.Exec("insert into user values(?,?,'0')", username, pass)
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

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err = db.Ping(); err != nil {
		log.Print(err)
		return
	}
	defer db.Close()
	_, e := db.Exec("UPDATE user SET level = ? WHERE uname = ?", level, username)

	if e != nil {
		fmt.Fprintf(w, "some error occured")
		return
	}
	fmt.Fprintf(w, getSeq(username))
}

var javaFile string = ""
var jFlag bool = false

func codeHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	name := r.FormValue("name")
	path := name // linux

	err := ioutil.WriteFile(path+".c", []byte(code), 0777)
	err = ioutil.WriteFile(path+".cpp", []byte(code), 0777)
	err = ioutil.WriteFile(path+".rb", []byte(code), 0777)
	err = ioutil.WriteFile(path+".py", []byte(code), 0777)

	if err != nil {
		log.Print(err)
		fmt.Fprintf(w, "error")
		return
	}

	a := fun1("/usr/bin/gcc", path+".c")
	if a == 1 {
		fmt.Fprintf(w, "C")
		return
	}
	a = fun1("/usr/bin/g++", path+".cpp")
	if a == 1 {
		fmt.Fprintf(w, "C++")
		return
	}
	a = fun1("/usr/bin/python", path+".py")
	if a == 1 {
		fmt.Fprintf(w, "python")
		return
	}
	a = fun1("/usr/bin/ruby", path+".rb")
	if a == 1 {
		fmt.Fprintf(w, "ruby")
		return
	}
	err = ioutil.WriteFile(path+".java", []byte(code), 0777)
	jFlag = false
	_ = renameFile(path + ".java")
	if jFlag {
		path = javaFile
		fmt.Println("Applying file name to java-->", path)
	}
	err = ioutil.WriteFile(path+".java", []byte(code), 0777)
	a = fun1("/usr/bin/javac", path+".java")
	if a == 1 {
		fmt.Fprintf(w, "java")
		return
	}
	//fmt.Println(a,b, c)
}

/////////////////////////////////////////////
func renameFile(s string) string {
	f, err := os.Open(s)
	if err != nil {
		fmt.Println("error opening file= ", err)
		os.Exit(1)
	}
	r := bufio.NewReader(f)
	s, e := Readln(r)
	for e == nil {
		s, e = Readln(r)
		if s != "" {
			k := strings.TrimSpace(s)
			fmt.Println(k)
			return strings.TrimSpace(s)
		}
	}
	return ""
}
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	//fmt.Println("-->",string(ln))
	s := string(ln)
	if strings.Contains(s, "public class") {
		fmt.Println("Found public class")
		s = strings.Replace(s, "public class", "", -1)
		s = strings.Replace(s, "{", "", -1)
		//fmt.Println(s)
		fmt.Println("java file name found to be --> ", s)
		javaFile = strings.TrimSpace(s)
		jFlag = true
		return s, err
	}
	return "", err
}
func fun(p string, f string) int {
	cmd := exec.Command("usr/bin/bash echo", f)
	var out1 bytes.Buffer
	cmd.Stdout = &out1
	err1 := cmd.Run()
	if err1 != nil {
		//log.Fatal(err1)
	}
	fmt.Println(p, out1.String())
	return len(out1.String())
}
func fun1(p string, f string) int {
	cmd, err1 := exec.Command(p, f).Output()

	if err1 != nil {
		log.Print(err1)
		return 0
	}
	fmt.Println(p, string(cmd))
	return 1
	//return len(out1.String())
}

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
	http.HandleFunc("/code", codeHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
