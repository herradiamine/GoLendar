# API Routes - GoLendar

<div align="center">
  <img src="assets/GoLendar-Logo.png" alt="GoLendar Logo" width="200"/>
  <h2>📚 Documentation complète des routes API</h2>
  <p><em>Base URL: <code>http://localhost:8080</code></em></p>
</div>

---

## 📋 Table des matières

- [🔍 Health Check](#-health-check)
- [🔐 Authentification](#-authentification)
- [👥 Gestion des utilisateurs](#-gestion-des-utilisateurs)
- [🎭 Gestion des rôles](#-gestion-des-rôles)
- [🔗 Liaisons utilisateur-calendrier](#-liaisons-utilisateur-calendrier)
- [📅 Gestion des calendriers](#-gestion-des-calendriers)
- [📝 Gestion des événements](#-gestion-des-événements)
- [🔒 Niveaux d'autorisation](#-niveaux-dautorisation)

---

## 🔍 Health Check

### Vérification de l'état du service
- **URL** : `GET http://localhost:8080/health`
- **Description** : Endpoint public pour vérifier que l'API est opérationnelle
- **Réponse** : Statut du service et informations de base
- **Authentification** : ❌ Aucune requise

---

## 🔐 Authentification

### Routes publiques (inscription et connexion)

#### Connexion utilisateur
- **URL** : `POST http://localhost:8080/auth/login`
- **Description** : Authentification d'un utilisateur avec email et mot de passe
- **Corps** : `{"email": "user@example.com", "password": "password123"}`
- **Réponse** : Token de session et informations utilisateur
- **Authentification** : ❌ Aucune requise

#### Rafraîchissement de token
- **URL** : `POST http://localhost:8080/auth/refresh`
- **Description** : Renouvellement d'un token de session expiré
- **Corps** : `{"refresh_token": "token_value"}`
- **Réponse** : Nouveau token de session
- **Authentification** : ❌ Aucune requise

### Routes protégées (gestion des sessions)

#### Déconnexion utilisateur
- **URL** : `POST http://localhost:8080/auth/logout`
- **Description** : Déconnexion et invalidation de la session actuelle
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Confirmation de déconnexion
- **Authentification** : ✅ Token requis

#### Informations de l'utilisateur connecté
- **URL** : `GET http://localhost:8080/auth/me`
- **Description** : Récupération des informations de l'utilisateur authentifié avec ses rôles
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Profil utilisateur complet avec rôles
- **Authentification** : ✅ Token requis

#### Liste des sessions utilisateur
- **URL** : `GET http://localhost:8080/auth/sessions`
- **Description** : Récupération de toutes les sessions actives de l'utilisateur
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Liste des sessions avec leurs détails
- **Authentification** : ✅ Token requis

#### Suppression d'une session
- **URL** : `DELETE http://localhost:8080/auth/sessions/:session_id`
- **Description** : Suppression d'une session spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `session_id` - ID de la session à supprimer
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token requis

---

## 👥 Gestion des utilisateurs

### Routes publiques (inscription)

#### Création d'un nouvel utilisateur
- **URL** : `POST http://localhost:8080/user`
- **Description** : Inscription d'un nouvel utilisateur
- **Corps** : `{"lastname": "Dupont", "firstname": "Jean", "email": "jean@example.com", "password": "password123"}`
- **Réponse** : Confirmation de création avec ID utilisateur
- **Authentification** : ❌ Aucune requise

### Routes protégées (gestion du profil)

#### Récupération du profil utilisateur
- **URL** : `GET http://localhost:8080/user/me`
- **Description** : Récupération des informations du profil de l'utilisateur connecté
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Profil utilisateur
- **Authentification** : ✅ Token requis

#### Modification du profil utilisateur
- **URL** : `PUT http://localhost:8080/user/me`
- **Description** : Mise à jour des informations du profil utilisateur
- **Headers** : `Authorization: Bearer <token>`
- **Corps** : `{"lastname": "NouveauNom", "email": "nouveau@example.com"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token requis

#### Suppression du compte utilisateur
- **URL** : `DELETE http://localhost:8080/user/me`
- **Description** : Suppression du compte de l'utilisateur connecté (soft delete)
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token requis

### Routes admin (gestion de tous les utilisateurs)

#### Récupération d'un utilisateur par ID
- **URL** : `GET http://localhost:8080/user/:user_id`
- **Description** : Récupération des informations d'un utilisateur spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Réponse** : Profil utilisateur
- **Authentification** : ✅ Token + Rôle admin requis

#### Modification d'un utilisateur par ID
- **URL** : `PUT http://localhost:8080/user/:user_id`
- **Description** : Mise à jour des informations d'un utilisateur spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Corps** : `{"lastname": "NouveauNom", "email": "nouveau@example.com"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token + Rôle admin requis

#### Suppression d'un utilisateur par ID
- **URL** : `DELETE http://localhost:8080/user/:user_id`
- **Description** : Suppression d'un utilisateur spécifique (soft delete)
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token + Rôle admin requis

#### Récupération d'un utilisateur avec ses rôles
- **URL** : `GET http://localhost:8080/user/:user_id/with-roles`
- **Description** : Récupération d'un utilisateur avec la liste de ses rôles
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Réponse** : Profil utilisateur avec rôles
- **Authentification** : ✅ Token + Rôle admin requis

---

## 🎭 Gestion des rôles

### Routes admin (gestion des rôles)

#### Liste de tous les rôles
- **URL** : `GET http://localhost:8080/roles`
- **Description** : Récupération de la liste de tous les rôles disponibles
- **Headers** : `Authorization: Bearer <token>`
- **Réponse** : Liste des rôles avec leurs détails
- **Authentification** : ✅ Token + Rôle admin requis

#### Récupération d'un rôle par ID
- **URL** : `GET http://localhost:8080/roles/:id`
- **Description** : Récupération des détails d'un rôle spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `id` - ID du rôle
- **Réponse** : Détails du rôle
- **Authentification** : ✅ Token + Rôle admin requis

#### Création d'un nouveau rôle
- **URL** : `POST http://localhost:8080/roles`
- **Description** : Création d'un nouveau rôle dans le système
- **Headers** : `Authorization: Bearer <token>`
- **Corps** : `{"name": "moderator", "description": "Rôle de modérateur"}`
- **Réponse** : Confirmation de création avec ID du rôle
- **Authentification** : ✅ Token + Rôle admin requis

#### Modification d'un rôle
- **URL** : `PUT http://localhost:8080/roles/:id`
- **Description** : Mise à jour des informations d'un rôle existant
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `id` - ID du rôle
- **Corps** : `{"name": "moderator", "description": "Nouvelle description"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token + Rôle admin requis

#### Suppression d'un rôle
- **URL** : `DELETE http://localhost:8080/roles/:id`
- **Description** : Suppression d'un rôle du système
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `id` - ID du rôle
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token + Rôle admin requis

#### Attribution d'un rôle à un utilisateur
- **URL** : `POST http://localhost:8080/roles/assign`
- **Description** : Attribution d'un rôle à un utilisateur spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Corps** : `{"user_id": 123, "role_id": 456}`
- **Réponse** : Confirmation d'attribution
- **Authentification** : ✅ Token + Rôle admin requis

#### Révocation d'un rôle d'un utilisateur
- **URL** : `POST http://localhost:8080/roles/revoke`
- **Description** : Révocation d'un rôle d'un utilisateur spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Corps** : `{"user_id": 123, "role_id": 456}`
- **Réponse** : Confirmation de révocation
- **Authentification** : ✅ Token + Rôle admin requis

#### Récupération des rôles d'un utilisateur
- **URL** : `GET http://localhost:8080/roles/user/:user_id`
- **Description** : Récupération de tous les rôles attribués à un utilisateur
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Réponse** : Liste des rôles de l'utilisateur
- **Authentification** : ✅ Token + Rôle admin requis

---

## 🔗 Liaisons utilisateur-calendrier

### Routes admin (gestion des accès)

#### Récupération d'une liaison utilisateur-calendrier
- **URL** : `GET http://localhost:8080/user-calendar/:user_id/:calendar_id`
- **Description** : Récupération des détails d'accès d'un utilisateur à un calendrier
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur, `calendar_id` - ID du calendrier
- **Réponse** : Détails de la liaison avec permissions
- **Authentification** : ✅ Token + Rôle admin requis

#### Liste des calendriers d'un utilisateur
- **URL** : `GET http://localhost:8080/user-calendar/:user_id`
- **Description** : Récupération de tous les calendriers auxquels un utilisateur a accès
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur
- **Réponse** : Liste des calendriers avec permissions
- **Authentification** : ✅ Token + Rôle admin requis

#### Création d'une liaison utilisateur-calendrier
- **URL** : `POST http://localhost:8080/user-calendar/:user_id/:calendar_id`
- **Description** : Attribution d'un accès à un calendrier pour un utilisateur
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur, `calendar_id` - ID du calendrier
- **Corps** : `{"permission": "read_write"}`
- **Réponse** : Confirmation de création
- **Authentification** : ✅ Token + Rôle admin requis

#### Modification d'une liaison utilisateur-calendrier
- **URL** : `PUT http://localhost:8080/user-calendar/:user_id/:calendar_id`
- **Description** : Modification des permissions d'accès d'un utilisateur à un calendrier
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur, `calendar_id` - ID du calendrier
- **Corps** : `{"permission": "read_only"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token + Rôle admin requis

#### Suppression d'une liaison utilisateur-calendrier
- **URL** : `DELETE http://localhost:8080/user-calendar/:user_id/:calendar_id`
- **Description** : Suppression de l'accès d'un utilisateur à un calendrier
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `user_id` - ID de l'utilisateur, `calendar_id` - ID du calendrier
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token + Rôle admin requis

---

## 📅 Gestion des calendriers

### Routes protégées (gestion des calendriers)

#### Création d'un nouveau calendrier
- **URL** : `POST http://localhost:8080/calendar`
- **Description** : Création d'un nouveau calendrier par l'utilisateur connecté
- **Headers** : `Authorization: Bearer <token>`
- **Corps** : `{"name": "Mon Calendrier", "description": "Calendrier personnel"}`
- **Réponse** : Confirmation de création avec ID du calendrier
- **Authentification** : ✅ Token requis

#### Récupération d'un calendrier
- **URL** : `GET http://localhost:8080/calendar/:calendar_id`
- **Description** : Récupération des détails d'un calendrier (accès requis)
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier
- **Réponse** : Détails du calendrier
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Modification d'un calendrier
- **URL** : `PUT http://localhost:8080/calendar/:calendar_id`
- **Description** : Mise à jour des informations d'un calendrier (accès requis)
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier
- **Corps** : `{"name": "Nouveau Nom", "description": "Nouvelle description"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Suppression d'un calendrier
- **URL** : `DELETE http://localhost:8080/calendar/:calendar_id`
- **Description** : Suppression d'un calendrier (accès requis)
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token + Accès au calendrier requis

---

## 📝 Gestion des événements

### Routes protégées (gestion des événements)

#### Récupération d'un événement
- **URL** : `GET http://localhost:8080/calendar-event/:calendar_id/:event_id`
- **Description** : Récupération des détails d'un événement spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `event_id` - ID de l'événement
- **Réponse** : Détails de l'événement
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Liste des événements par mois
- **URL** : `GET http://localhost:8080/calendar-event/:calendar_id/month/:year/:month`
- **Description** : Récupération de tous les événements d'un mois spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `year` - Année, `month` - Mois (1-12)
- **Réponse** : Liste des événements du mois
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Liste des événements par semaine
- **URL** : `GET http://localhost:8080/calendar-event/:calendar_id/week/:year/:week`
- **Description** : Récupération de tous les événements d'une semaine spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `year` - Année, `week` - Numéro de semaine
- **Réponse** : Liste des événements de la semaine
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Liste des événements par jour
- **URL** : `GET http://localhost:8080/calendar-event/:calendar_id/day/:year/:month/:day`
- **Description** : Récupération de tous les événements d'un jour spécifique
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `year` - Année, `month` - Mois, `day` - Jour
- **Réponse** : Liste des événements du jour
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Création d'un nouvel événement
- **URL** : `POST http://localhost:8080/calendar-event/:calendar_id`
- **Description** : Création d'un nouvel événement dans un calendrier
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier
- **Corps** : `{"title": "Réunion", "description": "Réunion d'équipe", "start_time": "2025-01-15T10:00:00Z", "end_time": "2025-01-15T11:00:00Z"}`
- **Réponse** : Confirmation de création avec ID de l'événement
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Modification d'un événement
- **URL** : `PUT http://localhost:8080/calendar-event/:calendar_id/:event_id`
- **Description** : Mise à jour des informations d'un événement existant
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `event_id` - ID de l'événement
- **Corps** : `{"title": "Nouveau titre", "start_time": "2025-01-15T11:00:00Z"}`
- **Réponse** : Confirmation de mise à jour
- **Authentification** : ✅ Token + Accès au calendrier requis

#### Suppression d'un événement
- **URL** : `DELETE http://localhost:8080/calendar-event/:calendar_id/:event_id`
- **Description** : Suppression d'un événement d'un calendrier
- **Headers** : `Authorization: Bearer <token>`
- **Paramètres** : `calendar_id` - ID du calendrier, `event_id` - ID de l'événement
- **Réponse** : Confirmation de suppression
- **Authentification** : ✅ Token + Accès au calendrier requis

---

## 🔒 Niveaux d'autorisation

### 📊 Résumé des niveaux d'accès

| Niveau | Description | Routes concernées |
|--------|-------------|-------------------|
| **Public** | Aucune authentification requise | `/health`, `/auth/login`, `/auth/refresh`, `/user` (POST) |
| **Authentifié** | Token de session valide requis | `/auth/*` (protégées), `/user/me/*`, `/calendar/*`, `/calendar-event/*` |
| **Admin** | Token + rôle admin requis | `/user/:id/*`, `/roles/*`, `/user-calendar/*` |

### 🔐 Types de permissions

#### Permissions de calendrier
- **`read_only`** : Lecture seule des événements
- **`read_write`** : Lecture et modification des événements
- **`admin`** : Contrôle total (création, modification, suppression)

#### Rôles système
- **`user`** : Utilisateur standard
- **`admin`** : Administrateur avec accès complet
- **`moderator`** : Modérateur (si configuré)

---

## 📚 Exemples d'utilisation

### 🔐 Authentification
```bash
# Connexion
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Utilisation du token
curl -X GET http://localhost:8080/user/me \
  -H "Authorization: Bearer <token>"
```

### 📅 Création d'un événement
```bash
curl -X POST http://localhost:8080/calendar-event/123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Réunion d'équipe",
    "description": "Réunion hebdomadaire",
    "start_time": "2025-01-15T10:00:00Z",
    "end_time": "2025-01-15T11:00:00Z"
  }'
```

---

<div align="center">
  <p><em>Documentation API GoLendar v1.3.0</em></p>
  <p>📅 <strong>Dernière mise à jour :</strong> 30 juin 2025</p>
  <p>🏷️ <strong>Version :</strong> v1.3.0</p>
</div> 