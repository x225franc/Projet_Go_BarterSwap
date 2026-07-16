package main

import "net/http"


func handleCreateExchange(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Header.Get("X-User-ID")
	if requesterID == "" {
		writeError(w, newError(ErrUnauthorized, "Authentification requise"))
		return
	}

	var req createExchangeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	id, err := createExchange(r.Context(), requesterID, req.ServiceID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"message": "Demande créée", "exchange_id": id})
}

func handleGetExchanges(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	status := r.URL.Query().Get("status")

	exchanges, err := listExchanges(r.Context(), userID, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchanges)
}

func handleGetExchangeByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")

	e, err := getExchange(r.Context(), id, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func handleAcceptExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ownerID := r.Header.Get("X-User-ID")

	if err := acceptExchange(r.Context(), id, ownerID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Echange accepté, crédits bloqués"})
}

func handleRejectExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ownerID := r.Header.Get("X-User-ID")

	if err := rejectExchange(r.Context(), id, ownerID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Echange refusé"})
}

func handleCompleteExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	requesterID := r.Header.Get("X-User-ID")

	if err := completeExchange(r.Context(), id, requesterID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Echange terminé, crédits transférés"})
}

func handleCancelExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")

	message, err := cancelExchange(r.Context(), id, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": message})
}

func handleCreateReview(w http.ResponseWriter, r *http.Request) {
	exchangeID := r.PathValue("id")
	authorID := r.Header.Get("X-User-ID")
	if authorID == "" {
		writeError(w, newError(ErrUnauthorized, "Authentification requise"))
		return
	}

	var rev Review
	if err := decodeJSON(r, &rev); err != nil {
		writeError(w, err)
		return
	}

	created, err := createReview(r.Context(), exchangeID, authorID, rev)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
