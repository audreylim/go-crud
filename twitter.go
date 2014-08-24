package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"net/http"
	"log"
	"os"
)

//User table
type User struct {
	Userid   int64
	Username string
	Password string
	Posts    []*Post
}

//Posts table
type Post struct {
	Tweetid  string
	Username string
	Status   string
}

//global variables
var AddUser User = User{}
var AddTweet string
var logname string
var logpass string
var currentuser string
var postvalue string
var statusid string
var settweet string
var db *sql.DB
var err error

//db
func ReadStatus() (res [][]string) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT id, tweet, username FROM posts WHERE username = ? ORDER BY id DESC", currentuser)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var tweet, id, username string
	for rows.Next() {
		err := rows.Scan(&id, &tweet, &username)
		if err != nil {
		 return [""][""]
		}
		var a []string
		a = append(a, id, tweet, username)
		res = append(res, a)
	}
	return
}

func ReadSingleStatus() (res string) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT tweet FROM posts WHERE id = ?", statusid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var status string
	for rows.Next() {
		err := rows.Scan(&status)
		if err != nil {
			log.Fatal(err)
		}
		res = status
	}
	return
}

func ReadStatusId() (res int) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT id FROM posts WHERE tweet = ?", AddTweet)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var status string
	for rows.Next() {
		err := rows.Scan(&status)
		if err != nil {
		log.Fatal(err)
		}
	}
	return
}

func InsertData() {
	stmt, err := db.Prepare("INSERT INTO users(id, username, password) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(AddUser.Userid, AddUser.Username, AddUser.Password)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func InsertTweetData() {
	stmt, err := db.Prepare("INSERT INTO posts(tweet, username) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(AddTweet, AddUser.Username)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func DeleteTweet() {
	stmt, err := db.Prepare("DELETE FROM posts where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(postvalue)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func EditTweet() {
	stmt, err := db.Prepare("UPDATE posts SET tweet = ? where id = ?")
	a, err := stmt.Exec(settweet, statusid)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := a.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := a.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func AuthUser() (res string) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT username FROM users WHERE username = ?", logname)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var logusername string
	for rows.Next() {
		err := rows.Scan(&logusername)
		if err != nil {
			return ""
		}
		res = logusername
	}
	return
}

func AuthPw() (res string) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT password FROM users WHERE username = ?", logname)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var logpw string
	for rows.Next() {
		err := rows.Scan(&logpw)
		if err != nil {
			return ""
		}
		res = logpw
	}
	return
}

//handlers
func logHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "log", "log", User{})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	logname = r.FormValue("name")
	logpass = r.FormValue("password")
	AddUser.Password = logpass
	fmt.Println(AuthPw(), logpass)
	redirectTarget := "/"
	switch {
	//existing user correct pw
	case logname != "" && logpass != "" && logname == AuthUser() && logpass == AuthPw():
		setSession(logname, w)
		redirectTarget = "/home"
	//existing user wrong pw
	case logname != "" && logpass != "" && logname == AuthUser() && logpass != AuthPw():
		fmt.Fprintf(w, "<html>wrong password</html>")
	//new user
	case logname != "" && logpass != "" && logname != AuthUser():
		AddUser.Username = logname
		AddUser.Password = logpass
		InsertData()
		setSession(logname, w)
		redirectTarget = "/home"
	}
	http.Redirect(w, r, redirectTarget, 302)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	currentuser = getUserName(r)
	if currentuser != "" {
		var as []Post
		a := User{}
		m := ReadStatus()
		for i := 0; i < len(m); i++ {
			as = append(as, Post{Tweetid: m[i][0], Username: currentuser, Status: m[i][1]})
		}
		a = User{Username: currentuser}
		for i := 0; i < len(m); i++ {
			a.Posts = append(a.Posts, &as[i])
		}
		renderTemplate(w, "home", "homepage", a)
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func usertweetHandler(w http.ResponseWriter, r *http.Request) {
	AddUser.Username = getUserName(r)
	AddTweet = r.FormValue("twt")
	InsertTweetData()
	http.Redirect(w, r, "/home", 302)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	currentuser = getUserName(r)
	postvalue = r.PostFormValue("xdel")
	DeleteTweet()
	http.Redirect(w, r, "/home", 302)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	statusid = r.PostFormValue("xedit")
	as := Post{Tweetid: statusid, Status: ReadSingleStatus()}
	a := User{Posts: []*Post{&as}}
	renderTemplate(w, "edit", "editpage", a)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	statusid = r.PostFormValue("xsave")
	settweet = r.PostFormValue("textedit")
	EditTweet()
	http.Redirect(w, r, "/home", 302)
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(r *http.Request) (currentuser string) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			currentuser = cookieValue["name"]
		}
	}
	return
}

func setSession(userName string, w http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}

func clearSession(r http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(r, cookie)
}

func renderTemplate(w http.ResponseWriter, tmpl string, def string, x User) {
	t := template.Must(template.New("tele").ParseFiles("layout/" + tmpl + ".html"))
	if err := t.ExecuteTemplate(w, def, x); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var router = mux.NewRouter()

func main() {
	db, err = sql.Open("mysql", os.Getenv("DB_USERNAME") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_CLEARDB") + ":3306)/" + os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal(err)
	} 
	defer db.Close()

	router.HandleFunc("/", logHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/home", homeHandler)
	router.HandleFunc("/home/tweets", usertweetHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/home/delete", deleteHandler).Methods("POST")
	router.HandleFunc("/home/edit", editHandler).Methods("POST")
	router.HandleFunc("/home/save", saveHandler).Methods("POST")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./layout/")))

	http.Handle("/", router)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
