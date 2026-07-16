package main

import "testing"

func TestUserHasSkill(t *testing.T) {
	skills := []Skill{
		{Nom: "Jardinage", Niveau: "expert"},
		{Nom: " Bricolage ", Niveau: "débutant"},
	}

	cases := []struct {
		name      string
		categorie string
		want      bool
	}{
		{"correspondance exacte", "Jardinage", true},
		{"insensible à la casse", "jardinage", true},
		{"tolère les espaces autour du nom stocké", "Bricolage", true},
		{"compétence non possédée", "Cuisine", false},
		{"chaîne vide", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := userHasSkill(skills, tc.categorie); got != tc.want {
				t.Errorf("userHasSkill(%v, %q) = %v, want %v", skills, tc.categorie, got, tc.want)
			}
		})
	}
}

func TestUserHasSkillEmptyList(t *testing.T) {
	if userHasSkill(nil, "Jardinage") {
		t.Error("un utilisateur sans compétence ne devrait jamais matcher")
	}
}

func TestIsSelfExchange(t *testing.T) {
	cases := []struct {
		name        string
		ownerID     int
		requesterID string
		want        bool
	}{
		{"même utilisateur", 5, "5", true},
		{"utilisateurs différents", 5, "6", false},
		{"requesterID non numérique", 5, "abc", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSelfExchange(tc.ownerID, tc.requesterID); got != tc.want {
				t.Errorf("isSelfExchange(%d, %q) = %v, want %v", tc.ownerID, tc.requesterID, got, tc.want)
			}
		})
	}
}

func TestIsValidNote(t *testing.T) {
	cases := []struct {
		note int
		want bool
	}{
		{0, false},
		{1, true},
		{3, true},
		{5, true},
		{6, false},
		{-1, false},
	}

	for _, tc := range cases {
		if got := isValidNote(tc.note); got != tc.want {
			t.Errorf("isValidNote(%d) = %v, want %v", tc.note, got, tc.want)
		}
	}
}

func TestIsActiveExchangeStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{"pending", true},
		{"accepted", true},
		{"completed", false},
		{"rejected", false},
		{"cancelled", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := isActiveExchangeStatus(tc.status); got != tc.want {
			t.Errorf("isActiveExchangeStatus(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestIsValidCategory(t *testing.T) {
	cases := []struct {
		categorie string
		want      bool
	}{
		{"Jardinage", true},
		{"Autre", true},
		{"jardinage", false}, // la casse compte pour la liste fermée
		{"Inexistant", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := isValidCategory(tc.categorie); got != tc.want {
			t.Errorf("isValidCategory(%q) = %v, want %v", tc.categorie, got, tc.want)
		}
	}
}
