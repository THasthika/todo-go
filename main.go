package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/globalsign/mgo"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

// Server dependency manager
type Server struct {
	db *mgo.Session
}

// Config structure
type Config struct {
	Host string `env:"HOST,default=0.0.0.0"`
	Port uint16 `env:"PORT,default=8080"`

	Mongo struct {
		Host string `env:"MONGO_HOST,default=localhost"`
		Port uint16 `env:"MONGO_PORT,default=27017"`
		User string `env:"MONGO_USER,required"`
		Pass string `env:"MONGO_PASS,required"`
		Name string `env:"MONGO_NAME,required"`
	}
}

type contextKey struct {
	name string
}

var contextKeyAPIKey = &contextKey{"api-key"}

func main() {

	// INIT APP

	// environment variables loading
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config
	err = envdecode.Decode(&cfg)
	if err != nil {
		log.Fatalln("failed to load environment variables:", err)
	}
	//
	log.Println("Dialing mongo", cfg.Mongo.Host)
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", cfg.Mongo.User, cfg.Mongo.Pass, cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.Name)
	db, err := mgo.Dial(connStr)
	if err != nil {
		log.Fatalln("failed to connect to mongo:", err)
	}
	defer db.Close()
	s := &Server{
		db: db,
	}

	// REGISTER ROUTES
	r := mux.NewRouter()
	r.HandleFunc("/todos", withAPIKey(s.handleTodosGet)).Methods("GET")
	r.HandleFunc("/todos", withAPIKey(s.handleTodoPost)).Methods("POST")
	r.HandleFunc("/todos/{id}", withAPIKey(s.handleTodoGet)).Methods("GET")
	r.HandleFunc("/todos/{id}", withAPIKey(s.handleTodoDelete)).Methods("DELETE")

	listenAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	fmt.Println("Listening on", listenAddr)
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		handlers.CORS()(r),
	)
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-KEY")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}
