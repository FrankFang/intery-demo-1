package model

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var signKey *rsa.PrivateKey

type User struct {
	BaseModel
	Name string `gorm:"type:varchar(100);not null"`
}

func (u User) JWT() string {
	token, err := createToken(u.ID)
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%v", token)
}

type CustomClaims struct {
	*jwt.StandardClaims
	UserId uint `json:"user_id"`
}

func createToken(userId uint) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims = &CustomClaims{
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
		},
		userId,
	}

	signBytes, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}
	return t.SignedString(signKey)
}