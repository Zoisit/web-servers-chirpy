package main

import (
	"log"
	"net/http"
)

type loginParams struct {
	password string
	email    string
}

func (cfg *apiConfig) handlerLogin(rw http.ResponseWriter, req *http.Request) {
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirps, err := cfg.db.GetAllChirps(req.Context())

	if err != nil {
		log.Printf("Error getting all chirps: %s", err)
		respondWithJSON(rw, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusOK, chirps)
}
