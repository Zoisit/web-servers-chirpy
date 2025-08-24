package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/Zoisit/web-servers-chirpy/internal/database"
)

type CreateUserRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerCreateUser(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := CreateUserRequestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		rw.WriteHeader(500)
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		rw.WriteHeader(500)
		return
	}

	user_params := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
	}
	user, err := cfg.db.CreateUser(req.Context(), user_params)

	if err != nil {
		log.Printf("Error creating user: %s", err)
		rw.WriteHeader(500)
		return
	}

	respondWithJSON(rw, http.StatusCreated, user)
}
