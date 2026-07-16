package main

import "net/http"


func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := decodeJSON(r, &u); err != nil {
		writeError(w, err)
		return
	}

	created, err := createUser(r.Context(), u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	u, err := getUserProfile(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isOwner(r, id) {
		writeError(w, newError(ErrUnauthorized, "Non autorisé"))
		return
	}

	var u User
	if err := decodeJSON(r, &u); err != nil {
		writeError(w, err)
		return
	}

	if err := updateUserProfile(r.Context(), id, u); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Profil mis à jour"})
}

func handleGetSkills(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	skills, err := getUserSkills(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func handleUpdateSkills(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isOwner(r, id) {
		writeError(w, newError(ErrUnauthorized, "Non autorisé"))
		return
	}

	var skills []Skill
	if err := decodeJSON(r, &skills); err != nil {
		writeError(w, err)
		return
	}

	if err := setUserSkills(r.Context(), id, skills); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Compétences mises à jour"})
}

func handleGetUserReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	reviews, err := getUserReviews(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}
