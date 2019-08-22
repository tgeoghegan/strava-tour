package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/strava/go.strava"
)

// Prompts the user (i.e., athlete) to visit a URL to authorize this application
// to read activities from their account. Returns an access token that can be
// used in strava.NewClient().
func Authorize(appClientId int, appClientSecret string) (string, error) {
	authenticator := strava.OAuthAuthenticator{
		CallbackURL:            "http://localhost:8080/exchange_token",
		RequestClientGenerator: nil,
	}

	// Bizarre: we set the client ID and client secret by assigning to global
	// variables in the strava package.
	strava.ClientId = appClientId
	strava.ClientSecret = appClientSecret

	httpServer := http.Server{Addr: ":8080"}

	// Kinda gross: if we invoke the global http.HandleFunc, it adds handlers to
	// DefaultServeMux. If we leave the Handler field on httpServer nil, it
	// defaults to DefaultServeMux, so we can provide handlers for specific
	// request paths that way, and only that way, as http.Server exposes no
	// method to do so, and I don't want to rewrite path matching.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The Strava Go API doesn't appear to provide enough functionality to
		// define fine grained access scopes, but ViewPrivate is good enough for us
		fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.ViewPrivate, true))
		fmt.Fprint(w, `<img src="http://strava.github.io/api/images/ConnectWithStrava.png" />`)
		fmt.Fprint(w, `</a>`)
	})

	path, err := authenticator.CallbackPath()
	if err != nil {
		return "", err
	}

	tokenChannel := make(chan string, 1)
	errorChannel := make(chan error, 1)

	http.HandleFunc(path, authenticator.HandlerFunc(
		// Auth success
		func(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "SUCCESS:\nAt this point you can use this information to create a new user or link the account to one of your existing users\n")
			fmt.Fprintf(w, "State: %s\n\n", auth.State)
			fmt.Fprintf(w, "Access Token: %s\n\n", auth.AccessToken)

			fmt.Fprintf(w, "The Authenticated Athlete (you):\n")
			content, _ := json.MarshalIndent(auth.Athlete, "", " ")
			fmt.Fprint(w, string(content))

			if auth.AccessToken == "" {
				errorChannel <- fmt.Errorf("no access token in Strava AuthorizationResponse")
			} else {
				tokenChannel <- auth.AccessToken
			}
		},
		// Auth fail
		func(err error, w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Authorization Failure:\n")

			// some standard error checking
			if err == strava.OAuthAuthorizationDeniedErr {
				fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
				fmt.Fprint(w, "This is the main error your application should handle.")
			} else if err == strava.OAuthInvalidCredentialsErr {
				fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
			} else if err == strava.OAuthInvalidCodeErr {
				fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
			} else if err == strava.OAuthServerErr {
				fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
			} else {
				fmt.Fprint(w, err)
			}

			errorChannel <- err
		}))

	go func() {
		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			errorChannel <- fmt.Errorf("Oauth server exited abnormally: %s\n", err)
		}
	}()

	fmt.Printf("Please visit %s to authorize this application to access your account\n", "http://localhost:8080")

	for {
		select {
		case token := <-tokenChannel:
			fmt.Printf("got token %s\n", token)
			return token, nil
		case err := <-errorChannel:
			fmt.Printf("got error\n")
			return "", err
		}
	}
}
