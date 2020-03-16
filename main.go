package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	graphql "github.com/graph-gophers/graphql-go"
	"google.golang.org/api/option"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	_ "github.com/lib/pq"
)

const UserIDKey = UserID("userID")

var (
	// Postgres database
	db *sql.DB

	// Firebase app and authentication client
	app    *firebase.App
	client *auth.Client

	// GraphQL executable schema
	schema *graphql.Schema
)

type UserID string

// Represents a GraphQL query or mutation.
type Query struct {
	// FIXME: OperationName is not appearing
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// var query Query
// err := json.NewDecoder(r.Body).Decode(&query)
// if err != nil {
// 	log.Print(fmt.Errorf("error due to json.NewDecoder.Decode: %w", err))
// 	RespondServerError(w)
// 	return
// }

// client, err := app.Auth(context.TODO())
// if err != nil {
// 	log.Print(fmt.Errorf("error due to app.Auth: %w", err))
// 	RespondServerError(w)
// 	return
// }

var whiteSpaceRe = regexp.MustCompile(`( |\t|\n)+`)

func handler(w http.ResponseWriter, r *http.Request) {
	// Set headers:
	setHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Get current user ID (authenticated):
	var userID string
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No-op; defer to query or mutation
	} else {
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := client.VerifyIDToken(context.TODO(), idToken)
		if err != nil {
			log.Print(fmt.Errorf("error due to client.VerifyUserID: %w", err))
			RespondUnauthorized(w)
			return
		}
		userID = token.UID
	}
	// Read request body:
	dataIn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(fmt.Errorf("error due to ioutil.ReadAll: %w", err))
		RespondServerError(w)
		return
	}
	// Decode query:
	var query Query
	err = json.Unmarshal(dataIn, &query)
	if err != nil {
		log.Print(fmt.Errorf("error due to json.Unmarshal: %w", err))
		RespondServerError(w)
		return
	}
	// Execute query:
	debugQuery := strings.TrimSpace(whiteSpaceRe.ReplaceAllString(query.Query, " "))
	log.Printf("query=%s variables=%+v", debugQuery, query.Variables)
	ctx := context.WithValue(context.TODO(), UserIDKey, userID)
	resp := schema.Exec(ctx, query.Query, query.OperationName, query.Variables)
	if resp.Errors != nil {
		log.Printf("error due to schema.Exec: %+v", resp.Errors)
	}
	// Encode response:
	dataOut, err := json.Marshal(resp)
	if err != nil {
		log.Print(fmt.Errorf("error due to json.MarshalIndent: %w", err))
		RespondServerError(w)
		return
	}
	// Done:
	fmt.Fprintln(w, string(dataOut))
}

func main() {
	var err error
	/*
	 * Postgres
	 */
	log.Print("setting up postgres")
	db, err = sql.Open("cloudsqlpostgres", fmt.Sprintf(`
		host=codex-ef322:us-west1:codex-db
		user=postgres
		password=%s
		dbname=codex
		sslmode=disable
	`, os.Getenv("PSQL_PW")))
	must(err, "crash due to sql.Open")
	var testStr string
	err = db.QueryRow(`select 'hello, world!'`).Scan(&testStr)
	must(err, "crash due to db.QueryRow")
	defer db.Close()
	/*
	 * Firebase auth
	 */
	log.Print("setting up firebase auth")
	opt := option.WithCredentialsFile("secret/firebase-admin-sdk.json")
	app, err = firebase.NewApp(context.TODO(), nil, opt)
	must(err, "crash due to firebase.NewApp")
	client, err = app.Auth(context.TODO())
	must(err, "crash due to app.Auth")
	/*
	 * Schema
	 */
	log.Print("setting up schema")
	bstr, err := ioutil.ReadFile("schema.graphql")
	must(err, "crash due to ioutil.ReadFile")
	schema, err = graphql.ParseSchema(string(bstr), &RootResolver{})
	must(err, "crash due to graphql.ParseSchema")
	/*
	 * Web server
	 */
	log.Print("setting up web server")
	log.Print("ready")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/graphql", handler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	must(err, "crash due to http.ListenAndServe")
}
