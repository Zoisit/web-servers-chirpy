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
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	user_params := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
	}
	user, err := cfg.db.CreateUser(req.Context(), user_params)

	if err != nil {
		log.Printf("Error creating user: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(rw, http.StatusCreated, user)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, req *http.Request) {
	//get acces token
	accesToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(w, "Unauthorized: missing or malformed authorization information", http.StatusUnauthorized)
		return
	}

	//get user from acces token
	id, err := auth.ValidateJWT(accesToken, cfg.secret)
	if err != nil {
		http.Error(w, "Unauthorized: JWT validation failed", http.StatusUnauthorized)
		return
	}

	//get request data
	decoder := json.NewDecoder(req.Body)
	params := CreateUserRequestParams{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//make new password
	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//update user information in database
	sqlParams := database.UpdateSingleUserParams{
		ID:             id,
		Email:          params.Email,
		HashedPassword: params.Password,
	}
	user, err := cfg.db.UpdateSingleUser(req.Context(), sqlParams)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}
