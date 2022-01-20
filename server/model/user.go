package model

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var signKey *rsa.PrivateKey

type User struct {
	BaseModel
	Name string `gorm:"type:varchar(100);not null"`
}

func (u User) JWT() (token string, err error) {
	token, err = createToken(u.ID)
	return
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
		log.Println(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Println(err)
	}
	return t.SignedString(signKey)
}
