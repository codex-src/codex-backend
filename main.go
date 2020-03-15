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
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	graphql "github.com/graph-gophers/graphql-go"
	"google.golang.org/api/option"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	_ "github.com/lib/pq"
)

// // DELETEME
// const idToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhjZjBjNjQyZDQwOWRlODJlY2M5MjI4ZTRiZDc5OTkzOTZiNTY3NDAiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiWmF5ZGVrIiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hLS9BT2gxNEdpRENxc1Y5bmU1MjdZNnpwejUxOE40VHNNY3J0VU5TNWtubXdSNXV3IiwiaXNzIjoiaHR0cHM6Ly9zZWN1cmV0b2tlbi5nb29nbGUuY29tL2NvZGV4LWVmMzIyIiwiYXVkIjoiY29kZXgtZWYzMjIiLCJhdXRoX3RpbWUiOjE1ODQxMjQyMTEsInVzZXJfaWQiOiJ0WG14SGJpUElDYkRCdDFPckFFbHJUSERTQTQyIiwic3ViIjoidFhteEhiaVBJQ2JEQnQxT3JBRWxyVEhEU0E0MiIsImlhdCI6MTU4NDEyNDIxMywiZXhwIjoxNTg0MTI3ODEzLCJlbWFpbCI6InpheWRla2RvdGNvbUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJnb29nbGUuY29tIjpbIjEwNzI1Mjk3Mzk2NDE0NDA1NzU3MyJdLCJlbWFpbCI6WyJ6YXlkZWtkb3Rjb21AZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoiZ29vZ2xlLmNvbSJ9fQ.pOsCxDnGDNedKB0Xs07-MWmfR76N_Qk6E3S6Wx2zXnZyqZjlma6c1tpuOVTIwK_FjCb-UIpFCczlgIi_nmN8-UQYbibVUVVd_SXMCNU3zXPRAvRXjqpvgnupW4nNg93js7m2lJDggN0qpSRz98_9pcUo9zn7SGDB2BKq82U3RaVN_WAXemoDdXAYR4ePUA2UsFFkO1zcgDJz84dVkzqRK4rDCSu3qXS0CDNK2XYZNTfA7tUyJMzuE3RDLqSV7ckjJmCtSVB2UYTlT_KeWjIEbY2vUzto88jsIBz_RbngUTsRT_w7-vY9Wt4VK_-alu3GR1H7Q8YkU_VCQW_SQ3hpDg"

var (
	// Postgres database
	db *sql.DB

	// Firebase app and authentication client
	app    *firebase.App
	client *auth.Client

	// GraphQL executable schema
	schema *graphql.Schema
)

// Represents a GraphQL query or mutation.
type Query struct {
	UserID    string
	Query     string
	Variables map[string]interface{}
}

type UserID string

const UserIDKey = UserID("userID")

func handler(w http.ResponseWriter, r *http.Request) {
	// CORS:
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Check auth:
	//
	// TODO: Extract to hasAuth
	userID := ""
	if r.Header.Get("Authorization") != "" {
		idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := client.VerifyIDToken(context.TODO(), idToken)
		if err != nil {
			log.Print(fmt.Errorf("error due to client.VerifyUserID: %w", err))
			RespondUnauthorized(w)
			return
		}
		userID = token.UID
	}
	// Parse query:
	var query Query // query := Query{UserID: token.UID}
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		log.Print(fmt.Errorf("error due to json.NewDecoder.Decode: %w", err))
		RespondServerError(w)
		return
	}
	// Execute query and respond:
	ctx := context.TODO()
	ctx = context.WithValue(ctx, UserIDKey, userID)
	res := schema.Exec(ctx, query.Query, "", query.Variables)
	if res.Errors != nil {
		log.Printf("res.Errors=%+v", res.Errors)
	}
	bstr, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		log.Print(fmt.Errorf("error due to json.MarshalIndent: %w", err))
		RespondServerError(w)
		return
	}
	fmt.Fprintln(w, string(bstr))
}

// GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/service-account-file.json"
// app, err := firebase.NewApp(context.Background(), nil)
func main() {
	var err error
	/*
	 * Postgres
	 */
	log.Print("setting up database")
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
	 * Firebase
	 */
	log.Print("setting up firebase auth")
	opt := option.WithCredentialsFile("secret/firebase-admin-sdk.json")
	app, err = firebase.NewApp(context.TODO(), nil, opt)
	must(err, "crash due to firebase.NewApp")
	client, err = app.Auth(context.TODO())
	must(err, "crash due to app.Auth")
	/*
	 * GraphQL
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
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	must(err, "crash due to http.ListenAndServe")
}
