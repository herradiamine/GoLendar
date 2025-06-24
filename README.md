<div align="center" vertical-align="center">
  <img src="assets/GoLendar-Logo.png" alt="GoLendar Logo" width="240"/>
</div>

GoLendar API est une API RESTful modulaire écrite en Go pour la gestion de calendriers, d'événements, d'utilisateurs et de leurs relations. Le projet met l'accent sur la propreté, la sécurité, la maintenabilité et la traçabilité du code.

## 🚀 Fonctionnalités principales

- Gestion des utilisateurs, calendriers, événements et liaisons utilisateur-calendrier
- API RESTful respectant les bonnes pratiques (IDs dans l'URL, statuts HTTP, etc.)
- Middlewares robustes pour la sécurité et la gestion des permissions
- Centralisation des messages d'erreur et de succès
- Logging structuré (JSON, production-ready) avec `log/slog`
- Tests unitaires et d'intégration exhaustifs
- Pipeline CI/CD GitHub Actions (lint, build, tests, logs)
- Helpers mutualisés pour la manipulation des pointeurs
- Collection Postman à jour pour tester toute l'API

## 🗂️ Structure du projet

```
GoLendar/
├── cmd/app/main.go           # Point d'entrée de l'application
├── internal/
│   ├── user/                 # Logique métier utilisateur
│   ├── calendar/             # Logique métier calendrier
│   ├── calendar_event/       # Logique métier événement
│   ├── user_calendar/        # Liaisons utilisateur-calendrier
│   ├── common/               # Config, modèles, helpers, messages, logger
│   ├── middleware/           # Middlewares Gin
│   └── routes/               # Définition des routes
├── testutils/                # Helpers pour les tests
├── resources/
│   ├── postman_collection.json
│   └── schema.sql            # Schéma SQL de la base
├── .github/workflows/ci.yml  # Pipeline CI/CD
├── docker-compose.yml        # Stack de dev (MySQL, etc.)
├── go.mod / go.sum           # Dépendances Go
└── logs/                     # Logs applicatifs (ignorés par git)
```

## 🚀 Démarrage Rapide

### Avec Docker (Recommandé)

1. **Cloner le repository**
   ```bash
   git clone <repository-url>
   cd GoLendar
   ```

2. **Build et démarrer avec Docker Compose**
   ```bash
   # Build de l'image
   ./scripts/build.sh
   
   # Ou directement avec docker-compose
   docker-compose up --build -d
   ```

3. **Vérifier que l'application fonctionne**
   ```bash
   # Vérifier le statut des conteneurs
   docker-compose ps
   
   # Voir les logs
   docker-compose logs -f golendar
   
   # Tester l'endpoint de santé
   curl http://localhost:8080/health
   ```

4. **Arrêter l'application**
   ```bash
   docker-compose down
   ```

### Commandes Docker utiles

```bash
# Build de l'image
docker build -t golendar .

# Démarrer les services
docker-compose up -d

# Voir les logs en temps réel
docker-compose logs -f golendar

# Arrêter les services
docker-compose down

# Nettoyer les conteneurs et images
docker-compose down --remove-orphans
docker system prune -f
```

### Développement Local

1. **Prérequis**
   - Go 1.24+
   - MySQL 8.0+
   - Git

2. **Installation**
   ```bash
   git clone <repository-url>
   cd GoLendar
   go mod download
   ```

3. **Configuration de la base de données**
   ```bash
   # Démarrer MySQL avec Docker
   docker-compose up mysql -d
   
   # Ou configurer votre propre instance MySQL
   ```

4. **Lancer l'application**
   ```bash
   go run cmd/app/main.go
   ```

## 🧪 Tests

- Lancer tous les tests :
  ```bash
  go test ./... -v
  ```
- Les tests utilisent une base dédiée (voir `testutils/testutils.go`).

## 🔎 CI/CD

- Pipeline GitHub Actions : lint (`go vet`, `staticcheck`), build, tests, upload des logs en cas d'échec
- Version de Go alignée sur l'environnement de dev

## 🔐 Sécurité & Qualité

- Middlewares pour la vérification d'existence et de permissions
- Gestion centralisée des erreurs et des logs
- Respect des standards Go et des bonnes pratiques REST

## 📚 Documentation API

- Voir la collection Postman : `resources/postman_collection.json`
- Endpoints principaux :
  - `/user`, `/calendar`, `/calendar-event`, `/user-calendar`
  - Tous les endpoints utilisent des IDs en path et des statuts HTTP explicites

### Endpoints User Calendar

- `GET /user-calendar/:user_id` - Liste tous les calendriers d'un utilisateur avec leurs détails
- `GET /user-calendar/:user_id/:calendar_id` - Récupère une liaison spécifique utilisateur-calendrier
- `POST /user-calendar/:user_id/:calendar_id` - Crée une nouvelle liaison utilisateur-calendrier
- `PUT /user-calendar/:user_id/:calendar_id` - Met à jour une liaison utilisateur-calendrier
- `DELETE /user-calendar/:user_id/:calendar_id` - Supprime une liaison utilisateur-calendrier

### Endpoints Calendar Events

- `GET /calendar-event/:user_id/:calendar_id/:event_id` - Récupère un événement spécifique
- `POST /calendar-event/:user_id/:calendar_id` - Crée un nouvel événement
- `PUT /calendar-event/:user_id/:calendar_id/:event_id` - Met à jour un événement
- `DELETE /calendar-event/:user_id/:calendar_id/:event_id` - Supprime un événement

#### Récupération des événements avec filtres temporels

- `GET /calendar-event/:user_id/:calendar_id/month/:year/:month` - Liste tous les événements d'un mois
- `GET /calendar-event/:user_id/:calendar_id/week/:year/:week` - Liste tous les événements d'une semaine ISO
- `GET /calendar-event/:user_id/:calendar_id/day/:year/:month/:day` - Liste tous les événements d'un jour

**Exemples d'utilisation :**
- Tous les événements de janvier 2024 : `/calendar-event/1/1/month/2024/1`
- Tous les événements de la semaine 3 de 2024 : `/calendar-event/1/1/week/2024/3`
- Tous les événements du 15 janvier 2024 : `/calendar-event/1/1/day/2024/1/15`

### Exemple de réponse pour la liste des calendriers

```json
{
  "success": true,
  "message": "Liste des calendriers récupérée avec succès",
  "data": [
    {
      "user_calendar_id": 1,
      "user_id": 1,
      "calendar_id": 1,
      "title": "Calendrier Personnel",
      "description": "Mon calendrier personnel pour les événements privés",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": null,
      "deleted_at": null
    },
    {
      "user_calendar_id": 2,
      "user_id": 1,
      "calendar_id": 2,
      "title": "Calendrier Professionnel",
      "description": "Calendrier pour les réunions et événements professionnels",
      "created_at": "2024-01-02T14:30:00Z",
      "updated_at": null,
      "deleted_at": null
    }
  ]
}
```

### Exemple de réponse pour la liste des événements

```json
{
  "success": true,
  "message": "Liste des événements récupérée avec succès",
  "data": [
    {
      "event_id": 1,
      "title": "Réunion équipe",
      "description": "Réunion hebdomadaire de l'équipe de développement",
      "start": "2024-01-15T10:00:00Z",
      "duration": 60,
      "canceled": false,
      "created_at": "2024-01-10T09:00:00Z",
      "updated_at": null,
      "deleted_at": null
    },
    {
      "event_id": 2,
      "title": "Déjeuner client",
      "description": "Déjeuner avec le client pour discuter du projet",
      "start": "2024-01-15T12:30:00Z",
      "duration": 90,
      "canceled": false,
      "created_at": "2024-01-12T14:00:00Z",
      "updated_at": null,
      "deleted_at": null
    }
  ]
}
```

## 📝 Helpers utiles

- `common.StringPtr`, `common.IntPtr`, `common.BoolPtr` pour manipuler facilement les pointeurs dans les tests et la logique métier

## 🛠️ Contribution

1. Forkez le repo
2. Créez une branche (`feature/ma-feature`)
3. Commitez vos modifications
4. Ouvrez une Pull Request

## 📄 Licence

MIT

---

**Auteur :** Amine Herradi

Pour toute question ou suggestion, ouvrez une issue ou contactez-moi !

## Fonctionnalités

### Gestion des utilisateurs
- ✅ CRUD complet (Create, Read, Update, Delete)
- ✅ Validation des données (email, mot de passe)
- ✅ Gestion des erreurs centralisée

### Gestion des calendriers
- ✅ CRUD complet (Create, Read, Update, Delete)
- ✅ Association avec les utilisateurs
- ✅ Gestion des erreurs centralisée

### Gestion des événements
- ✅ CRUD complet (Create, Read, Update, Delete)
- ✅ Filtrage par période (jour, semaine, mois)
- ✅ Routes RESTful explicites
- ✅ Gestion des erreurs centralisée

### Gestion des relations utilisateur-calendrier
- ✅ CRUD complet des liaisons
- ✅ Liste des calendriers d'un utilisateur
- ✅ Gestion des erreurs centralisée

### Architecture et bonnes pratiques
- ✅ Structure modulaire et extensible
- ✅ Tests unitaires et d'intégration complets
- ✅ Logging structuré avec slog
- ✅ Gestion centralisée des messages d'erreur
- ✅ Validation des données avec go-playground/validator
- ✅ Middleware pour la vérification d'accès
- ✅ Documentation complète avec exemples
- ✅ Containerisation Docker avec build multi-stage
- ✅ Health checks et monitoring 