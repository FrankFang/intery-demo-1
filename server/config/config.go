package config

import (
	"os"
)

var oAuth2States = map[string]struct{}{}

func init() {

}

func GetDomain() (domain string) {
	domain = os.Getenv("DOMAIN")
	if domain == "" {
		domain = "http://127.0.0.1:3000"
	}
	return
}

func GetString(name string) string {
	if name == "port" {
		return "0.0.0.0:8080"
	}
	return ""
}

func GetOAuth2States() map[string]struct{} {
	return oAuth2States
}
func AddOAuth2State(state string) {
	oAuth2States[state] = struct{}{}
}
func UseOAuth2State(state string) bool {
	if _, ok := oAuth2States[state]; ok {
		delete(oAuth2States, state)
		return true
	} else {
		return false
	}
}
