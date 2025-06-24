#!/bin/bash

# Script de build et déploiement Docker pour GoLendar

set -e

echo "🚀 Démarrage du build et déploiement Docker pour GoLendar..."

# Variables
IMAGE_NAME="golendar"
TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages colorés
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Vérifier que Docker est installé
if ! command -v docker &> /dev/null; then
    print_error "Docker n'est pas installé ou n'est pas dans le PATH"
    exit 1
fi

# Vérifier que Docker Compose est installé (V1 ou V2)
if command -v docker compose &> /dev/null; then
    DC="docker compose"
elif command -v docker-compose &> /dev/null; then
    DC="docker-compose"
else
    print_error "Docker Compose n'est pas installé ou n'est pas dans le PATH"
    exit 1
fi

# Nettoyer les anciens conteneurs et images (optionnel)
if [ "$1" = "--clean" ]; then
    print_warning "Nettoyage des anciens conteneurs et images..."
    $DC down --remove-orphans
    docker system prune -f
fi

# Build de l'image
print_status "Build de l'image Docker..."
docker build -t ${FULL_IMAGE_NAME} .

if [ $? -eq 0 ]; then
    print_status "✅ Build réussi !"
else
    print_error "❌ Échec du build"
    exit 1
fi

# Afficher les informations sur l'image
print_status "Informations sur l'image créée :"
docker images ${FULL_IMAGE_NAME}

# Arrêter les conteneurs existants
print_status "Arrêt des conteneurs existants..."
$DC down --remove-orphans

# Démarrer l'application
print_status "Démarrage de l'application avec Docker Compose..."
$DC up -d

# Attendre que les services soient prêts
print_status "Attente du démarrage des services..."
sleep 10

# Attendre que MySQL soit complètement prêt
print_status "Attente que MySQL soit prêt..."
until $DC exec -T golendar_db mysqladmin ping -h localhost --silent; do
    print_status "MySQL démarre encore..."
    sleep 5
done
print_status "✅ MySQL est prêt !"

# Importer le schéma SQL
print_status "Import du schéma SQL..."
$DC exec -T golendar_db mysql -u root -ppassword calendar < resources/schema.sql
if [ $? -eq 0 ]; then
    print_status "✅ Schéma SQL importé avec succès !"
else
    print_warning "⚠️  Erreur lors de l'import du schéma SQL (peut-être déjà présent)"
fi

# Vérifier le statut des conteneurs
print_status "Statut des conteneurs :"
$DC ps

# Tester l'endpoint de santé
print_status "Test de l'endpoint de santé..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    print_status "✅ Application démarrée avec succès !"
    print_status "🌐 API accessible sur : http://localhost:8080"
    print_status "📊 Health check : http://localhost:8080/health"
else
    print_warning "⚠️  L'application démarre encore..."
    print_status "Vérifiez les logs avec : $DC logs -f golendar_app"
fi

print_status "🎉 Build et déploiement terminés !"
print_status "Commandes utiles :"
print_status "  - Voir les logs : $DC logs -f golendar_app"
print_status "  - Arrêter l'app : $DC down"
print_status "  - Redémarrer : $DC restart" 