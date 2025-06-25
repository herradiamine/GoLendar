package common

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleDBError_NoRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	res := HandleDBError(c, sql.ErrNoRows, http.StatusNotFound, "not found", "internal error")
	if !res {
		t.Error("HandleDBError devrait retourner true pour sql.ErrNoRows")
	}
}

func TestHandleDBError_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	err := errors.New("erreur générique")
	res := HandleDBError(c, err, http.StatusNotFound, "not found", "internal error")
	if !res {
		t.Error("HandleDBError devrait retourner true pour une erreur générique")
	}
}

func TestHandleDBError_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	res := HandleDBError(c, nil, http.StatusNotFound, "not found", "internal error")
	if res {
		t.Error("HandleDBError devrait retourner false si err est nil")
	}
}
