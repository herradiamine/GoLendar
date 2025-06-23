package common

// Logs génériques
const (
	LogAppStart            = "[main][main]: Démarrage de l'application"
	LogDBConnectionSuccess = "[common][InitDB]: Connexion à la base de données réussie"
	LogDBConnectionError   = "[common][InitDB]: Erreur de connexion à la base de données"
	LogHTTPReceivedRequest = "[http][middleware]: Requête reçue"
	LogHTTPResponseSent    = "[http][middleware]: Réponse envoyée"
)

// Logs pour le package user
const (
	LogUserGet    = "[user][Get]: Récupération d'un utilisateur"
	LogUserAdd    = "[user][Add]: Création d'un utilisateur"
	LogUserUpdate = "[user][Update]: Mise à jour d'un utilisateur"
	LogUserDelete = "[user][Delete]: Suppression d'un utilisateur"
)

// Logs pour le package calendar
const (
	LogCalendarGet    = "[calendar][Get]: Récupération d'un calendrier"
	LogCalendarAdd    = "[calendar][Add]: Création d'un calendrier"
	LogCalendarUpdate = "[calendar][Update]: Mise à jour d'un calendrier"
	LogCalendarDelete = "[calendar][Delete]: Suppression d'un calendrier"
)

// Logs pour le package calendar_event
const (
	LogEventGet    = "[calendar_event][Get]: Récupération d'un événement"
	LogEventAdd    = "[calendar_event][Add]: Création d'un événement"
	LogEventUpdate = "[calendar_event][Update]: Mise à jour d'un événement"
	LogEventDelete = "[calendar_event][Delete]: Suppression d'un événement"
)

// Logs pour le package user_calendar
const (
	LogUserCalendarGet    = "[user_calendar][Get]: Récupération d'une liaison utilisateur-calendrier"
	LogUserCalendarAdd    = "[user_calendar][Add]: Création d'une liaison utilisateur-calendrier"
	LogUserCalendarUpdate = "[user_calendar][Update]: Mise à jour d'une liaison utilisateur-calendrier"
	LogUserCalendarDelete = "[user_calendar][Delete]: Suppression d'une liaison utilisateur-calendrier"
)
