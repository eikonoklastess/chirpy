package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type responsUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	var loginRequest params

	err := decoder.Decode(&loginRequest)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode response body")
		return
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

	respondWithJSON(w, http.StatusOK, responsUser{
		ID:    user.ID,
		Email: user.Email,
	})
}
