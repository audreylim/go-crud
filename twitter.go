package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"net/http"
	"os"
)

//User table without Posts
type User struct {
	Userid    	int64
	Username  	string
	Password 	string
	Posts 		[]*Post
}

//Posts table
type Post struct {
	Tweetid 	string
	Username    string
	Status 		string
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

//db
func ReadStatus() (res [][]string) { //make Read all data //Read with current user
	err := db.Ping()
		checkError(err)
	rows, err := db.Query("select id, tweet, username from posts where username = ? order by id DESC", currentuser)
		checkError(err)
	defer rows.Close()
	var tweet, id, username string
	for rows.Next() {
		err := rows.Scan(&id, &tweet, &username)
		checkError(err)
		var a []string
		a = append(a, id, tweet, username)
		res = append(res, a)
	}
	return
}

func ReadSingleStatus() (res string) {
	err := db.Ping()
		checkError(err)
	rows, err := db.Query("select tweet from posts where id = ?", statusid)
		checkError(err)
	defer rows.Close()
	var status string
	for rows.Next() {
		err := rows.Scan(&status)
		checkError(err)
		res = status
	}
	return
}

func ReadStatusId() (res int) {
	err := db.Ping()
		checkError(err)
	rows, err := db.Query("Select id from posts where tweet = ?", AddTweet)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var status string
	for rows.Next() {
		err := rows.Scan(&status)
			checkError(err)
	}
	return
}

func InsertData() {
	stmt, err := db.Prepare("INSERT INTO users(id, username, password) VALUES(?, ?, ?)")
		checkError(err)
	res, err := stmt.Exec(AddUser.Userid, AddUser.Username, AddUser.Password)
		checkError(err)
	lastId, err := res.LastInsertId()
		checkError(err)
	rowCnt, err := res.RowsAffected()
		checkError(err)
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func InsertTweetData() {
	stmt, err := db.Prepare("INSERT INTO posts(tweet, username) VALUES(?, ?)")
		checkError(err)
	res, err := stmt.Exec(AddTweet, AddUser.Username)
		checkError(err)
	lastId, err := res.LastInsertId()
		checkError(err)
	rowCnt, err := res.RowsAffected()
		checkError(err)
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func DeleteTweet() {
	stmt, err := db.Prepare("DELETE FROM posts where id = ?") 
		checkError(err)
	a, err := stmt.Exec(postvalue)
		checkError(err)
	lastId, err := a.LastInsertId()
		checkError(err)
	rowCnt, err := a.RowsAffected()
		checkError(err)
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func EditTweet() {
	stmt, err := db.Prepare("UPDATE posts set tweet = ? where id = ?")
	a, err := stmt.Exec(settweet, statusid)
		checkError(err)
	lastId, err := a.LastInsertId()
		checkError(err)
	rowCnt, err := a.RowsAffected()
		checkError(err)
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}

func AuthUser() (res string) {
	err := db.Ping()
		checkError(err)
	rows, err := db.Query("select username from users where username = ?", logname)
		checkError(err)
	defer rows.Close()
	var logusername string
	for rows.Next() {
		err := rows.Scan(&logusername)
		checkError(err)
		res = logusername
	}
	return 
}

func AuthPw() (res string) {
	err := db.Ping()
		checkError(err)
	rows, err := db.Query("select password from users where username = ?", logname)
		checkError(err)
	defer rows.Close()
	var logpw string
	for rows.Next() {
		err := rows.Scan(&logpw)
			checkError(err)
		res = logpw
	}
	return 
}

//handlers
func loghandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "log", "log", User{})
}

func loginhandler(w http.ResponseWriter, r *http.Request) {
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

func logouthandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
}

func homehandler(w http.ResponseWriter, r *http.Request) {
	currentuser = getUserName(r)
	if currentuser != "" {
		var as []Post 
		a := User{}
		m := ReadStatus()
		for i:=0;i<len(m);i++{
		as = append(as, Post{Tweetid: m[i][0], Username: currentuser, Status: m[i][1]})
	    }
	    a = User{Username: currentuser}
	    for i:=0;i<len(m);i++ {
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
	/*var as []Post 
	a := User{}
	m := ReadStatus()
	for i:=0;i<len(m);i++{
		as = append(as, Post{Tweetid: m[i][0], Username: currentuser, Status: m[i][1]})
	    }
	    a = User{Username: currentuser}
	    for i:=0;i<len(m);i++{
	    	a.Posts = append(a.Posts, &as[i])
	    }
		renderTemplate(w, "home", "homepage", a)*/
		http.Redirect(w, r, "/home", 302)
}

func deletehandler(w http.ResponseWriter, r *http.Request) {
	currentuser = getUserName(r)
	postvalue = r.PostFormValue("xdel")
	DeleteTweet() 
	var as []Post 
	a := User{}
	m := ReadStatus()
	for i:=0;i<len(m);i++{
		as = append(as, Post{Tweetid: m[i][0], Username: currentuser, Status: m[i][1]})
	    }
	    a = User{Username: currentuser}
	for i:=0;i<len(m);i++{
	    a.Posts = append(a.Posts, &as[i])
	}
	renderTemplate(w, "home", "homepage", a)
}

func edithandler(w http.ResponseWriter, r *http.Request) {
	statusid = r.PostFormValue("xedit")
	as := Post{Tweetid: statusid, Status: ReadSingleStatus()}
	person := User{Posts: []*Post{&as}}
	renderTemplate(w, "edit", "editpage", person)//use URL Path method global
}

func savehandler(w http.ResponseWriter, r *http.Request) {
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
		panic(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error", err.Error())
		os.Exit(1)
	}
}

var router = mux.NewRouter()

func main() {
	db, err = sql.Open("mysql", "b7ce733b97afad:414aa83f@tcp(us-cdbr-iron-east-01.cleardb.net:3306)/heroku_31467bc306ebc54")
	checkError(err)
	defer db.Close()
	router.HandleFunc("/", loghandler)
	router.HandleFunc("/login", loginhandler)
	router.HandleFunc("/home", homehandler)
	router.HandleFunc("/home/tweets", usertweetHandler).Methods("POST")
	router.HandleFunc("/logout", logouthandler).Methods("POST")
	router.HandleFunc("/home/delete", deletehandler).Methods("POST")
	router.HandleFunc("/home/edit", edithandler).Methods("POST")
	router.HandleFunc("/home/save", savehandler).Methods("POST")

    router.PathPrefix("/").Handler(http.FileServer(http.Dir("./layout/")))

	http.Handle("/", router)
	err := http.ListenAndServe(":"+os.Getenv("PORT") , nil)
	if err != nil {
		panic(err)
	}

}
