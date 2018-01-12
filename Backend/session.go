package main

import (
	"fmt"
	"net/http"
	"gopkg.in/mgo.v2/bson"

)

func getUser(w http.ResponseWriter, req *http.Request) user {
	c, err := req.Cookie("session")

	if err != nil{
		sID := bson.NewObjectId()
		if err != nil {
			fmt.Println(err)
		}
		c = &http.Cookie{
			Name: "session",
			Value: sID.Hex(),
		}
	}
	var u user
	http.SetCookie(w, c)
	var ValidSession ssession
	db := session.DB("fagside").C("Sessions")
	if err := db.FindId(bson.ObjectIdHex(c.Value)).One(&ValidSession); err != nil {
	    return u
	}
	db = session.DB("fagside").C("Users")
	if err := db.Find(bson.M{"UserName":ValidSession.Un}).One(&u); err != nil {
	    return u
	}

	return u
}

func alreadyLoggedIn(req *http.Request) bool {
	var ValidSession ssession
	var result []ssession
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}
	db := session.DB("fagside").C("Sessions")
	err = db.Find(nil).Sort("-_id").Limit(50).All(&result)
    	if err != nil {
   			fmt.Println(err)
    	}
	if err := db.FindId(bson.ObjectIdHex(c.Value)).One(&ValidSession); err != nil {
	    return false
	}
	return true
}
