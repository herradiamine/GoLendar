package common

import "testing"

func TestInitDB_InvalidConfig(t *testing.T) {
	cfg := DBConfig{
		User:     "invalid",
		Password: "invalid",
		Host:     "invalid",
		Port:     3306,
		Name:     "invalid",
	}
	err := InitDB(cfg)
	if err == nil {
		t.Error("InitDB avec une config invalide devrait retourner une erreur")
	}
}
