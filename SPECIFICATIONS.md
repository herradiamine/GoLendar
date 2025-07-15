# Spécifications des routes de l'API GoLendar

> Ce document décrit chaque route de l'API GoLendar, ses spécifications, un exemple de requête et de réponse, en français.

---

## 1. GET /health
**Description :** Vérifie l'état de santé de l'API.
**Spécifications :**
- Accès public.
- Retourne un statut simple.

**Exemple de réponse :**
```json
{
  "status": "healthy",
  "service": "GoLendar API"
}
```

---

## 2. POST /auth/login
**Description :** Authentifie un utilisateur et crée une session.
**Spécifications :**
- Accès public.
- Reçoit un email et un mot de passe.
- Retourne un token de session, un refresh token, l'utilisateur et ses rôles.

**Exemple de requête :**
```json
{
  "email": "jean.dupont@example.com",
  "password": "motdepasse123"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Connexion réussie",
  "data": {
    "user": {
      "user_id": 1,
      "lastname": "Dupont",
      "firstname": "Jean",
      "email": "jean.dupont@example.com",
      "created_at": "2024-06-01T12:00:00Z"
    },
    "session_token": "...",
    "refresh_token": "...",
    "expires_at": "2024-06-01T13:00:00Z",
    "roles": [
      { "role_id": 1, "name": "user", "description": "Utilisateur standard" }
    ]
  }
}
```

---

## 3. POST /auth/refresh
**Description :** Rafraîchit un token de session à partir d'un refresh token.
**Spécifications :**
- Accès public.
- Reçoit un refresh token.
- Retourne un nouveau session token et sa date d'expiration.

**Exemple de requête :**
```json
{
  "refresh_token": "..."
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Token rafraîchi avec succès",
  "data": {
    "session_token": "...",
    "expires_at": "2024-06-01T14:00:00Z"
  }
}
```

---

## 4. POST /auth/logout
**Description :** Déconnecte l'utilisateur en supprimant sa session.
**Spécifications :**
- Accès protégé (token requis dans le header Authorization).
- Désactive la session courante.

**Exemple de requête :**
Header :
```
Authorization: Bearer <session_token>
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Déconnexion réussie"
}
```

---

## 5. GET /auth/me
**Description :** Récupère les informations de l'utilisateur authentifié et ses rôles.
**Spécifications :**
- Accès protégé.
- Retourne l'utilisateur et ses rôles.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "lastname": "Dupont",
    "firstname": "Jean",
    "email": "jean.dupont@example.com",
    "created_at": "2024-06-01T12:00:00Z",
    "roles": [
      { "role_id": 1, "name": "user", "description": "Utilisateur standard" }
    ]
  }
}
```

---

## 6. GET /auth/sessions
**Description :** Récupère toutes les sessions actives de l'utilisateur.
**Spécifications :**
- Accès protégé.
- Retourne la liste des sessions (tokens masqués).

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "user_session_id": 1,
      "user_id": 1,
      "device_info": "Mozilla/5.0 ...",
      "ip_address": "192.168.1.1",
      "location": "Paris, France",
      "is_active": true,
      "created_at": "2024-06-01T12:00:00Z"
    }
  ]
}
```

---

## 7. DELETE /auth/sessions/:session_id
**Description :** Supprime une session spécifique de l'utilisateur.
**Spécifications :**
- Accès protégé.
- L'utilisateur ne peut supprimer que ses propres sessions.

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Session supprimée avec succès"
}
```

---

## 8. POST /user
**Description :** Crée un nouvel utilisateur (inscription).
**Spécifications :**
- Accès public.
- Reçoit nom, prénom, email, mot de passe.
- Retourne l'ID de l'utilisateur créé.

**Exemple de requête :**
```json
{
  "lastname": "Dupont",
  "firstname": "Jean",
  "email": "jean.dupont@example.com",
  "password": "motdepasse123"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Utilisateur créé avec succès",
  "data": { "user_id": 1 }
}
```

---

## 9. GET /user/me
**Description :** Récupère les informations de l'utilisateur connecté.
**Spécifications :**
- Accès protégé.
- Retourne les informations de l'utilisateur.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "user_id": 1,
    "lastname": "Dupont",
    "firstname": "Jean",
    "email": "jean.dupont@example.com",
    "created_at": "2024-06-01T12:00:00Z"
  }
}
```

---

## 10. PUT /user/me
**Description :** Met à jour les informations de l'utilisateur connecté.
**Spécifications :**
- Accès protégé.
- Peut modifier nom, prénom, email, mot de passe.

**Exemple de requête :**
```json
{
  "lastname": "Martin",
  "firstname": "Jean-Pierre",
  "email": "jean.martin@example.com",
  "password": "nouveaumotdepasse"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Utilisateur mis à jour avec succès"
}
```

---

## 11. DELETE /user/me
**Description :** Supprime (soft delete) l'utilisateur connecté.
**Spécifications :**
- Accès protégé.
- Supprime l'utilisateur et son mot de passe (soft delete).

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Utilisateur supprimé avec succès"
}
```

---

## 12. GET /user/:user_id
**Description :** Récupère un utilisateur par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Retourne les informations de l'utilisateur.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "user_id": 2,
    "lastname": "Martin",
    "firstname": "Marie",
    "email": "marie.martin@example.com",
    "created_at": "2024-06-01T12:00:00Z"
  }
}
```

---

## 13. PUT /user/:user_id
**Description :** Met à jour un utilisateur par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Peut modifier nom, prénom, email, mot de passe.

**Exemple de requête :**
```json
{
  "lastname": "Durand"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Utilisateur mis à jour avec succès"
}
```

---

## 14. DELETE /user/:user_id
**Description :** Supprime un utilisateur par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Supprime l'utilisateur et son mot de passe (soft delete).

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Utilisateur supprimé avec succès"
}
```

---

## 15. GET /user/:user_id/with-roles
**Description :** Récupère un utilisateur et ses rôles par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Retourne l'utilisateur et ses rôles.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "user_id": 2,
    "lastname": "Martin",
    "firstname": "Marie",
    "email": "marie.martin@example.com",
    "created_at": "2024-06-01T12:00:00Z",
    "roles": [
      { "role_id": 2, "name": "moderator", "description": "Modérateur" }
    ]
  }
}
```

---

## 16. GET /roles
**Description :** Récupère la liste de tous les rôles (admin).
**Spécifications :**
- Accès admin requis.
- Retourne la liste des rôles.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    { "role_id": 1, "name": "user", "description": "Utilisateur standard" },
    { "role_id": 2, "name": "admin", "description": "Administrateur" }
  ]
}
```

---

## 17. GET /roles/:id
**Description :** Récupère un rôle par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Retourne le rôle.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "role_id": 2,
    "name": "admin",
    "description": "Administrateur"
  }
}
```

---

## 18. POST /roles
**Description :** Crée un nouveau rôle (admin).
**Spécifications :**
- Accès admin requis.
- Reçoit nom et description.
- Retourne l'ID du rôle créé.

**Exemple de requête :**
```json
{
  "name": "moderator",
  "description": "Modérateur avec permissions limitées"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Rôle créé avec succès",
  "data": { "role_id": 3 }
}
```

---

## 19. PUT /roles/:id
**Description :** Met à jour un rôle par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Peut modifier nom et description.

**Exemple de requête :**
```json
{
  "name": "superadmin"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Rôle mis à jour avec succès"
}
```

---

## 20. DELETE /roles/:id
**Description :** Supprime un rôle par son ID (admin).
**Spécifications :**
- Accès admin requis.
- Ne peut pas supprimer le rôle admin.

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Rôle supprimé avec succès"
}
```

---

## 21. POST /roles/assign
**Description :** Attribue un rôle à un utilisateur (admin).
**Spécifications :**
- Accès admin requis.
- Reçoit user_id et role_id.

**Exemple de requête :**
```json
{
  "user_id": 2,
  "role_id": 3
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Rôle attribué avec succès"
}
```

---

## 22. POST /roles/revoke
**Description :** Révoque un rôle d'un utilisateur (admin).
**Spécifications :**
- Accès admin requis.
- Reçoit user_id et role_id.

**Exemple de requête :**
```json
{
  "user_id": 2,
  "role_id": 3
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Rôle révoqué avec succès"
}
```

---

## 23. GET /roles/user/:user_id
**Description :** Récupère les rôles d'un utilisateur (admin).
**Spécifications :**
- Accès admin requis.
- Retourne la liste des rôles de l'utilisateur.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    { "role_id": 2, "name": "moderator", "description": "Modérateur" }
  ]
}
```

---

## 24. GET /user-calendar/:user_id/:calendar_id
**Description :** Récupère une liaison user-calendar spécifique (admin).
**Spécifications :**
- Accès admin requis.
- Retourne la liaison user-calendar.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "user_calendar_id": 1,
    "user_id": 2,
    "calendar_id": 1,
    "created_at": "2024-06-01T12:00:00Z"
  }
}
```

---

## 25. GET /user-calendar/:user_id
**Description :** Récupère toutes les liaisons user-calendar d'un utilisateur (admin).
**Spécifications :**
- Accès admin requis.
- Retourne la liste des liaisons user-calendar.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "user_calendar_id": 1,
      "user_id": 2,
      "calendar_id": 1,
      "title": "Calendrier Pro",
      "description": "Calendrier professionnel"
    }
  ]
}
```

---

## 26. POST /user-calendar/:user_id/:calendar_id
**Description :** Crée une liaison user-calendar (admin).
**Spécifications :**
- Accès admin requis.
- Retourne l'ID de la liaison créée.

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Liaison user-calendar créée avec succès",
  "data": { "user_calendar_id": 1 }
}
```

---

## 27. PUT /user-calendar/:user_id/:calendar_id
**Description :** Met à jour une liaison user-calendar (admin).
**Spécifications :**
- Accès admin requis.
- Met à jour la date de modification.

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Liaison user-calendar mise à jour avec succès"
}
```

---

## 28. DELETE /user-calendar/:user_id/:calendar_id
**Description :** Supprime une liaison user-calendar (admin).
**Spécifications :**
- Accès admin requis.
- Supprime la liaison (soft delete).

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Liaison user-calendar supprimée avec succès"
}
```

---

## 29. GET /user-calendar/me
**Description :** Récupère toutes les liaisons user-calendar de l'utilisateur connecté.
**Spécifications :**
- Accès protégé.
- Retourne la liste des calendriers de l'utilisateur.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "user_calendar_id": 1,
      "user_id": 1,
      "calendar_id": 1,
      "title": "Calendrier Perso",
      "description": "Calendrier personnel"
    }
  ]
}
```

---

## 30. POST /calendar
**Description :** Crée un nouveau calendrier pour l'utilisateur connecté.
**Spécifications :**
- Accès protégé.
- Reçoit titre et description.
- Retourne l'ID du calendrier créé et l'ID utilisateur.

**Exemple de requête :**
```json
{
  "title": "Calendrier Pro",
  "description": "Calendrier professionnel"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Calendrier créé avec succès",
  "data": {
    "calendar_id": 1,
    "user_id": 1
  }
}
```

---

## 31. GET /calendar/:calendar_id
**Description :** Récupère un calendrier par son ID (accès si partagé ou propriétaire).
**Spécifications :**
- Accès protégé.
- Retourne le calendrier.

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Calendrier récupéré avec succès",
  "data": {
    "calendar_id": 1,
    "title": "Calendrier Pro",
    "description": "Calendrier professionnel"
  }
}
```

---

## 32. PUT /calendar/:calendar_id
**Description :** Met à jour un calendrier par son ID.
**Spécifications :**
- Accès protégé.
- Peut modifier titre et description.

**Exemple de requête :**
```json
{
  "title": "Nouveau titre",
  "description": "Nouvelle description"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Calendrier mis à jour avec succès"
}
```

---

## 33. DELETE /calendar/:calendar_id
**Description :** Supprime un calendrier par son ID (soft delete).
**Spécifications :**
- Accès protégé.
- Supprime le calendrier, ses liaisons user-calendar et ses événements (soft delete).

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Calendrier supprimé avec succès"
}
```

---

## 34. GET /calendar-event/:calendar_id/:event_id
**Description :** Récupère un événement par son ID dans un calendrier.
**Spécifications :**
- Accès protégé.
- Retourne l'événement.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "event_id": 1,
    "title": "Réunion équipe",
    "description": "Réunion hebdomadaire",
    "start": "2024-06-01T14:00:00Z",
    "duration": 60,
    "canceled": false
  }
}
```

---

## 35. GET /calendar-event/:calendar_id/month/:year/:month
**Description :** Récupère les événements d'un calendrier pour un mois donné.
**Spécifications :**
- Accès protégé.
- Retourne la liste des événements du mois.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "event_id": 1,
      "title": "Réunion équipe",
      "start": "2024-06-15T10:00:00Z",
      "duration": 60
    }
  ]
}
```

---

## 36. GET /calendar-event/:calendar_id/week/:year/:week
**Description :** Récupère les événements d'un calendrier pour une semaine donnée.
**Spécifications :**
- Accès protégé.
- Retourne la liste des événements de la semaine.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "event_id": 2,
      "title": "Sprint planning",
      "start": "2024-06-17T09:00:00Z",
      "duration": 90
    }
  ]
}
```

---

## 37. GET /calendar-event/:calendar_id/day/:year/:month/:day
**Description :** Récupère les événements d'un calendrier pour un jour donné.
**Spécifications :**
- Accès protégé.
- Retourne la liste des événements du jour.

**Exemple de réponse :**
```json
{
  "success": true,
  "data": [
    {
      "event_id": 3,
      "title": "Appel client",
      "start": "2024-06-18T11:00:00Z",
      "duration": 30
    }
  ]
}
```

---

## 38. POST /calendar-event/:calendar_id
**Description :** Crée un nouvel événement dans un calendrier.
**Spécifications :**
- Accès protégé.
- Reçoit titre, description, start, duration, canceled (optionnel).
- Retourne l'ID de l'événement créé et l'ID du calendrier.

**Exemple de requête :**
```json
{
  "title": "Réunion équipe",
  "description": "Réunion hebdomadaire",
  "start": "2024-06-20T10:00:00Z",
  "duration": 60
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Événement créé avec succès",
  "data": {
    "event_id": 4,
    "calendar_id": 1
  }
}
```

---

## 39. PUT /calendar-event/:calendar_id/:event_id
**Description :** Met à jour un événement dans un calendrier.
**Spécifications :**
- Accès protégé.
- Peut modifier titre, description, start, duration, canceled.

**Exemple de requête :**
```json
{
  "title": "Réunion modifiée"
}
```
**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Événement mis à jour avec succès"
}
```

---

## 40. DELETE /calendar-event/:calendar_id/:event_id
**Description :** Supprime un événement d'un calendrier (soft delete).
**Spécifications :**
- Accès protégé.
- Supprime l'événement (soft delete).

**Exemple de réponse :**
```json
{
  "success": true,
  "message": "Événement supprimé avec succès"
}
``` 