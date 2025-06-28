package user_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-averroes/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var testGinRouter *gin.Engine // Routeur de test global
var testServer *httptest.Server
var testClient *http.Client

// TestMain configure l'environnement de test global
func TestMain(m *testing.M) {
	if err := testutils.SetupTestEnvironment(); err != nil {
		panic("Impossible d'initialiser l'environnement de test: " + err.Error())
	}
	testGinRouter = testutils.CreateTestRouter()
	testServer = httptest.NewServer(testGinRouter)
	testClient = testServer.Client()
	code := m.Run()
	if err := testutils.TeardownTestEnvironment(); err != nil {
		panic("Impossible de nettoyer l'environnement de test: " + err.Error())
	}
	testServer.Close()
	os.Exit(code)
}

// TestRouteExample teste la route POST/PUT d'exemple avec plusieurs cas
func TestRoutePostPutExample(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		RequestData      func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Case name",
			CaseUrl:  "Url",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Success message",
			ExpectedError:    "Error message",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			// On traite les cas de test un par un.
			require.Equal(t, testCase.CaseUrl, "Url")
			require.Equal(t, testCase.ExpectedHttpCode, http.StatusOK)
			require.Equal(t, testCase.ExpectedMessage, "Success message")
			require.Equal(t, testCase.ExpectedError, "Error message")
			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestRouteExample teste la route GET/DELETE d'exemple avec plusieurs cas
func TestRouteGetDeleteExample(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Case name",
			CaseUrl:  "Url",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Success message",
			ExpectedError:    "Error message",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			// On traite les cas de test un par un.
			require.Equal(t, testCase.CaseUrl, "Url")
			require.Equal(t, testCase.ExpectedHttpCode, http.StatusOK)
			require.Equal(t, testCase.ExpectedMessage, "Success message")
			require.Equal(t, testCase.ExpectedError, "Error message")
			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
