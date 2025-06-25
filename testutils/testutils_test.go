package testutils

import (
	"testing"
)

func TestSetupTestRouter(t *testing.T) {
	r := SetupTestRouter()
	if r == nil {
		t.Error("SetupTestRouter doit retourner un *gin.Engine non nil")
	}
}

// ATTENTION : ce test suppose que la base de test est accessible en local.
// Il vérifie simplement que la fonction ne panique pas.
func TestSetupTestDB(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetupTestDB a paniqué: %v", r)
		}
	}()
	SetupTestDB()
}
