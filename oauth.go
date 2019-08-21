package main

import (
	//"encoding/json"
	//"flag"
	"fmt"
	//"net/http"
	//"os"
	//"github.com/strava/go.strava"
)

const authUrlTemplate = "http://www.strava.com/oauth/authorize?client_id=%d&response_type=code&redirect_uri=http://localhost/exchange_token&approval_prompt=force&scope=read"

// Prompts the user (i.e., athlete) to visit a URL to authorize this application
// to read activities from their account. Returns an access token that can be
// used in strava.NewClient().
func Authorize(appClientId int, appClientSecret string) (string, error) {
	athleteAuthUrl := fmt.Sprintf(authUrlTemplate, appClientId)

	// start up http server listening on localhost

	fmt.Printf("Please visit %s to authorize this application to access your account\n", athleteAuthUrl)

	// block until strava hits us to provide auth or not

	return "", nil
}
