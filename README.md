# GoLendar

GoLendar est une API RESTful modulaire écrite en Go pour la gestion de calendriers, d'événements, d'utilisateurs et de leurs relations. Le projet met l'accent sur la propreté, la sécurité, la maintenabilité et la traçabilité du code.

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

## ⚙️ Installation & Lancement

1. **Cloner le repo**
   ```bash
   git clone <repo-url>
   cd GoLendar
   ```
2. **Configurer la base de données**
   - Modifier les variables d'environnement si besoin (voir `internal/common/config.go`)
   - Lancer la stack de dev :
     ```bash
     docker-compose up -d
     ```
   - Importer le schéma SQL :
     ```bash
     mysql -u root -p calendar < resources/schema.sql
     ```
3. **Installer les dépendances Go**
   ```bash
   go mod download
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