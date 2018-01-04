package main

import (
	"log"
	"github.com/satori/go.uuid"
	"net/http"

)

func getUser(w http.ResponseWriter, req *http.Request) user {
	c, err := req.Cookie("session")
	if err != nil{
		sID, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
		}
		c = &http.Cookie{
			Name: "session",
			Value: sID.String(),
		}
	}
	http.SetCookie(w, c)

	var u user
	if un, ok := dbSessions[c.Value]; ok {
		u = dbUsers[un]
	}

	return u
}

func alreadyLoggedIn(req *http.Request) bool {

	c, err := req.Cookie("session")
	if err != nil {

		return false
	}
	un := dbSessions[c.Value]
	_, ok := dbUsers[un]
	
	return ok
}
