# BarterSwap — API d'échange de compétences

BarterSwap est une API REST écrite en Go qui permet à des utilisateurs d'échanger des services entre eux (cours de guitare contre un coup de main en bricolage, par exemple) grâce à un système de crédits virtuels plutôt que d'argent réel.

Chaque utilisateur démarre avec **10 crédits**, publie des **services** dans une catégorie donnée, et peut lancer une **demande d'échange** sur le service d'un autre utilisateur. Une fois l'échange terminé, les deux parties peuvent se laisser un **avis**.


## Stack technique

- **Go** 1.25
- **MySQL** 8.0 (driver [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))
- **Docker** / **Docker Compose** pour l'environnement de dev

## Modèle de données

| Table | Description |
|---|---|
| `users` | Pseudo (unique), bio, ville, solde de crédits (10 par défaut) |
| `skills` | Compétences d'un utilisateur (nom + niveau) |
| `services` | Service proposé par un utilisateur : titre, description, catégorie, durée, coût en crédits, ville, actif/inactif |
| `exchanges` | Demande d'échange sur un service : statut (`pending`, `accepted`, `rejected`, `completed`, `cancelled`) |
| `credit_transactions` | Historique des mouvements de crédits (`spend`, `earn`, `refund`) liés à un échange |
| `reviews` | Avis laissé après un échange terminé (note de 1 à 5 + commentaire) |

Les tables sont créées automatiquement au démarrage de l'application si elles n'existent pas.

**Catégories de services valides :** `Informatique`, `Jardinage`, `Bricolage`, `Cuisine`, `Musique`, `Langues`, `Sport`, `Tutorat`, `Déménagement`, `Photographie`, `Animalier`, `Couture`, `Autre`.

## Installation

### Avec Docker 

```bash
git clone https://github.com/x225franc/Projet_Go_BarterSwap.git
cd Projet_Go_BarterSwap
docker-compose up --build
```

L'API est alors disponible sur `http://localhost:8080`, et MySQL sur le port `3306`.

### En local

Il faut une instance MySQL accessible avec un utilisateur `barteruser` / mot de passe `root` et une base `barterswap_db`.

```bash
git clone https://github.com/x225franc/Projet_Go_BarterSwap.git
cd Projet_Go_BarterSwap
go mod tidy
go run main.go
```

### Variable d'environnement

| Variable | Description | Défaut |
|---|---|---|
| `DB_HOST` | Hôte du serveur MySQL | `localhost` |

## Authentification

Il n'y a pas de système de token : les routes protégées vérifient simplement l'en-tête `X-User-ID`, qui doit correspondre à l'ID de l'utilisateur effectuant l'action (propriétaire du profil, du service ou de l'échange concerné).

```
X-User-ID: 1
```

## Endpoints

### Utilisateurs

```
en cours...
```

### Services

```
en cours...
```

### Échanges

```
en cours...
```

## Exemples d'utilisation

```
en cours...
```

## Tests

```
en cours...
```