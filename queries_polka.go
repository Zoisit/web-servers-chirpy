package main

import (
	"encoding/json"
	"net/http"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/google/uuid"
)

type PolkaWebhooksParams struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerPolkaWebhooks(rw http.ResponseWriter, req *http.Request) {
	key, err := auth.GetAPIKey(req.Header)
	if err != nil || key != cfg.polka {
		http.Error(rw, "Unauthorized: missing, wrong or malformed authorization information", http.StatusUnauthorized)
		return
	}

	//stop if it is not the expected event type
	decoder := json.NewDecoder(req.Body)
	params := PolkaWebhooksParams{}
	err = decoder.Decode(&params)
	if err != nil || params.Event != "user.upgraded" {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.db.UpgradeSingleUserToChirpyRed(req.Context(), params.Data.UserID)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
