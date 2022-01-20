package config

import (
	"log"
	"os"
)

func Init() {

}

func GetDomain() (domain string) {
	domain = os.Getenv("DOMAIN")
	log.Println("domain: " + domain)
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
