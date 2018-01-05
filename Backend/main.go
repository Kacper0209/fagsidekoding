package main

import (
	"log"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"fmt"
)

type user struct {
	UserName  string
	Password  []byte
	First     string
	Last      string
	Role      string
}

type blog struct {
	Title    string
	By       string
	Content  string
}

var tpl *template.Template
var dbUsers = map[string]user{}
var dbSessions = map[string]string{}
var dbBlog = map[string]blog{}



func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	bs, _ := bcrypt.GenerateFromPassword([]byte("password"),bcrypt.MinCost)
	dbUsers["test@test.com"] = user{"test@test.com", bs, "James", "Bond","admin"}
}

func main() {
	http.HandleFunc("/",index)
	http.HandleFunc("/blogg",blogg)
	http.HandleFunc("/signup",signup)
	http.HandleFunc("/login",login)
	http.HandleFunc("/logout",logout)
	http.HandleFunc("/newblog",nyblogg)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
} 

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func blogg(w http.ResponseWriter, req *http.Request) {
	// blogdata := dbBlog[]
	u := getUser(w, req)
	fmt.Println(dbBlog)
	fmt.Println(u)
	data := dbBlog

	//struct{
	// 	UserName string
	// 	Blog map[string]blog
	// }{
	// 	u.UserName,
	// 	dbBlog,
	// }
	//if !alreadyLoggedIn(req) {
	//	http.Redirect(w, req, "/login", http.StatusSeeOther)
	//	return
	//}
	//if u.Role != "admin" {
	//	http.Redirect(w, req, "/", http.StatusSeeOther)
	//}
	tpl.ExecuteTemplate(w, "blogg.gohtml", data)
}

func nyblogg(w http.ResponseWriter, req *http.Request) {
	//Logget inn og admin(for nå)?
	u := getUser(w, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	if u.Role != "admin" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		titl := req.FormValue("title")
		cont := req.FormValue("content")

		b := blog{titl,u.UserName,cont}

		bID, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}

		dbBlog[bID.String()] = b

		http.Redirect(w, req, "/blogg", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "nyblogg.gohtml", u)
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
		r := req.FormValue("admin")

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


		bs, err := bcrypt.GenerateFromPassword([]byte(p),bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		u = user{un, bs, f, l, "user"}
		dbUsers[un] = u
		if r == "tog" {
			u = user{un, bs, f, l, "admin"}
			dbUsers[un] = u
		}
			


		


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
	q := req.URL
	if q.String() == "/login?failed=true" {
		tpl.ExecuteTemplate(w, "login.gohtml", "Username and/or password incerrect")
		return
	}

	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		u, ok := dbUsers[un]
		if !ok {
			http.Redirect(w, req, "/login?failed=true", http.StatusSeeOther)
			return
		}

		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))

		if err != nil {
			http.Redirect(w, req, "/login?failed=true", http.StatusSeeOther)
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
	if q.String() != "/login?failed=true" {
		tpl.ExecuteTemplate(w, "login.gohtml", nil)
	}
}

func logout(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(w,req,"/", http.StatusSeeOther)
		return
	}
	c, _ := req.Cookie("session")
	delete(dbSessions, c.Value)
	c = &http.Cookie{
		Name: "session",
		Value: "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

	http.Redirect(w, req, "/login", http.StatusSeeOther)
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
