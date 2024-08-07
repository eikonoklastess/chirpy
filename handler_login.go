package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	const defaultLifetime = 90
	type params struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		TokenLifetime int    `json:"expires_in_seconds"`
	}
	type responsUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}
	type NumericDate struct {
		Timestamps int64 `json:"timestamps"`
	}

	decoder := json.NewDecoder(r.Body)
	var loginRequest params

	err := decoder.Decode(&loginRequest)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode response body")
		return
	}

	if loginRequest.TokenLifetime == 0 {
		loginRequest.TokenLifetime = defaultLifetime
	}

	user, err := cfg.DB.EmailGetUser(loginRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "no matching user with this email found")
		return
	}

	if bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(loginRequest.Password)) != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect password")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(loginRequest.TokenLifetime) * time.Second)),
		Subject:   strconv.Itoa(user.ID),
	})
	signedJWT, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "creationg of signed string failed")
		return
	}
	fmt.Println(signedJWT)

	respondWithJSON(w, http.StatusOK, responsUser{
		ID:    user.ID,
		Email: user.Email,
		Token: signedJWT,
	})
}
