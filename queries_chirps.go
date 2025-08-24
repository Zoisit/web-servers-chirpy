package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/Zoisit/web-servers-chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := database.CreateChirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting authorization from header: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("401: %s", err.Error())
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	params.UserID = id

	if len(params.Body) > 140 {
		type returnErr struct {
			Error string `json:"error"`
		}

		respBody := returnErr{
			Error: "Chirp is too long",
		}

		respondWithJSON(rw, 400, respBody)
		return
	}

	params.Body = cleanChirp(params.Body)

	chirp, err := cfg.db.CreateChirp(req.Context(), params)

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithJSON(rw, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerGetAllChirps(rw http.ResponseWriter, req *http.Request) {
	//determine sorting order
	sort_desc := false
	s := req.URL.Query().Get("sort")
	if s == "desc" {
		sort_desc = true
	}

	//get author, if given
	s = req.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if s == "" {
		chirps, err = cfg.db.GetAllChirps(req.Context())
	} else {
		id, err := uuid.Parse(s)
		if err != nil {
			http.Error(rw, "Given author id is not a valid uuid", http.StatusInternalServerError)
			return
		}
		chirps, err = cfg.db.GetAllChirpsFromAuthor(req.Context(), id)
	}

	if err != nil {
		log.Printf("Error getting all chirps: %s", err)
		respondWithJSON(rw, http.StatusNotFound, err.Error())
		return
	}

	if sort_desc {
		chirps = reverseCopy(chirps)
	}

	respondWithJSON(rw, http.StatusOK, chirps)
}

func reverseCopy[T any](original []T) []T {
	reversed := make([]T, len(original))
	copy(reversed, original)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (cfg *apiConfig) handlerGetSingleChirp(rw http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	chirp, err := cfg.db.GetSingleChirp(req.Context(), id)

	if err != nil {
		log.Printf("Error fetching chirp: %s", err)
		respondWithJSON(rw, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusOK, chirp)
}

func respondWithJSON(rw http.ResponseWriter, rc int, rb interface{}) {
	dat, err := json.Marshal(rb)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(rc)
	rw.Write(dat)
}

func cleanChirp(msg string) string {
	words := strings.Split(msg, " ")
	bad := [...]string{"kerfuffle", "sharbert", "fornax"}

	for i, w := range words {
		for _, b := range bad {
			if strings.ToLower(w) == b {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerDeleteSingleChirp(rw http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		http.Error(rw, "No chirp given", http.StatusNotFound)
		return
	}

	//authenticate user
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(rw, "Error getting authorization bearer from header", http.StatusUnauthorized)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	//check authorship
	chirp, err := cfg.db.GetSingleChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirp: %s", err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if id != chirp.UserID {
		log.Printf("User id does not match: \n Expected: %s \nGiven: %s", chirp.UserID, id)
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	//delete chirp
	err = cfg.db.DeleteSingleChirp(req.Context(), chirpID)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
