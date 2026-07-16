package main

import (
	"net/http"
	"testing"
)

func TestGetUserSkillsAndReviews(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")

	status, skills := apiCall(t, srv, http.MethodGet, "/api/users/"+bob+"/skills", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if _, ok := skills["skills"]; ok {
		t.Fatalf("la réponse devrait être un tableau JSON, pas un objet: %v", skills)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/users/"+bob+"/reviews", "", nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
}

func TestListAndFilterServices(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	apiCreateService(t, srv, bob, "Bricolage", 5)

	status, body := apiCall(t, srv, http.MethodGet, "/api/services", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%v", status, http.StatusOK, body)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/services?categorie=Bricolage&ville=Lyon", "", nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d (filtre categorie+ville)", status, http.StatusOK)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/services?search=Service", "", nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d (recherche textuelle)", status, http.StatusOK)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/services?categorie=Inexistante", "", nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d (filtre sans résultat)", status, http.StatusOK)
	}
}

func TestGetServiceByID_NotFound(t *testing.T) {
	srv := testServer(t)

	status, _ := apiCall(t, srv, http.MethodGet, "/api/services/999", "", nil)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
}

func TestUpdateService_RejectsNonOwner(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, _ := apiCall(t, srv, http.MethodPut, "/api/services/"+serviceID, alice, map[string]any{
		"titre": "Volé", "categorie": "Bricolage", "duree_minutes": 30, "credits": 1,
	})
	if status != http.StatusForbidden {
		t.Errorf("status = %d, want %d", status, http.StatusForbidden)
	}
}

func TestUpdateService_AllowsOwner(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, body := apiCall(t, srv, http.MethodPut, "/api/services/"+serviceID, bob, map[string]any{
		"titre": "Nouveau titre", "categorie": "Bricolage", "duree_minutes": 45, "credits": 6,
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%v", status, http.StatusOK, body)
	}

	_, service := apiCall(t, srv, http.MethodGet, "/api/services/"+serviceID, "", nil)
	if service["titre"] != "Nouveau titre" {
		t.Errorf("titre = %v, want Nouveau titre", service["titre"])
	}
}

func TestDeleteService_RejectsNonOwnerThenAllowsOwner(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, _ := apiCall(t, srv, http.MethodDelete, "/api/services/"+serviceID, alice, nil)
	if status != http.StatusForbidden {
		t.Errorf("status = %d, want %d (non propriétaire)", status, http.StatusForbidden)
	}

	status, _ = apiCall(t, srv, http.MethodDelete, "/api/services/"+serviceID, bob, nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d (propriétaire)", status, http.StatusOK)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/services/"+serviceID, "", nil)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d (service supprimé)", status, http.StatusNotFound)
	}
}

func TestGetServiceReviews_Empty(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, _ := apiCall(t, srv, http.MethodGet, "/api/services/"+serviceID+"/reviews", "", nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
}

func TestListExchanges_FiltersByStatus(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	_, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	exchangeID := idOf(t, body, "exchange_id")

	status, _ := apiCall(t, srv, http.MethodGet, "/api/exchanges", alice, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/exchanges?status=pending", alice, nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d (filtre status=pending)", status, http.StatusOK)
	}

	status, exchange := apiCall(t, srv, http.MethodGet, "/api/exchanges/"+exchangeID, alice, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%v", status, http.StatusOK, exchange)
	}
	if exchange["status"] != "pending" {
		t.Errorf("status métier = %v, want pending", exchange["status"])
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/api/exchanges/"+exchangeID, "999", nil)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d (utilisateur non impliqué)", status, http.StatusNotFound)
	}
}

func TestRejectExchange(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	_, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	exchangeID := idOf(t, body, "exchange_id")

	status, _ := apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/reject", bob, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	// Un échange déjà refusé ne peut pas être re-refusé.
	status, _ = apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/reject", bob, nil)
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (déjà refusé)", status, http.StatusBadRequest)
	}
}

func TestSwaggerAndOpenAPIEndpoints(t *testing.T) {
	srv := testServer(t)

	status, _ := apiCall(t, srv, http.MethodGet, "/docs", "", nil)
	if status != http.StatusOK {
		t.Errorf("GET /docs status = %d, want %d", status, http.StatusOK)
	}

	status, _ = apiCall(t, srv, http.MethodGet, "/openapi.json", "", nil)
	if status != http.StatusOK {
		t.Errorf("GET /openapi.json status = %d, want %d", status, http.StatusOK)
	}
}
