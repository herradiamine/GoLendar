#!/bin/bash

# Script de build et déploiement Docker pour GoLendar v1.2.0
# Inclut l'authentification, la gestion des rôles et l'intégration SonarCloud

set -e

echo "🚀 Démarrage du build et déploiement Docker pour GoLendar v1.2.0..."

# Variables
IMAGE_NAME="golendar"
TAG="v1.2.0"
FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"
VERSION="1.2.0"

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

print_success() {
    echo -e "${BLUE}[SUCCESS]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}[HEADER]${NC} $1"
}

# Fonction d'aide
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --clean     Nettoyer complètement (conteneurs, volumes, images)"
    echo "  --force     Forcer la reconstruction complète (no-cache)"
    echo "  --rebuild   Nettoyer + Forcer la reconstruction"
    echo "  --help      Afficher cette aide"
    echo ""
    echo "Exemples:"
    echo "  $0              # Build normal"
    echo "  $0 --clean      # Nettoyer avant build"
    echo "  $0 --force      # Build sans cache"
    echo "  $0 --rebuild    # Nettoyer + Build sans cache"
}

# Traitement des arguments
FORCE_REBUILD=false
CLEAN_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --force)
            FORCE_REBUILD=true
            shift
            ;;
        --rebuild)
            CLEAN_BUILD=true
            FORCE_REBUILD=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            print_error "Option inconnue: $1"
            show_help
            exit 1
            ;;
    esac
done

# Afficher l'en-tête
print_header "=========================================="
print_header "    GoLendar v1.2.0 - Déploiement"
print_header "=========================================="
print_header "Nouvelles fonctionnalités :"
print_header "  🔐 Authentification par sessions"
print_header "  🔑 Gestion des rôles et permissions"
print_header "  📊 Intégration SonarCloud"
print_header "  📚 Documentation enrichie"
print_header "=========================================="

if [ "$FORCE_REBUILD" = true ]; then
    print_warning "🔄 Mode reconstruction forcée activé (no-cache)"
fi

if [ "$CLEAN_BUILD" = true ]; then
    print_warning "🧹 Mode nettoyage complet activé"
fi

# Vérifier que Docker est installé
if ! command -v docker &> /dev/null; then
    print_error "Docker n'est pas installé ou n'est pas dans le PATH"
    exit 1
fi

# Vérifier que Docker Compose est installé (V1 ou V2)
if command -v docker compose &> /dev/null; then
    DC="docker compose"
    print_status "Utilisation de Docker Compose V2"
elif command -v docker-compose &> /dev/null; then
    DC="docker-compose"
    print_status "Utilisation de Docker Compose V1"
else
    print_error "Docker Compose n'est pas installé ou n'est pas dans le PATH"
    exit 1
fi

# Vérifier la version de Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Version Go détectée : $GO_VERSION"
    if [[ "$GO_VERSION" < "1.24" ]]; then
        print_warning "Version Go recommandée : 1.24+ (actuelle : $GO_VERSION)"
    fi
else
    print_warning "Go n'est pas installé localement (utilisera l'image Docker)"
fi

# Nettoyer les anciens conteneurs et images (si demandé)
if [ "$CLEAN_BUILD" = true ]; then
    print_warning "Nettoyage complet des anciens conteneurs, volumes et réseaux..."
    $DC down --volumes --remove-orphans 2>/dev/null || true
    
    print_status "Suppression des images Docker du projet..."
    docker rmi ${IMAGE_NAME}:${TAG} 2>/dev/null || true
    docker rmi ${IMAGE_NAME}:latest 2>/dev/null || true
    
    print_status "Nettoyage du cache Docker..."
    docker system prune -f
    
    print_success "Nettoyage terminé"
fi

# Vérifier les fichiers de configuration
print_status "Vérification des fichiers de configuration..."
if [ ! -f "docker-compose.yml" ]; then
    print_error "docker-compose.yml manquant"
    exit 1
fi

if [ ! -f "Dockerfile" ]; then
    print_error "Dockerfile manquant"
    exit 1
fi

if [ ! -f "resources/schema.sql" ]; then
    print_error "resources/schema.sql manquant"
    exit 1
fi

if [ ! -f "go.mod" ]; then
    print_error "go.mod manquant"
    exit 1
fi

print_success "Configuration validée"

# Nettoyer le cache Go local (si Go est installé)
if command -v go &> /dev/null; then
    print_status "Nettoyage du cache Go local..."
    go clean -cache -modcache -testcache 2>/dev/null || true
    print_success "Cache Go nettoyé"
fi

# Compiler le projet Go localement
print_status "Compilation du projet Go..."
if command -v go &> /dev/null; then
    # Compiler le binaire
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o golendar-app ./cmd/app
    if [ $? -eq 0 ]; then
        print_success "✅ Compilation Go réussie !"
    else
        print_error "❌ Échec de la compilation Go"
        exit 1
    fi
else
    print_error "Go n'est pas installé localement. Impossible de compiler."
    exit 1
fi

# Build de l'image avec tag de version
print_status "Build de l'image Docker (version $VERSION)..."
BUILD_ARGS=""
if [ "$FORCE_REBUILD" = true ]; then
    BUILD_ARGS="--no-cache --pull"
    print_status "Build sans cache activé"
fi

# Build avec le binaire pré-compilé
docker build ${BUILD_ARGS} \
    --build-arg VERSION=${VERSION} \
    -t ${FULL_IMAGE_NAME} -t ${IMAGE_NAME}:latest .

if [ $? -eq 0 ]; then
    print_success "✅ Build Docker réussi !"
    
    # Nettoyer le binaire temporaire
    rm -f golendar-app
    print_status "Binaire temporaire nettoyé"
else
    print_error "❌ Échec du build Docker"
    rm -f golendar-app
    exit 1
fi

# Afficher les informations sur l'image
print_status "Informations sur l'image créée :"
docker images ${IMAGE_NAME}

# Arrêter les conteneurs existants
print_status "Arrêt des conteneurs existants..."
$DC down --remove-orphans

# Démarrer l'application
print_status "Démarrage de l'application avec Docker Compose..."
$DC up -d

# Attendre que les services soient prêts
print_status "Attente du démarrage des services..."
sleep 15

# Attendre que MySQL soit complètement prêt
print_status "Attente que MySQL soit prêt..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if $DC exec -T golendar_db mysqladmin ping -h localhost --silent 2>/dev/null; then
        print_success "✅ MySQL est prêt !"
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        print_status "MySQL démarre encore... (tentative $RETRY_COUNT/$MAX_RETRIES)"
        sleep 5
    fi
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    print_error "❌ MySQL n'a pas démarré dans le délai imparti"
    print_status "Vérifiez les logs MySQL : $DC logs golendar_db"
    exit 1
fi

# Importer le schéma SQL
print_status "Import du schéma SQL..."
if $DC exec -T golendar_db mysql -u root -ppassword calendar < resources/schema.sql 2>/dev/null; then
    print_success "✅ Schéma SQL importé avec succès !"
else
    print_warning "⚠️  Erreur lors de l'import du schéma SQL (peut-être déjà présent)"
fi

# Vérifier le statut des conteneurs
print_status "Statut des conteneurs :"
$DC ps

# Tester l'endpoint de santé avec retry
print_status "Test de l'endpoint de santé..."
HEALTH_RETRIES=10
HEALTH_COUNT=0

while [ $HEALTH_COUNT -lt $HEALTH_RETRIES ]; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "✅ Application démarrée avec succès !"
        break
    else
        HEALTH_COUNT=$((HEALTH_COUNT + 1))
        print_status "Application démarre encore... (tentative $HEALTH_COUNT/$HEALTH_RETRIES)"
        sleep 3
    fi
done

if [ $HEALTH_COUNT -eq $HEALTH_RETRIES ]; then
    print_warning "⚠️  L'application prend du temps à démarrer"
    print_status "Vérifiez les logs : $DC logs -f golendar_back"
else
    # Afficher les informations de l'API
    print_success "🎉 GoLendar v1.2.0 déployé avec succès !"
    echo ""
    print_header "📋 Informations de l'API :"
    print_status "  🌐 URL principale : http://localhost:8080"
    print_status "  🏥 Health check : http://localhost:8080/health"
    print_status "  📚 Documentation : Voir README.md"
    print_status "  📊 SonarCloud : https://sonarcloud.io/project/overview?id=herradiamine_GoLendar"
    echo ""
    print_header "🔐 Nouveaux endpoints d'authentification :"
    print_status "  POST /auth/login - Connexion"
    print_status "  POST /auth/refresh - Renouvellement de token"
    print_status "  POST /auth/logout - Déconnexion"
    print_status "  GET  /auth/me - Profil utilisateur"
    echo ""
    print_header "🔑 Nouveaux endpoints de gestion des rôles :"
    print_status "  GET    /roles - Liste des rôles"
    print_status "  POST   /roles - Créer un rôle"
    print_status "  POST   /roles/assign - Attribuer un rôle"
    print_status "  POST   /roles/revoke - Révoquer un rôle"
    echo ""
    print_header "📅 Endpoints mis à jour :"
    print_status "  GET    /user/me - Mon profil"
    print_status "  GET    /calendar/:id - Calendrier (authentification requise)"
    print_status "  GET    /calendar-event/:calendar_id/:event_id - Événement"
    echo ""
    print_header "🧪 Test rapide :"
    print_status "  curl http://localhost:8080/health"
    print_status "  curl -X POST http://localhost:8080/user -d '{\"firstname\":\"Test\",\"lastname\":\"User\",\"email\":\"test@example.com\",\"password\":\"password123\"}' -H 'Content-Type: application/json'"
fi

echo ""
print_header "🛠️  Commandes utiles :"
print_status "  📋 Voir les logs : $DC logs -f golendar_back"
print_status "  🗄️  Logs base de données : $DC logs -f golendar_db"
print_status "  ⏹️  Arrêter l'app : $DC down"
print_status "  🔄 Redémarrer : $DC restart"
print_status "  🧹 Nettoyer : $DC down --volumes"
print_status "  📊 SonarCloud : ./scripts/sonar.sh"
echo ""
print_header "📚 Documentation :"
print_status "  📖 README complet : README.md"
print_status "  📋 Collection Postman : resources/postman_collection.json"
print_status "  🔧 Configuration : sonar-project.properties"
echo ""
print_success "🎉 Déploiement GoLendar v1.2.0 terminé !" 