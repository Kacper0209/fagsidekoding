package main

import (
	"log"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
)

type user struct {
	UserName  string
	Password  []byte
	First     string
	Last      string
}

var tpl *template.Template
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	// bs, _ := bcrypt.GenerateFromPassword([]byte("password",bcrypt.MinCost))
	// dbUsers["test@test.com"] = user{"test@test.com", bs, "James", "Bond"}
}

func main() {
	http.HandleFunc("/",index)
	http.HandleFunc("/blogg",blogg)
	http.HandleFunc("/signup",signup)
	http.HandleFunc("/login",login)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
} 

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func blogg(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "blogg.gohtml", u)
}

func signup(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	var u user


	if req.Method == http.MethodPost {


		un := req.FormValue("username")	
		p := req.FormValue("password")	
		f := req.FormValue("firstname")	
		l := req.FormValue("lastname")


		if _, ok := dbUsers[un]; ok {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return 
		}


		sID, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}
		c := &http.Cookie {
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un


		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		u = user{un, bs, f, l}
		dbUsers[un] = u


		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", u)
}

func login(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req,"/", http.StatusSeeOther)
		return
	}

	var u user

	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		u, ok := dbUsers[un]
		if !ok {
			http.Redirect(w, req, "/login?status=loginfail", http.StatusForbidden)
			return
		}

		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))

		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}

		sID, err := uuid.NewV4()
		if err != nil {
			http.Error(w, "Interal server error", http.StatusInternalServerError)
		}
		c := &http.Cookie {
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/", http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(w, "login.gohtml", u)
}

// func getUser(w http.ResponseWriter, req *http.Request) user {
// 	c, err := req.Cookie("session")
// 	if err != nil{
// 		sID, err := uuid.NewV4()
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		c = &http.Cookie{
// 			Name: "session",
// 			Value: sID.String(),
// 		}
// 	}
// 	http.SetCookie(w, c)

// 	var u user
// 	if un, ok := dbSessions[c.Value]; ok {
// 		u = dbUsers[un]
// 	}

// 	return u
// }

// func alreadyLoggedIn(req *http.Request) bool {

// 	c, err := req.Cookie("session")
// 	if err != nil {

// 		return false
// 	}
// 	un := dbSessions[c.Value]
// 	_, ok := dbUsers[un]
	
// 	return ok
// }
