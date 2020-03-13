package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var (
	app    *firebase.App
	client *auth.Client
)

const idToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhjZjBjNjQyZDQwOWRlODJlY2M5MjI4ZTRiZDc5OTkzOTZiNTY3NDAiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiWmF5ZGVrIiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hLS9BT2gxNEdpRENxc1Y5bmU1MjdZNnpwejUxOE40VHNNY3J0VU5TNWtubXdSNXV3IiwiaXNzIjoiaHR0cHM6Ly9zZWN1cmV0b2tlbi5nb29nbGUuY29tL2NvZGV4LWVmMzIyIiwiYXVkIjoiY29kZXgtZWYzMjIiLCJhdXRoX3RpbWUiOjE1ODQxMjQyMTEsInVzZXJfaWQiOiJ0WG14SGJpUElDYkRCdDFPckFFbHJUSERTQTQyIiwic3ViIjoidFhteEhiaVBJQ2JEQnQxT3JBRWxyVEhEU0E0MiIsImlhdCI6MTU4NDEyNDIxMywiZXhwIjoxNTg0MTI3ODEzLCJlbWFpbCI6InpheWRla2RvdGNvbUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJnb29nbGUuY29tIjpbIjEwNzI1Mjk3Mzk2NDE0NDA1NzU3MyJdLCJlbWFpbCI6WyJ6YXlkZWtkb3Rjb21AZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoiZ29vZ2xlLmNvbSJ9fQ.pOsCxDnGDNedKB0Xs07-MWmfR76N_Qk6E3S6Wx2zXnZyqZjlma6c1tpuOVTIwK_FjCb-UIpFCczlgIi_nmN8-UQYbibVUVVd_SXMCNU3zXPRAvRXjqpvgnupW4nNg93js7m2lJDggN0qpSRz98_9pcUo9zn7SGDB2BKq82U3RaVN_WAXemoDdXAYR4ePUA2UsFFkO1zcgDJz84dVkzqRK4rDCSu3qXS0CDNK2XYZNTfA7tUyJMzuE3RDLqSV7ckjJmCtSVB2UYTlT_KeWjIEbY2vUzto88jsIBz_RbngUTsRT_w7-vY9Wt4VK_-alu3GR1H7Q8YkU_VCQW_SQ3hpDg"

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
	// Setup Firebase (1 of 2):
	//
	// https://firebase.google.com/docs/admin/setup
	var err error
	opt := option.WithCredentialsFile(".secret/firebase-admin-sdk.json")
	app, err = firebase.NewApp(context.TODO(), nil, opt)
	must(err, "error due to firebase.NewApp")
	// Setup Firebase (2 of 2):
	client, err = app.Auth(context.TODO())
	must(err, "error due to app.Auth")
	// Listen and serve:
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}
