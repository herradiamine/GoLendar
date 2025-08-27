# Spécifications Frontend – GoLendar

## Table des matières
- [Introduction](#introduction)
- [Principes Généraux](#principes-généraux)
- [Architecture des Pages/Modules](#architecture-des-pagesmodules)
- [Spécifications par Module](#spécifications-par-module)
  - [1. Authentification & Sessions](#1-authentification--sessions)
  - [2. Utilisateurs](#2-utilisateurs)
  - [3. Rôles (Admin)](#3-rôles-admin)
  - [4. Calendriers](#4-calendriers)
  - [5. Événements](#5-événements)
  - [6. Permissions & Sécurité](#6-permissions--sécurité)
- [Gestion des Erreurs & Notifications](#gestion-des-erreurs--notifications)
- [Recommandations Techniques](#recommandations-techniques)

---

## Introduction
Ce document décrit les spécifications fonctionnelles et techniques pour le développement d'une interface Frontend qui communique avec l'API GoLendar.

## Principes Généraux
- **API RESTful** (voir `API_ROUTES.md`)
- **Authentification JWT** (token + refresh)
- **Gestion des rôles et permissions**
- **UX fluide** : feedback utilisateur, gestion des erreurs, navigation claire
- **Sécurité** : stockage sécurisé des tokens, contrôle d'accès côté Front

## Architecture des Pages/Modules
- Page de connexion
- Page d'inscription
- Tableau de bord (accueil)
- Gestion du profil utilisateur
- Gestion des sessions
- Liste des calendriers
- Détail d'un calendrier (événements, membres, permissions)
- Vue calendrier (mois/semaine/jour)
- Détail/création/édition d'événement
- (Admin) Gestion des utilisateurs
- (Admin) Gestion des rôles

## Spécifications par Module

### 1. Authentification & Sessions
#### Pages/Composants
- **Connexion** : formulaire email/mot de passe
- **Inscription** : formulaire complet
- **Gestion des sessions** : liste, suppression (avec localisation)
- **Profil connecté** : infos utilisateur, rôles

#### Flux API
- `POST /auth/login` → récupère `session_token`, `refresh_token`, infos utilisateur
- `POST /auth/refresh` → renouvelle le token
- `POST /auth/logout` → déconnexion
- `GET /auth/me` → profil utilisateur connecté
- `GET /auth/sessions` → liste des sessions actives (avec localisation)
- `DELETE /auth/sessions/:session_id` → suppression d'une session

#### Exemples de payloads
```json
// Connexion
{
  "email": "user@example.com",
  "password": "password123"
}

// Réponse succès
{
  "success": true,
  "data": {
    "user": {
      "user_id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-01T01:00:00Z",
    "roles": ["user"]
  }
}

// Réponse liste des sessions (avec localisation géographique)
{
  "success": true,
  "data": [
    {
      "user_session_id": 1,
      "user_id": 1,
      "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_at": "2024-01-01T01:00:00Z",
      "device_info": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "ip_address": "192.168.1.100",
      "location": "Paris, Île-de-France, France",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 2. Utilisateurs
#### Pages/Composants
- **Inscription** : formulaire
- **Profil** : affichage, édition, suppression
- **(Admin)** : liste, fiche, édition, suppression, gestion des rôles

#### Flux API
- `POST /user` → inscription
- `GET /user/me` → profil connecté
- `PUT /user/me` → édition profil
- `DELETE /user/me` → suppression compte
- (Admin) `GET/PUT/DELETE /user/:user_id`, `GET /user/:user_id/with-roles`

#### Exemples de payloads
```json
// Inscription
{
  "lastname": "Dupont",
  "firstname": "Jean",
  "email": "jean@example.co",
  "password": "password123"
}

// Réponse profil
{
  "success": true,
  "data": {
    "user_id": 1,
    "lastname": "Dupont",
    "firstname": "Jean",
    "email": "jean@example.com",
    ...
  }
}
```

### 3. Rôles (Admin)
#### Pages/Composants
- **Liste des rôles**
- **Création/édition/suppression**
- **Attribution/révocation à un utilisateur**

#### Flux API
- `GET /roles` → liste
- `POST /roles` → création
- `PUT /roles/:id` → édition
- `DELETE /roles/:id` → suppression
- `POST /roles/assign` / `POST /roles/revoke` → gestion des rôles utilisateur
- `GET /roles/user/:user_id` → rôles d'un utilisateur

#### Exemples de payloads
```json
// Création rôle
{
  "name": "moderator",
  "description": "Rôle de modérateur"
}
```

### 4. Calendriers
#### Pages/Composants
- **Liste des calendriers accessibles**
- **Liste de mes calendriers** (accessible à tout utilisateur connecté)
- **Création/édition/suppression**
- **Détail d'un calendrier** (événements, membres, permissions)
- **Gestion des accès** (admin)

#### Flux API
- `GET /user-calendar/me` → liste des calendriers de l'utilisateur connecté *(à implémenter côté API si non existant, sinon utiliser la route existante pour lister les accès de l'utilisateur courant)*
- `POST /calendar` → création
- `GET/PUT/DELETE /calendar/:calendar_id` → gestion
- (Admin) `GET/POST/PUT/DELETE /user-calendar/:user_id/:calendar_id` → gestion des accès

#### Exemples de payloads
```json
// Réponse liste de mes calendriers
{
  "success": true,
  "data": [
    {
      "calendar_id": 1,
      "title": "Mon Calendrier",
      "description": "...",
      ...
    },
    {
      "calendar_id": 2,
      "title": "Projets",
      ...
    }
  ]
}
```

### 5. Événements
#### Pages/Composants
- **Vue calendrier** (mois/semaine/jour)
- **Création/édition/suppression**
- **Détail d'un événement**

#### Flux API
- `POST /calendar-event/:calendar_id` → création
- `GET/PUT/DELETE /calendar-event/:calendar_id/:event_id` → gestion
- `GET /calendar-event/:calendar_id/month/:year/:month` → liste par mois
- `GET /calendar-event/:calendar_id/week/:year/:week` → liste par semaine
- `GET /calendar-event/:calendar_id/day/:year/:month/:day` → liste par jour

#### Exemples de payloads
```json
// Création événement
{
  "title": "Réunion",
  "description": "Réunion d'équipe",
  "start": "2025-01-15T10:00:00Z",
  "duration": 60,
  "calendar_id": 1
}

// Réponse événement
{
  "success": true,
  "data": {
    "event_id": 1,
    "title": "Réunion",
    "start": "2025-01-15T10:00:00Z",
    "duration": 60,
    ...
  }
}
```

### 6. Permissions & Sécurité
- **Contrôle d'accès** :
  - Affichage conditionnel des actions selon le rôle (`user`, `admin`, `moderator`) et la permission sur le calendrier (`read_only`, `read_write`, `admin`)
  - Redirection ou message d'erreur si accès refusé
- **Gestion des tokens** :
  - Stockage sécurisé (ex: localStorage chiffré, cookies HttpOnly si possible)
  - Rafraîchissement automatique du token

## Gestion des Erreurs & Notifications
- Affichage des messages de succès/erreur issus des champs `message` ou `error` des réponses API
- Notifications visuelles (toast, alertes, etc.)
- Gestion des cas d'expiration de session, permissions insuffisantes, erreurs réseau

## Recommandations Techniques
- **Framework recommandé** : React, Vue ou Angular (au choix)
- **Gestion du state** : Redux, Pinia, Vuex, ou Context API
- **Gestion des appels API** : Axios ou Fetch, avec intercepteurs pour les tokens
- **Routing** : React Router, Vue Router, etc.
- **Sécurité** :
  - Ne jamais exposer les tokens dans l'URL
  - Vérifier les permissions côté Front avant chaque action sensible
- **UX** :
  - Feedback utilisateur systématique
  - Navigation fluide entre les modules
  - Accessibilité (labels, navigation clavier)

---

*Document généré automatiquement à partir de la documentation GoLendar (API_ROUTES.md, README.md, modèles Go, collection Postman)* 