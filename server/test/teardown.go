package test

import (
	"os"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func Teardown(t *testing.T) {
	os.Unsetenv("PRIVATE_KEY")
	os.Unsetenv("PUBLIC_KEY")
	gock.Off()
}
