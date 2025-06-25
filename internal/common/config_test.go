package common

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_VAR"
	os.Setenv(key, "valeur")
	defer os.Unsetenv(key)

	if v := getEnv(key, "defaut"); v != "valeur" {
		t.Errorf("getEnv(%q, %q) = %q, want %q", key, "defaut", v, "valeur")
	}

	os.Unsetenv(key)
	if v := getEnv(key, "defaut"); v != "defaut" {
		t.Errorf("getEnv(%q, %q) = %q, want %q", key, "defaut", v, "defaut")
	}
}
