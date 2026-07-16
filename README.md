# BarterSwap — API d'échange de compétences

## Installation

### Avec Docker

```bash
git clone https://github.com/x225franc/Projet_Go_BarterSwap.git
cd Projet_Go_BarterSwap
docker-compose up --build
```

L'API est alors disponible sur `http://localhost:8080`, et MySQL sur le port `3306`.

## Documentation interactive

Une spécification OpenAPI 3.0 est embarquée dans le binaire et exposée par l'API elle-même :

- `GET /openapi.json` — la spec brute
- `GET /docs` — une interface Swagger UI pour explorer et tester les endpoints depuis le navigateur

## Endpoints

### Utilisateurs

| Méthode | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/users` | — | Créer un compte (10 crédits offerts) |
| GET | `/api/users/{id}` | — | Profil public d'un utilisateur |
| PUT | `/api/users/{id}` | `X-User-ID` = id | Modifier son profil (bio, ville) |
| GET | `/api/users/{id}/skills` | — | Compétences d'un utilisateur |
| PUT | `/api/users/{id}/skills` | `X-User-ID` = id | Définir ses compétences (écrase les précédentes) |
| GET | `/api/users/{id}/reviews` | — | Avis reçus par un utilisateur |
| GET | `/api/users/{id}/stats` | — | Statistiques du tableau de bord |

### Services

| Méthode | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/services` | `X-User-ID` | Créer une annonce de service |
| GET | `/api/services` | — | Liste des services (`?categorie=`, `?ville=`, `?search=`) |
| GET | `/api/services/{id}` | — | Détail d'un service |
| PUT | `/api/services/{id}` | `X-User-ID` = provider | Modifier son annonce |
| DELETE | `/api/services/{id}` | `X-User-ID` = provider | Supprimer son annonce |
| GET | `/api/services/{id}/reviews` | — | Avis sur un service |

### Échanges

| Méthode | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/exchanges` | `X-User-ID` | Créer une demande d'échange |
| GET | `/api/exchanges` | `X-User-ID` | Échanges de l'utilisateur (`?status=`) |
| GET | `/api/exchanges/{id}` | `X-User-ID` | Détail d'un échange |
| PUT | `/api/exchanges/{id}/accept` | `X-User-ID` = owner | Accepter (bloque les crédits) |
| PUT | `/api/exchanges/{id}/reject` | `X-User-ID` = owner | Refuser |
| PUT | `/api/exchanges/{id}/complete` | `X-User-ID` = requester | Terminer (transfère les crédits) |
| PUT | `/api/exchanges/{id}/cancel` | `X-User-ID` | Annuler (rembourse si crédits bloqués) |
| POST | `/api/exchanges/{id}/review` | `X-User-ID` | Laisser un avis sur un échange terminé |

## Exemples d'utilisation

### Créer deux utilisateurs

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"pseudo": "alice", "ville": "Paris"}'

curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"pseudo": "bob", "ville": "Lyon"}'
```

### Publier un service (bob propose 1h de jardinage pour 3 crédits)

```bash
curl -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 2" \
  -d '{
        "titre": "Cours de jardinage",
        "categorie": "Jardinage",
        "duree_minutes": 60,
        "credits": 3,
        "ville": "Lyon"
      }'
```

### Demander, accepter puis terminer un échange (alice ↔ service de bob)

```bash
# alice (id 1) demande le service 1
curl -X POST http://localhost:8080/api/exchanges \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"service_id": 1}'

# bob (id 2, propriétaire) accepte -> 3 crédits bloqués chez alice
curl -X PUT http://localhost:8080/api/exchanges/1/accept -H "X-User-ID: 2"

# alice (demandeuse) marque l'échange comme terminé -> crédits transférés à bob
curl -X PUT http://localhost:8080/api/exchanges/1/complete -H "X-User-ID: 1"
```

### Laisser un avis puis consulter les statistiques

```bash
curl -X POST http://localhost:8080/api/exchanges/1/review \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"note": 5, "commentaire": "Super échange !"}'

curl http://localhost:8080/api/users/2/stats
```

## Tests

```bash
go test -v -cover ./...
```

Les tests unitaires (validations, mapping d'erreurs) ne nécessitent pas de base de données. Les tests d'API (`httptest`) exercent le cycle de vie complet d'un échange et nécessitent une instance MySQL disponible (`docker-compose up -d db` avant de lancer `go test`).
