#!/bin/bash

# Script d'analyse SonarCloud pour GoLendar
set -e

echo "🚀 Démarrage de l'analyse SonarCloud pour GoLendar..."

# Configuration Git pour éviter le shallow clone
echo "🔧 Configuration Git pour l'analyse SonarCloud..."

# Vérifier si c'est un shallow clone et le corriger si nécessaire
if [ -f ".git/shallow" ]; then
    echo "📋 Détection d'un shallow clone, conversion en clone complet..."
    git fetch --unshallow
    git fetch --all
    git fetch --tags
    echo "✅ Conversion en clone complet terminée"
else
    echo "ℹ️  Le dépôt n'est pas un shallow clone"
fi

# Configuration Git pour SonarQube
git config --global fetch.unshallow true

# S'assurer que l'historique complet est disponible
echo "📥 Récupération de l'historique complet..."
git fetch --all --tags --unshallow || echo "⚠️  L'historique complet est déjà disponible"

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
go test -p=1 -coverprofile=coverage.out -covermode=atomic ./...

# Générer le rapport de tests
echo "📊 Génération du rapport de tests..."
go test -p=1 -json ./... > test-report.json

# Créer le dossier reports s'il n'existe pas
mkdir -p reports

# Déplacer les fichiers de rapport
mv coverage.out reports/
mv test-report.json reports/

echo "📁 Rapports générés dans le dossier reports/"

# Lancer l'analyse SonarCloud
echo "🔍 Lancement de l'analyse SonarCloud..."
sonar-scanner \
    -Dsonar.projectKey=herradiamine_GoLendar \
    -Dsonar.organization=herradiamine \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.login=$SONAR_TOKEN \
    -Dsonar.sources=cmd,internal \
    -Dsonar.tests=cmd,internal \
    -Dsonar.go.coverage.reportPaths=reports/coverage.out \
    -Dsonar.go.tests.reportPaths=reports/test-report.json \
    -Dsonar.exclusions=**/*_test.go,**/vendor/**,**/node_modules/**,**/*.pb.go,**/mocks/**,**/testutils/** \
    -Dsonar.test.inclusions=**/*_test.go \
    -Dsonar.qualitygate.wait=true \
    -Dsonar.scm.disabled=false \
    -Dsonar.scm.provider=git

echo "✅ Analyse SonarCloud terminée avec succès !"
echo "🔗 Consultez les résultats sur: https://sonarcloud.io/dashboard?id=herradiamine_GoLendar" 