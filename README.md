<div align="center" vertical-align="center">
  <img src="assets/GoLendar-Logo.png" alt="GoLendar Logo" width="240"/>
</div>

# GoLendar

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=herradiamine_GoLendar&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=herradiamine_GoLendar)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=herradiamine_GoLendar&metric=coverage)](https://sonarcloud.io/summary/new_code?id=herradiamine_GoLendar)

GoLendar API est une API RESTful modulaire écrite en Go pour la gestion de calendriers, d'événements, d'utilisateurs et de leurs relations. Le projet met l'accent sur la propreté, la sécurité, la maintenabilité et la traçabilité du code.

---

## 📖 Présentation

- **Langage** : Go (>= 1.24)
- **Base de données** : MySQL 8
- **Architecture** : RESTful, modulaire, testée
- **Déploiement** : Docker multi-stage, docker-compose

---

## 📂 Structure du projet

```
GoLendar/
├── assets/                # Fichiers statiques (logo, ...)
├── cmd/app/main.go        # Point d'entrée de l'application
├── internal/
│   ├── calendar/          # Logique métier des calendriers
│   ├── calendar_event/    # Logique métier des événements
│   ├── common/            # Config, modèles, erreurs, logger, ...
│   ├── middleware/        # Middlewares Gin (auth, validation, ...)
│   ├── routes/            # Définition des routes API
│   ├── user/              # Logique métier des utilisateurs
│   └── user_calendar/     # Gestion des liens utilisateur/calendrier
├── resources/             # Schéma SQL, collection Postman
├── scripts/               # Scripts utilitaires (build, ...)
├── testutils/             # Utilitaires de test
├── Dockerfile             # Build multi-stage Go
├── docker-compose.yml     # Stack complète (API + DB)
├── .dockerignore          # Exclusions Docker
├── go.mod / go.sum        # Dépendances Go
└── README.md              # Documentation (ce fichier)
```

---

## 🛠️ Bonnes pratiques appliquées

- **Séparation claire des responsabilités** (dossier `internal/`)
- **Centralisation des messages d'erreur** (`internal/common/messages.go`)
- **Validation stricte des entrées** (middlewares, modèles)
- **Logging structuré** (slog)
- **Tests unitaires et d'intégration** (dossiers `*_test.go`)
- **Configuration par variables d'environnement** (DB, ports, etc.)
- **Build Docker multi-stage** (image légère et sécurisée)
- **Healthcheck API** (endpoint `/health`)
- **Documentation et exemples d'utilisation** (README, Postman)

---

## 🚀 Installation & Lancement

### 1. Prérequis

- [Go 1.24+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) & [docker-compose](https://docs.docker.com/compose/)
- [Git](https://git-scm.com/)

### 2. Cloner le projet

```bash
git clone <url-du-repo>
cd GoLendar
```

### 3. Lancer avec Docker (recommandé)

```bash
./scripts/build.sh
```
Ce script :
- build l'image Docker
- arrête les anciens conteneurs
- lance la stack complète (API + MySQL)
- vérifie la santé de l'API

Vérifier l'API :
```bash
curl http://localhost:8080/health
# {"service":"GoLendar API","status":"healthy"}
```

#### Commandes utiles

- Voir les logs :  
  `docker-compose logs -f golendar_app`
- Arrêter la stack :  
  `docker-compose down`
- Redémarrer :  
  `docker-compose restart`

### 4. Lancer en local (développement)

1. Installer les dépendances :
   ```bash
   go mod download
   ```
2. Configurer la base de données (MySQL doit tourner sur `localhost:3306` ou adapter les variables d'environnement).
3. Lancer l'API :
   ```bash
   go run cmd/app/main.go
   ```

---

## 📚 Fonctionnalités principales

- CRUD utilisateurs, calendriers, événements
- Association utilisateurs/calendriers
- Filtrage des événements par jour, semaine, mois (routes RESTful)
- Gestion centralisée des erreurs
- Middleware d'authentification et de validation
- Logging structuré
- Tests unitaires et d'intégration
- Build Docker multi-stage optimisé
- Healthcheck API

---

## ✨ Pour aller plus loin

- Adapter la configuration dans `internal/common/config.go` si besoin
- Ajouter des tests ou des routes selon vos besoins
- Utiliser la collection Postman (`resources/postman_collection.json`) pour explorer l'API

---

**Auteur** : Amine HERRADI
**Licence** : MIT 