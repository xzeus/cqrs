package apiserver

import (
	"encoding/json"
	"fmt"
	"github.com/xzeus/cqrs/ioc"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

type Response interface {
	Flushed() bool
	ErrorUrl() string
	// Recorder returns the buffered response writer
	Recorder() *httptest.ResponseRecorder
	Header() http.Header
	// ResponseWriter  provides direct access to the underlying http.ResponseWriter
	ResponseWriter() http.ResponseWriter
	Empty(status int) error
	// Json writes the provided object as json to the buffer defaults status
	// to 200 if not provided, only reads first value if many are provided
	Json(data interface{}, status int) error
	View(viewmodel interface{}, err error)
	Text(content string, status int) error
	Binary(data []byte, mimeType string, status int) error
	// Error writes the provided message and data as a json struct to the buffer
	// and defaults status to 404 if not provided, only reads first value if many
	// are provided
	Error(message string, code int, status int, err error) error
	//
	Fail(err error) error
	Reset() error

	Flush() error
	Errorf(string, ...interface{})
}

type response struct {
	// True if the/(a) response has already been written
	flushed bool
	// Path pattern used to create error urls by code
	errorUrl string
	// Intermediary writer to capture buffered results
	recorder *httptest.ResponseRecorder
	// Writer that sends result directly to the client
	writer http.ResponseWriter
	// Provides a function to log an error message
	errorf func(string, ...interface{})
}

func NewResponse(w http.ResponseWriter) Response {
	return &response{
		flushed:  false,
		errorUrl: `https://api.vizidrix.com/v1/errors/%d`,
		recorder: httptest.NewRecorder(),
		writer:   w,
		errorf:   log.Printf,
	}
}

func (r *response) Flushed() bool {
	return r.flushed
}

func (r *response) ErrorUrl() string {
	return r.errorUrl
}

func (r *response) Recorder() *httptest.ResponseRecorder {
	return r.recorder
}

func (r *response) Header() http.Header {
	return r.Header()
}

func (r *response) ResponseWriter() http.ResponseWriter {
	return r.writer
}

func (r *response) Empty(status int) error {
	if r.flushed {
		return ErrResponseFlushed
	}
	r.recorder.WriteHeader(status)
	return nil
}

// Write encodes the provided object using the json serializer regardless of client request
func (r *response) Json(obj interface{}, status int) error {
	if r.flushed {
		return ErrResponseFlushed
	}
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	r.recorder.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.recorder.WriteHeader(status)
	if _, err := r.recorder.Write(data); err != nil {
		return err
	}
	return nil
}

func (r *response) View(viewmodel interface{}, err error) {
	switch err {
	case nil:
		r.Json(viewmodel, 200)
	case ioc.ErrNoSuchData:
		r.Error("no such data", 40000, 404, err)
	default:
		r.Error("invalid view response", 50000, 500, err)
	}
}

func (r *response) Text(content string, status int) error {
	if r.flushed {
		return ErrResponseFlushed
	}
	r.recorder.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.recorder.WriteHeader(status)
	io.WriteString(r.recorder, content)
	return nil
}

func (r *response) Binary(data []byte, mimeType string, status int) error {
	if r.flushed {
		return ErrResponseFlushed
	}
	r.recorder.Header().Set("Content-Type", mimeType)
	r.recorder.WriteHeader(status)
	r.recorder.Write(data)
	return nil
}

func (r *response) Error(message string, code int, status int, err error) error {
	if r.flushed {
		return ErrResponseFlushed
	}
	error_url := fmt.Sprintf(r.ErrorUrl(), code)
	return r.Json(struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Code    int    `json:"code"`
		Url     string `json:"url"`
	}{
		Status:  status,
		Message: message,
		Code:    code,
		Url:     error_url,
	}, status)
}

func (r *response) Fail(err error) error {
	r.Reset() // Clear any previous data
	return r.Error("Internal server error", 50000, 500, err)
}

func (r *response) Reset() error {
	if r.flushed {
		return ErrResponseFlushed
	}
	r.recorder = httptest.NewRecorder()
	return nil
}

func (r *response) Flush() error {
	if r.flushed {
		return ErrResponseFlushed
	}
	r.flushed = true
	for k, v := range r.recorder.Header() {
		r.writer.Header()[k] = v
	}
	// Set custom headers
	//r.writer.header().Set("blah", "blah")
	r.writer.WriteHeader(r.recorder.Code)
	r.writer.Write(r.recorder.Body.Bytes())
	r.recorder.Flush()
	return nil
}

func (r *response) Errorf(message string, values ...interface{}) {
	log.Printf("** apiresponse.Errorf:\n[\n%s\n%#v\n]\n", message, values)
	r.errorf(message, values...)
}
