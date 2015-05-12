package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	j "github.com/vizidrix/jose"
	"github.com/xzeus/cqrs/ioc"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type Int32Extractor func() (int32, bool)
type Int64Extractor func() (int64, bool)
type StringExtractor func() (string, bool)
type BinaryExtractor func() ([]byte, bool)

type Request interface {
	Deps() ioc.Dependencies
	Request() *http.Request
	GetToken(string) (*j.TokenDef, error)
	BaseUri() string
	Segment() (string, error)
	ExtractInt32ElementId() (int32, bool)
	ExtractInt64ElementId() (int64, bool)
	ExtractStringElementId() (string, bool)
	Int32ExtractorByQuery(string) Int32Extractor
	Int64ExtractorByQuery(string) Int64Extractor
	StringExtractorByQuery(string) StringExtractor
	Json(viewmodel interface{}) error
	QueryParams() map[string][]string
}

type request struct {
	deps    ioc.Dependencies
	request *http.Request
	tokens  map[string]*j.TokenDef // Only populated with verified tokens
}

func NewRequest(deps ioc.Dependencies, r *http.Request) Request {
	return &request{
		deps:    deps,
		request: r,
		tokens:  make(map[string]*j.TokenDef),
	}
}

func (r *request) Deps() ioc.Dependencies {
	return r.deps
}

func (r *request) Request() *http.Request {
	return r.request
}

func (r *request) PutToken(key string, t *j.TokenDef) {
	r.tokens[key] = t
}

func (r *request) GetToken(key string) (*j.TokenDef, error) {
	if t, ok := r.tokens[key]; ok {
		return t, nil
	}
	return nil, ErrInvalidToken
}

func (r *request) BaseUri() (result string) {
	scheme := "http"
	if r.Request().TLS != nil {
		scheme = "https"
	}
	b := r.deps.HttpClient().BaseUri()
	result = fmt.Sprintf("%s://%s/", scheme, b)
	return
}

func (r *request) Segment() (string, error) {
	return path.Base(r.Request().URL.Path), nil
}

func (r *request) ExtractInt32ElementId() (int32, bool) {
	var str, _ = r.Segment()
	id, err := Int32(str)
	return int32(id), err == nil
}

func (r *request) ExtractInt64ElementId() (int64, bool) {
	var str, _ = r.Segment()
	id, err := Int64(str)
	return id, err == nil
}

func (r *request) ExtractStringElementId() (string, bool) {
	var id, err = r.Segment()
	return id, id == "" || err != nil
}

func (r *request) Int32ExtractorByQuery(key string) Int32Extractor {
	return func() (int32, bool) {
		var params, ok = r.QueryParams()[key]
		if !ok || len(params) == 0 {
			return 0, false
		}
		query_id, err := Int32(params[0])
		return int32(query_id), err == nil
	}
}

func (r *request) Int64ExtractorByQuery(key string) Int64Extractor {
	return func() (int64, bool) {
		var params, ok = r.QueryParams()[key]
		if !ok || len(params) == 0 {
			return 0, false
		}
		query_id, err := Int64(params[0])
		return query_id, err == nil
	}
}

func (r *request) StringExtractorByQuery(key string) StringExtractor {
	return func() (string, bool) {
		var params, ok = r.QueryParams()[key]
		if !ok || len(params) == 0 {
			return "", false
		}
		return params[0], true
	}
}

func (r *request) Json(viewmodel interface{}) (err error) {
	var data []byte
	if data, err = ioutil.ReadAll(r.Request().Body); err != nil {
		return errors.New(fmt.Sprintf("%s [ %s ]", err, data))
	}
	if err = json.Unmarshal(data, viewmodel); err != nil {
		return errors.New(fmt.Sprintf("%s [ %s ]", err, data))
	}
	return
}

func (r *request) QueryParams() map[string][]string {
	var result, _ = url.ParseQuery(r.Request().URL.RawQuery)
	return result
}
