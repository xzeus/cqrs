package ioc

import (
	"net/http"
)

type HttpClient interface {
	Ptr() *http.Client
	BaseUri() string
	MakeUri(root string, data map[string]string) string
	Exec(action string, data map[string]string) ([]byte, error)
	Json(action string, data map[string]string, result interface{}) ([]byte, error)
}
