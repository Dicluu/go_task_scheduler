package testhlpr

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func SendRequest(handler http.HandlerFunc, t *testing.T, url, method, input string, status int) string {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(input)))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, status)

	return rr.Body.String()
}
