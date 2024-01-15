package noundo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/schema"
)

var jwtCookieKey string

var hmacSecret []byte

var decoder = schema.NewDecoder()

func init() {
	// TODO (change this)
	stringSecret := os.Getenv("FORUM_JWT_HMAC_SECRET")
	if len(stringSecret) == 0 {
		stringSecret = "hmacSampleSecret"
	}
	hmacSecret = []byte(stringSecret)

	// Its for a `Key: Value` pair. Just a random key so no collisions will occur
	jwtCookieKey = "bueykivxivcxf436ugkhgu8owy3886^$&7ae"
}

// Decode `http.Request.Body` into `AuthMe struct`
func GetSignInRequest(r *http.Request) (*SignInRequest, error) {
	err := r.ParseForm()

	if err != nil {
		return nil, err
	}
	var signIn SignInRequest
	err = decoder.Decode(&signIn, r.Form)
	return &signIn, err
}

// Parse http.Request.Body, then check if data passed is valid and if so set
// header on ResponseWriter to JWT with appropriate data.
func SignInUser(auth AuthenticatorIface, w http.ResponseWriter, r *http.Request) error {
	authMe, err := GetSignInRequest(r)
	if err != nil {
		log.Printf("parse Credentials from request: %s\n", err)
		return err
	}

	err = auth.ValidateAuthMe(authMe)
	if err != nil {
		log.Printf("validate Credentials: %s\n", err)
		return err
	}

	user := auth.GetUserByEmail(authMe.Email)

	// TODO: currently unsecure
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		newJWTMapClaims(JWTFields{Username: user.Username()}),
	)

	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		log.Printf("signing token error: %s\n", err)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieKey,
		Value:    tokenString,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode, // TODO: Unsecure?
	})

	return nil
}

func SignOutUser(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieKey,
		Value:    "",
		HttpOnly: true,
	})
}

func GetSignUpRequest(r *http.Request) (*SignUpRequest, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	var signUp SignUpRequest
	err = decoder.Decode(&signUp, r.Form)
	return &signUp, err
}

func SignUpUser(auth AuthenticatorIface, r *http.Request) *SignUpResponse {
	signUp, err := GetSignUpRequest(r)

	if err != nil {
		return &SignUpResponse{RestResp{false, DecodeErr}}
	}

	return auth.SignUpUser(signUp)
}

// Validate the JWT sent with the incoming Request
func JWTCheckAndParse(r *http.Request) (JWTFields, error) {
	var jfEmpty JWTFields
	c, err := r.Cookie(jwtCookieKey)
	if err != nil {
		return jfEmpty, err
	}

	token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSecret, nil
	})

	// token.Valid is (I think) redudant, because it is set just before `return` from `jwt.ParseWithClaims` - which is called by `jwt.Parse`.
	if err != nil || !token.Valid {
		return jfEmpty, err
	}

	// Type assertion. Can be performed on interfaces only.
	// token.Claims is an interface
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jfEmpty, errors.New("failed type assertion: `token.Claims.(jwt.MapClaims)`")
	}

	jf, err := jwtFieldsFromMapClaims(mapClaims)
	if err != nil {
		return jfEmpty, err
	}

	return jf, nil
}
