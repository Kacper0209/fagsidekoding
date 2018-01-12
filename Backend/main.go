package main

import (
	"image"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)
//Lage struct til kommentarer, brukere og blogginnlegg
type user struct {
	UserName  string    `bson:"UserName"`
	Password  []byte
	First     string
	Last      string
	FullName  string
	Role      string
}
type comment struct {
	Bn       string
	Bl       string
	Content  string
}

type blog struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	IDS		 string
	Title    string  	   
	By       string        
	Content  string        
	Comment  []comment     `bson:"cmtarray"`
}
type ssession struct {
	BID      bson.ObjectId `bson:"_id,omitempty"`
	Un       string
}

//deklarere variabler med package scope 
var session *mgo.Session
var img image.Image
var tpl *template.Template

func init() {
	//hente templates
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	//Koble til database
	s, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	session = s
	defer session.Close()
	//lage alle sidene og koble til funksjoner
	http.HandleFunc("/",index)
	http.HandleFunc("/blogg",blogg)
	http.HandleFunc("/aktiviteter",activities)
	http.HandleFunc("/kontakt",kontakt)
	http.HandleFunc("/signup",signup)
	http.HandleFunc("/login",login)
	http.HandleFunc("/logout",logout)
	http.HandleFunc("/newblog",nyblogg)
	//CSS og bilder
  	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
  	http.Handle("favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
} 
func activities(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	tpl.ExecuteTemplate(w,"activities.gohtml",nil)
}
func kontakt(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w,"contact.gohtml", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func blogg(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	//get user
	u := getUser(w,req)

	//Åpne database
	db := session.DB("fagside").C("Blogg")

	//Variabler
	var blogg blog
	var result []blog

	//Har URL en ID query?
	q := req.URL.Query()
    bID := q.Get("id")
    if bID == "" {
    	//Hvis ikke, finn alle blogginlegg
    	err := db.Find(nil).Sort("-_id").Limit(50).All(&result)
    	if err != nil {
   			fmt.Println(err)
    	}
    	//execute
    	tpl.ExecuteTemplate(w, "blogg.gohtml", result)
    	return
    }else {

    	//Ellers finn blogginnlegg med spesefik ID
    	//Hente form verdier(Kommentar)
	    if req.Method == http.MethodPost {
	    	if alreadyLoggedIn(req) {
		    	com := req.FormValue("content")
		    	FN := u.First
		    	LN := u.Last
		    	cmt:= comment{
		    		Bn : FN,
		    		Bl : LN,
		    		Content : com, 
		    		}
		    	//Oppdatere databasen
		    	blogpost := bson.M{"_id" : bson.ObjectIdHex(bID)}
				PushToArray := bson.M{"$push": bson.M{"cmtarray":cmt}}
				err := db.Update(blogpost, PushToArray)
				if err != nil {
					fmt.Println(err)
				}		
			}else {
				w.WriteHeader(http.StatusForbidden)
			}
		}
		if err := db.FindId(bson.ObjectIdHex(bID)).One(&blogg); err != nil {
	    		fmt.Println(err)
	    }
    	//execute
    	tpl.ExecuteTemplate(w,"id.gohtml", blogg)
    	return
    }
}


func nyblogg(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	u := getUser(w, req)
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	 	return
	}
	if req.Method == http.MethodPost {

		var b blog
		titl := req.FormValue("title")
		cont := req.FormValue("content")
		bID := bson.NewObjectId()
		sid := bID.Hex()
		b = blog{Id: bID,IDS: sid,Title: titl, By: u.FullName, Content: cont}
		db := session.DB("fagside").C("Blogg")
		err := db.Insert(&b)
		if err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, req, "/blogg", http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(w, "NewBlog.gohtml", u)
}

func signup(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	var u user
	if req.Method == http.MethodPost {
		un := req.FormValue("username")	
		p := req.FormValue("password")	
		pr := req.FormValue("repassword")	
		f := req.FormValue("firstname")	
		l := req.FormValue("lastname")
		db := session.DB("fagside").C("Users")
		if err := db.Find(bson.M{"UserName": un}).One(nil); err == nil{
			http.Redirect(w,req,"/signup",http.StatusSeeOther)
			return
		}
		if p != pr{
			http.Redirect(w,req,"/signup",http.StatusSeeOther)
			return
		}
		sID := bson.NewObjectId()
		c := &http.Cookie {
			Name: "session",
			Value: sID.Hex(),
		}
		csession := ssession{
			BID : sID,
			Un  : un, 
		}
		http.SetCookie(w, c)
		bs, err := bcrypt.GenerateFromPassword([]byte(p),bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fullname := f+" "+l
		u = user{un, bs, f, l,fullname,"user"}
		err = db.Insert(&u)

		db = session.DB("fagside").C("Sessions")
		err = db.Insert(&csession)
		var test user

		err = db.Find(bson.M{"UserName": un}).One(&test)
		fmt.Println(test)
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
		}
		tpl.ExecuteTemplate(w, "signup.gohtml", u)
	}

func login(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	if alreadyLoggedIn(req) {
		http.Redirect(w, req,"/", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		db := session.DB("fagside").C("Users")
		var test user
		err := db.Find(bson.M{"UserName": un}).One(&test)
		if err != nil {
			http.Redirect(w, req, "/login?failed=true", http.StatusSeeOther)
			return
		}
		err = bcrypt.CompareHashAndPassword(test.Password, []byte(p))
		if err != nil {
			http.Redirect(w, req, "/login?failed=true", http.StatusSeeOther)
			return
		}
		bid := bson.NewObjectId()
		c := &http.Cookie {
			Name: "session",
			Value: bid.Hex(),
		}
		csession := ssession{
			BID: bid,
			Un : un,
		}
		http.SetCookie(w, c)
		db = session.DB("fagside").C("Sessions")
		err = db.Insert(&csession)
		http.Redirect(w, req, "/", http.StatusSeeOther)
	}
	q := req.URL
	if q.String() == "/login?failed=true" {
		tpl.ExecuteTemplate(w, "login.gohtml", "Username and/or password incorrect")
		return
	}
	if q.String() != "/login?failed=true" {
		tpl.ExecuteTemplate(w, "login.gohtml", nil)
		return
	}
}

func logout(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type","text/html;charset=utf-8;")
	if !alreadyLoggedIn(req) {
		http.Redirect(w,req,"/", http.StatusSeeOther)
		return
	}
	db := session.DB("fagside").C("Sessions")
	cookie, _ := req.Cookie("session")
	err := db.Remove(bson.M{"_id": bson.ObjectIdHex(cookie.Value)})
	if err != nil{
		fmt.Println(err)
	}
	cookie = &http.Cookie{
		Name: "session",
		Value: "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}
