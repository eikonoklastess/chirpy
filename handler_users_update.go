package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	type responsUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if ok && token.Valid {
		fmt.Printf("token is valid. Claims: %+v\n", claims)
	} else {
		respondWithError(w, http.StatusInternalServerError, "claim is either not ok or token invalid")
	}

	idString, err := claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while getting token's claim's subject")
		return
	}

	userId, err := strconv.Atoi(idString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while converting id int")
		return
	}

	param := params{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while decoding request body")
		return
	}

	err = cfg.DB.UpdateUser(userId, param.Email, param.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while updating user")
		return
	}

	respondWithJSON(w, http.StatusOK, responsUser{
		ID:    userId,
		Email: param.Email,
	})
}
