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

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	graphql "github.com/graph-gophers/graphql-go"
	"google.golang.org/api/option"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	_ "github.com/lib/pq"
)

var (
	// Postgres
	db *sql.DB

	// Firebase
	app    *firebase.App
	client *auth.Client

	// GraphQL
	schema *graphql.Schema
)

const idToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhjZjBjNjQyZDQwOWRlODJlY2M5MjI4ZTRiZDc5OTkzOTZiNTY3NDAiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiWmF5ZGVrIiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hLS9BT2gxNEdpRENxc1Y5bmU1MjdZNnpwejUxOE40VHNNY3J0VU5TNWtubXdSNXV3IiwiaXNzIjoiaHR0cHM6Ly9zZWN1cmV0b2tlbi5nb29nbGUuY29tL2NvZGV4LWVmMzIyIiwiYXVkIjoiY29kZXgtZWYzMjIiLCJhdXRoX3RpbWUiOjE1ODQxMjQyMTEsInVzZXJfaWQiOiJ0WG14SGJpUElDYkRCdDFPckFFbHJUSERTQTQyIiwic3ViIjoidFhteEhiaVBJQ2JEQnQxT3JBRWxyVEhEU0E0MiIsImlhdCI6MTU4NDEyNDIxMywiZXhwIjoxNTg0MTI3ODEzLCJlbWFpbCI6InpheWRla2RvdGNvbUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJnb29nbGUuY29tIjpbIjEwNzI1Mjk3Mzk2NDE0NDA1NzU3MyJdLCJlbWFpbCI6WyJ6YXlkZWtkb3Rjb21AZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoiZ29vZ2xlLmNvbSJ9fQ.pOsCxDnGDNedKB0Xs07-MWmfR76N_Qk6E3S6Wx2zXnZyqZjlma6c1tpuOVTIwK_FjCb-UIpFCczlgIi_nmN8-UQYbibVUVVd_SXMCNU3zXPRAvRXjqpvgnupW4nNg93js7m2lJDggN0qpSRz98_9pcUo9zn7SGDB2BKq82U3RaVN_WAXemoDdXAYR4ePUA2UsFFkO1zcgDJz84dVkzqRK4rDCSu3qXS0CDNK2XYZNTfA7tUyJMzuE3RDLqSV7ckjJmCtSVB2UYTlT_KeWjIEbY2vUzto88jsIBz_RbngUTsRT_w7-vY9Wt4VK_-alu3GR1H7Q8YkU_VCQW_SQ3hpDg"

var Rx = RootResolver{}

type RootResolver struct{}

func (r *RootResolver) Ping(ctx context.Context) string {
	return "pong"
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	// Assert auth:
	//
	// https://firebase.google.com/docs/auth/admin/verify-id-tokens#verify_id_tokens_using_the_firebase_admin_sdk
	token, err := client.VerifyIDToken(ctx, idToken) // FIXME: Decode idToken from header?
	if err != nil {
		err = fmt.Errorf("error due to client.VerifyIDToken: %w", err)
		log.Print(err)
		RespondUnauthorized(w)
		return
	}
	bstr, err := json.MarshalIndent(token, "", "\t")
	if err != nil {
		err = fmt.Errorf("error due to json.MarshalIndent: %w", err)
		log.Print(err)
		RespondUnauthorized(w)
		return
	}
	fmt.Println(string(bstr))
}

// GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/service-account-file.json"
// func main() {
// 	app, err := firebase.NewApp(context.Background(), nil)
// 	must(err)
// }

func main() {
	// Connect to Postgres:
	var err error
	db, err = sql.Open("cloudsqlpostgres", fmt.Sprintf(`
		host=codex-ef322:us-west1:codex-db
		user=postgres
		password=%s
		dbname=codex
		sslmode=disable
	`, os.Getenv("CLOUD_SQL_PASSWORD")))
	must(err, "error due to sql.Open")
	defer db.Close()

	// Setup Firebase (1 of 2):
	opt := option.WithCredentialsFile(".secret/firebase-admin-sdk.json")
	app, err = firebase.NewApp(context.TODO(), nil, opt)
	must(err, "error due to firebase.NewApp")
	client, err = app.Auth(context.TODO())
	must(err, "error due to app.Auth")

	// Parse GraphQL schema:
	bstr, err := ioutil.ReadFile("schema.graphql")
	must(err, "error due to ioutil.ReadFile")
	schema, err = graphql.ParseSchema(string(bstr), &RootResolver{})
	must(err, "error due to graphql.ParseSchema")

	// Listen and serve:
	log.Print("ok")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	must(err, "error due to http.ListenAndServe")
}

// // Postgres database:
// var DB *sql.DB
//
// // GraphQL schema:
// var Schema *graphql.Schema
//
// type Query struct {
// 	Query     string
// 	Variables map[string]interface{}
// }
//
// func handleGraphQL(w http.ResponseWriter, r *http.Request) {
// 	// Enable cross-origin resource sharing:
// 	writeCORSHeaders(w)
// 	if r.Method == "OPTIONS" {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
// 	// Extend the current session if authenticated:
// 	curr, err := ExtendCurrentSession(w, r)
// 	if err != nil {
// 		http.Error(w, "500 Server Error", http.StatusInternalServerError)
// 		check(err, "ExtendCurrentSession")
// 		return
// 	}
// 	// Create a context with the current session as a value:
// 	ctx := WithCurrentSession(context.Background(), curr)
// 	// Unmarshal query and variables:
// 	var query Query
// 	err = json.NewDecoder(r.Body).Decode(&query)
// 	if err != nil {
// 		http.Error(w, "500 Server Error", http.StatusInternalServerError)
// 		check(err, "json.NewDecoder")
// 		return
// 	}
// 	// Execute query and marshal response and errors:
// 	res := Schema.Exec(ctx, query.Query, "", query.Variables)
// 	b, err := json.MarshalIndent(res, "", "\t")
// 	if err != nil {
// 		http.Error(w, "500 Server Error", http.StatusInternalServerError)
// 		check(err, "json.MarshalIndent")
// 		return
// 	}
// 	// Write response:
// 	fmt.Fprintln(w, string(b))
// }
//
// func main() {
// 	// Connect to the database:
// 	var err error
// 	DB, err = sql.Open("postgres", "postgres://zaydek@localhost/codex?sslmode=disable")
// 	must(err, "sql.Open")
// 	err = DB.Ping()
// 	must(err, "DB.Ping")
// 	defer DB.Close()
// 	// Parse the schema:
// 	b, err := ioutil.ReadFile("schema.graphql")
// 	must(err, "ioutil.ReadFile")
// 	Schema, err = graphql.ParseSchema(string(b), &RootRx{})
// 	must(err, "graphql.ParseSchema")
// 	// Listen and serve:
// 	http.HandleFunc("/graphql", handleGraphQL)
// 	// http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
// 	// 	w.WriteHeader(http.StatusOK)
// 	// 	return
// 	// })
// 	err = http.ListenAndServe(":8000", nil)
// 	must(err, "http.ListenAndServe")
// }
