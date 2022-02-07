package config

import (
	"os"
	"sync"
)

type container struct {
	m            sync.Mutex
	oAuth2States map[string]struct{}
}

var c container

func init() {
	c = container{
		oAuth2States: map[string]struct{}{},
	}
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
	c.m.Lock()
	defer c.m.Unlock()
	return c.oAuth2States
}
func AddOAuth2State(state string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.oAuth2States[state] = struct{}{}
}
func UseOAuth2State(state string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.oAuth2States[state]; ok {
		delete(c.oAuth2States, state)
		return true
	} else {
		return false
	}
}
