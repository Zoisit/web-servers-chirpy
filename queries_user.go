package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/Zoisit/web-servers-chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateUser(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := database.CreateUserParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		rw.WriteHeader(500)
		return
	}

	params.HashedPassword, err = auth.HashPassword(params.HashedPassword)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		rw.WriteHeader(500)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params)

	if err != nil {
		log.Printf("Error creating user: %s", err)
		rw.WriteHeader(500)
		return
	}

	respondWithJSON(rw, http.StatusCreated, user)
}
