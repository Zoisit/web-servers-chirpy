package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/Zoisit/web-servers-chirpy/internal/database"
)

type loginParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserLoginResponse struct {
	User         database.CreateUserRow `json:"user"`
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refresh_token"`
}

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerLogin(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.GetSingleUser(req.Context(), params.Email)

	if err != nil {
		log.Printf("Error getting user: %s", err)
		respondWithJSON(rw, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil {
		log.Printf("Error getting user: %s", err)
		respondWithJSON(rw, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)

	if err != nil {
		log.Printf("Error authenticating user: %s", err)
		respondWithJSON(rw, http.StatusUnauthorized, "Access token could not be created")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error authenticating user: %s", err)
		respondWithJSON(rw, http.StatusUnauthorized, "Refresh token could not be created")
		return
	}

	_, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Printf("Error authenticating user: %s", err)
		respondWithJSON(rw, http.StatusUnauthorized, "Refresh token could not be stored in database")
		return
	}

	u := UserLoginResponse{
		User: database.CreateUserRow{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(rw, http.StatusOK, u)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(w, "Unauthorized: missing or malformed authorization information", http.StatusUnauthorized)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(req.Context(), token)
	if err != nil || time.Now().UTC().After(refreshToken.ExpiresAt) || refreshToken.RevokedAt.Valid {
		http.Error(w, "Unauthorized: missing, revoked or expired token", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		http.Error(w, "Error: could not create new JWT access token", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: newAccessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(w, "Unauthorized: missing or malformed authorization information", http.StatusUnauthorized)
		return
	}

	params := database.RevokeRefreshTokenParams{
		Token: token,
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), params)
	if err != nil {
		http.Error(w, "refresh token does not exist", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
