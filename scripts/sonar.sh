#!/bin/bash

# Script d'analyse SonarCloud pour GoLendar
set -e

echo "🚀 Démarrage de l'analyse SonarCloud pour GoLendar..."

# Configuration Git pour éviter le shallow clone
echo "🔧 Configuration Git pour l'analyse SonarCloud..."
git config --global fetch.unshallow true
git fetch --unshallow || echo "⚠️  Le dépôt n'est pas un shallow clone ou l'historique complet est déjà disponible"

# Vérifier si le token SonarCloud est défini
if [ -z "$SONAR_TOKEN" ]; then
    echo "❌ Variable d'environnement SONAR_TOKEN non définie"
    echo "💡 Exportez votre token SonarCloud avec: export SONAR_TOKEN=votre_token"
    echo "🔗 Obtenez votre token sur: https://sonarcloud.io/account/security/"
    exit 1
fi

echo "✅ Token SonarCloud détecté"

# Générer la couverture de code
echo "🧪 Génération de la couverture de code..."
go test -coverprofile=coverage.out -covermode=atomic ./...

# Générer le rapport de tests
echo "📊 Génération du rapport de tests..."
go test -json ./... > test-report.json

# Installer sonar-scanner si nécessaire
if ! command -v sonar-scanner &> /dev/null; then
    echo "📦 Installation de sonar-scanner..."
    if command -v docker &> /dev/null; then
        echo "🐳 Utilisation de sonar-scanner via Docker..."
        SONAR_SCANNER="docker run --rm -v $(pwd):/usr/src -e SONAR_TOKEN=$SONAR_TOKEN sonarqube:latest sonar-scanner"
    else
        echo "❌ Sonar-scanner n'est pas installé et Docker n'est pas disponible"
        echo "💡 Installez sonar-scanner ou Docker pour continuer"
        exit 1
    fi
else
    SONAR_SCANNER="sonar-scanner"
fi

# Lancer l'analyse SonarCloud
echo "🔍 Lancement de l'analyse SonarCloud..."
$SONAR_SCANNER

echo "✅ Analyse SonarCloud terminée!"
echo "🌐 Consultez les résultats sur: https://sonarcloud.io/project/overview?id=herradiamine_GoLendar" 