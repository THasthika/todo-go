package main

import (
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type todo struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Title string        `json:"title"`
	Done  bool          `json:"done"`
}

func homeGet(w http.ResponseWriter, r *http.Request) {
	x := struct {
		Ok string `json:"ok"`
	}{
		Ok: "OK!",
	}
	respond(w, r, 200, x)
}

func (s *Server) handleTodosGet(w http.ResponseWriter, r *http.Request) {
	session := s.db.Copy()
	defer session.Close()
	q := session.DB("").C("todos").Find(nil)
	var result []*todo
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}

func (s *Server) handleTodoGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["id"]
	session := s.db.Copy()
	defer session.Close()
	if !bson.IsObjectIdHex(todoID) {
		respondErr(w, r, http.StatusInternalServerError, "Not an Object ID")
		return
	}
	q := session.DB("").C("todos").FindId(bson.ObjectIdHex(todoID))
	var result *todo
	if err := q.One(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}

func (s *Server) handleTodoPost(w http.ResponseWriter, r *http.Request) {
	session := s.db.Copy()
	defer session.Close()
	c := session.DB("").C("todos")
	var t todo
	if err := decodeBody(r, &t); err != nil {
		respondErr(w, r, http.StatusBadRequest, "failed to read todo from request", err)
		return
	}
	t.ID = bson.NewObjectId()
	if err := c.Insert(t); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "failed to insert todo", err)
		return

	}
	respond(w, r, http.StatusCreated, t)
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["id"]
	session := s.db.Copy()
	defer session.Close()
	if !bson.IsObjectIdHex(todoID) {
		respondErr(w, r, http.StatusInternalServerError, "Not an Object ID")
		return
	}
	if err := session.DB("").C("todos").RemoveId(bson.ObjectIdHex(todoID)); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, nil)
}
