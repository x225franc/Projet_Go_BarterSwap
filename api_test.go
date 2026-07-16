package main


import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "3306")
	}
	connectToDB()
	createTables()
	os.Exit(m.Run())
}

func resetDB(t *testing.T) {
	t.Helper()
	tables := []string{"reviews", "credit_transactions", "exchanges", "services", "skills", "users"}
	for _, tbl := range tables {
		if _, err := db.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("reset table %s: %v", tbl, err)
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", tbl)); err != nil {
			t.Fatalf("reset auto_increment %s: %v", tbl, err)
		}
	}
}

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	resetDB(t)
	srv := httptest.NewServer(newRouter())
	t.Cleanup(srv.Close)
	return srv
}

func apiCall(t *testing.T, srv *httptest.Server, method, path, userID string, body any) (int, map[string]any) {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	var parsed map[string]any
	json.NewDecoder(resp.Body).Decode(&parsed)
	return resp.StatusCode, parsed
}

func idOf(t *testing.T, body map[string]any, field string) string {
	t.Helper()
	v, ok := body[field]
	if !ok {
		t.Fatalf("champ %q absent de la réponse: %v", field, body)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("champ %q n'est pas numérique: %v", field, v)
	}
	return fmt.Sprintf("%d", int(f))
}

func apiCreateUser(t *testing.T, srv *httptest.Server, pseudo string) string {
	t.Helper()
	status, body := apiCall(t, srv, http.MethodPost, "/api/users", "", map[string]any{
		"pseudo": pseudo, "ville": "Lyon",
	})
	if status != http.StatusCreated {
		t.Fatalf("création utilisateur %q: status = %d, body = %v", pseudo, status, body)
	}
	return idOf(t, body, "id")
}

func apiCreateUserWithSkill(t *testing.T, srv *httptest.Server, pseudo, skillNom string) string {
	t.Helper()
	id := apiCreateUser(t, srv, pseudo)
	status, body := apiCall(t, srv, http.MethodPut, "/api/users/"+id+"/skills", id, []map[string]string{
		{"nom": skillNom, "niveau": "expert"},
	})
	if status != http.StatusOK {
		t.Fatalf("attribution compétence à %q: status = %d, body = %v", pseudo, status, body)
	}
	return id
}

func apiCreateService(t *testing.T, srv *httptest.Server, providerID, categorie string, credits int) string {
	t.Helper()
	status, body := apiCall(t, srv, http.MethodPost, "/api/services", providerID, map[string]any{
		"titre": "Service de test", "categorie": categorie,
		"duree_minutes": 60, "credits": credits, "ville": "Lyon",
	})
	if status != http.StatusCreated {
		t.Fatalf("création service: status = %d, body = %v", status, body)
	}
	return idOf(t, body, "id")
}

func TestCreateUser_Success(t *testing.T) {
	srv := testServer(t)

	status, body := apiCall(t, srv, http.MethodPost, "/api/users", "", map[string]any{"pseudo": "alice"})

	if status != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%v)", status, http.StatusCreated, body)
	}
	if body["pseudo"] != "alice" {
		t.Errorf("pseudo = %v, want alice", body["pseudo"])
	}
	if body["credit_balance"].(float64) != 10 {
		t.Errorf("credit_balance = %v, want 10 crédits de bienvenue", body["credit_balance"])
	}
}

func TestCreateUser_EmptyPseudo(t *testing.T) {
	srv := testServer(t)

	status, _ := apiCall(t, srv, http.MethodPost, "/api/users", "", map[string]any{"pseudo": ""})

	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestCreateUser_DuplicatePseudo(t *testing.T) {
	srv := testServer(t)

	apiCreateUser(t, srv, "alice")
	status, _ := apiCall(t, srv, http.MethodPost, "/api/users", "", map[string]any{"pseudo": "alice"})

	if status != http.StatusConflict {
		t.Errorf("status = %d, want %d (pseudo déjà pris)", status, http.StatusConflict)
	}
}

func TestUpdateUser_RequiresOwnership(t *testing.T) {
	srv := testServer(t)
	alice := apiCreateUser(t, srv, "alice")

	status, _ := apiCall(t, srv, http.MethodPut, "/api/users/"+alice, "999", map[string]any{"bio": "hack"})

	if status != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestCreateService_RejectsUnownedSkill(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Jardinage")

	status, body := apiCall(t, srv, http.MethodPost, "/api/services", bob, map[string]any{
		"titre": "Montage meuble", "categorie": "Bricolage",
		"duree_minutes": 60, "credits": 5,
	})

	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (compétence non possédée), body=%v", status, http.StatusBadRequest, body)
	}
}

func TestCreateService_AcceptsOwnedSkill(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")

	status, body := apiCall(t, srv, http.MethodPost, "/api/services", bob, map[string]any{
		"titre": "Montage meuble", "categorie": "Bricolage",
		"duree_minutes": 60, "credits": 5,
	})

	if status != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%v", status, http.StatusCreated, body)
	}
}

func TestCreateExchange_RejectsSelfExchange(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, _ := apiCall(t, srv, http.MethodPost, "/api/exchanges", bob, map[string]any{"service_id": mustAtoi(t, serviceID)})

	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (auto-échange)", status, http.StatusBadRequest)
	}
}

func TestCreateExchange_RejectsInsufficientCredits(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 15) // > 10 crédits de bienvenue

	status, _ := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})

	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (crédits insuffisants)", status, http.StatusBadRequest)
	}
}

func TestCreateExchange_RejectsAlreadyReservedService(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	charlie := apiCreateUser(t, srv, "charlie")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 5)

	status, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	if status != http.StatusCreated {
		t.Fatalf("première demande: status = %d, body = %v", status, body)
	}

	status, _ = apiCall(t, srv, http.MethodPost, "/api/exchanges", charlie, map[string]any{"service_id": mustAtoi(t, serviceID)})
	if status != http.StatusConflict {
		t.Errorf("status = %d, want %d (service déjà réservé)", status, http.StatusConflict)
	}
}

func TestExchangeFullLifecycle(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	status, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	if status != http.StatusCreated {
		t.Fatalf("création échange: status = %d, body = %v", status, body)
	}
	exchangeID := idOf(t, body, "exchange_id")

	// Acceptation : les crédits d'alice sont bloqués.
	status, body = apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/accept", bob, nil)
	if status != http.StatusOK {
		t.Fatalf("accept: status = %d, body = %v", status, body)
	}
	_, aliceBody := apiCall(t, srv, http.MethodGet, "/api/users/"+alice, "", nil)
	if got := aliceBody["credit_balance"].(float64); got != 6 {
		t.Errorf("crédits alice après acceptation = %v, want 6 (10 - 4)", got)
	}

	// Complétion : les crédits sont transférés à bob.
	status, body = apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/complete", alice, nil)
	if status != http.StatusOK {
		t.Fatalf("complete: status = %d, body = %v", status, body)
	}
	_, bobBody := apiCall(t, srv, http.MethodGet, "/api/users/"+bob, "", nil)
	if got := bobBody["credit_balance"].(float64); got != 14 {
		t.Errorf("crédits bob après complétion = %v, want 14 (10 + 4)", got)
	}

	// Notation par le demandeur : cible = bob.
	status, _ = apiCall(t, srv, http.MethodPost, "/api/exchanges/"+exchangeID+"/review", alice, map[string]any{
		"note": 5, "commentaire": "Impeccable",
	})
	if status != http.StatusCreated {
		t.Fatalf("review: status = %d", status)
	}

	// Stats de bob : 1 échange complété, note moyenne 5, 4 crédits gagnés.
	status, stats := apiCall(t, srv, http.MethodGet, "/api/users/"+bob+"/stats", "", nil)
	if status != http.StatusOK {
		t.Fatalf("stats: status = %d, body = %v", status, stats)
	}
	if got := stats["echanges_completes"].(float64); got != 1 {
		t.Errorf("echanges_completes = %v, want 1", got)
	}
	if got := stats["note_moyenne"].(float64); got != 5 {
		t.Errorf("note_moyenne = %v, want 5", got)
	}
	if got := stats["total_gagne"].(float64); got != 4 {
		t.Errorf("total_gagne = %v, want 4", got)
	}
	if got := stats["credit_balance"].(float64); got != 14 {
		t.Errorf("credit_balance (stats) = %v, want 14", got)
	}
}

func TestCancelAcceptedExchange_RefundsCredits(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	_, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	exchangeID := idOf(t, body, "exchange_id")

	apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/accept", bob, nil)

	status, _ := apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/cancel", alice, nil)
	if status != http.StatusOK {
		t.Fatalf("cancel: status = %d", status)
	}

	_, aliceBody := apiCall(t, srv, http.MethodGet, "/api/users/"+alice, "", nil)
	if got := aliceBody["credit_balance"].(float64); got != 10 {
		t.Errorf("crédits alice après annulation = %v, want 10 (remboursés)", got)
	}
}

func TestReview_RejectsNonCompletedExchange(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	_, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	exchangeID := idOf(t, body, "exchange_id")

	status, _ := apiCall(t, srv, http.MethodPost, "/api/exchanges/"+exchangeID+"/review", alice, map[string]any{"note": 5})
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (échange pas terminé)", status, http.StatusBadRequest)
	}
}

func TestReview_RejectsDuplicateReview(t *testing.T) {
	srv := testServer(t)
	bob := apiCreateUserWithSkill(t, srv, "bob", "Bricolage")
	alice := apiCreateUser(t, srv, "alice")
	serviceID := apiCreateService(t, srv, bob, "Bricolage", 4)

	_, body := apiCall(t, srv, http.MethodPost, "/api/exchanges", alice, map[string]any{"service_id": mustAtoi(t, serviceID)})
	exchangeID := idOf(t, body, "exchange_id")
	apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/accept", bob, nil)
	apiCall(t, srv, http.MethodPut, "/api/exchanges/"+exchangeID+"/complete", alice, nil)

	status, _ := apiCall(t, srv, http.MethodPost, "/api/exchanges/"+exchangeID+"/review", alice, map[string]any{"note": 5})
	if status != http.StatusCreated {
		t.Fatalf("premier avis: status = %d", status)
	}

	status, _ = apiCall(t, srv, http.MethodPost, "/api/exchanges/"+exchangeID+"/review", alice, map[string]any{"note": 3})
	if status != http.StatusConflict {
		t.Errorf("status = %d, want %d (avis déjà déposé)", status, http.StatusConflict)
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		t.Fatalf("mustAtoi(%q): %v", s, err)
	}
	return n
}
