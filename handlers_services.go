package main

import "net/http"

func handleCreateService(w http.ResponseWriter, r *http.Request) {
	providerID := r.Header.Get("X-User-ID")
	if providerID == "" {
		writeError(w, newError(ErrUnauthorized, "Authentification requise (X-User-ID manquant)"))
		return
	}

	var s Service
	if err := decodeJSON(r, &s); err != nil {
		writeError(w, err)
		return
	}

	created, err := createService(r.Context(), providerID, s)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func handleGetServices(w http.ResponseWriter, r *http.Request) {
	categorie := r.URL.Query().Get("categorie")
	ville := r.URL.Query().Get("ville")
	search := r.URL.Query().Get("search")

	services, err := listServices(r.Context(), categorie, ville, search)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func handleGetService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s, err := getService(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func handleUpdateService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	providerID := r.Header.Get("X-User-ID")

	var s Service
	if err := decodeJSON(r, &s); err != nil {
		writeError(w, err)
		return
	}

	if err := updateService(r.Context(), id, providerID, s); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Service mis à jour avec succès"})
}

func handleDeleteService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	providerID := r.Header.Get("X-User-ID")

	if err := deleteService(r.Context(), id, providerID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Service supprimé avec succès"})
}

func handleGetServiceReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	reviews, err := getServiceReviews(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}
