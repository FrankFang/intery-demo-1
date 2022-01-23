package config

import (
	"os"
)

func Init() {

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
