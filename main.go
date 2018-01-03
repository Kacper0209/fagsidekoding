package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Et blogginnlegg
type Post struct {
	ID      bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name    string
	Content string
}

var session mgo.Session

func getPostHandler(w http.ResponseWriter, r *http.Request) {

	//TODO Ikke lag en ny session hver gang
	s, err1 := mgo.Dial("127.0.0.1")
	if err1 != nil {
		panic(err1)
	}
	ID := r.URL.Query().Get("id")

	defer s.Close()
	c := s.DB("test").C("posts")
	fmt.Println(c.FullName)

	result := Post{}
	err := c.FindId(bson.ObjectIdHex(ID)).One(&result)
	if err != nil {
		w.WriteHeader(402)
		return
	}

	/*result := Post{bson.NewObjectId(), "Gjøre noen ting", "Jeg vil at du skal gjøre noe"}*/
	b, err := json.Marshal(result)
	if err != nil {
		return
	}

	/*t := Thing{"Gjøre noen ting", "add3d8de-420c-476c-8b81-5a6db123f579", "Jeg vil at du skal gjøre noe"}
	b, err := json.Marshal(t)
	if err != nil {
		return
	}*/

	w.Write(b)
}
func addPostHandler(w http.ResponseWriter, r *http.Request) {

	//TODO Samme her
	s, err1 := mgo.Dial("127.0.0.1")
	if err1 != nil {
		w.WriteHeader(500)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var p Post
	err := decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(403)
		return
	}
	objID := bson.NewObjectId()
	p.ID = objID

	c := s.DB("test").C("posts")
	err = c.Insert(&p)
	fmt.Println(objID.Hex())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func main() {
	//gir mening når man kopier denne og ikke lager sin egen session
	/*session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}

	defer session.Close()*/

	http.HandleFunc("/load", getPostHandler)
	http.HandleFunc("/save", addPostHandler)
	http.ListenAndServe(":8080", nil)
}
