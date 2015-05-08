package apiserver

import (
	j "github.com/vizidrix/jose"
	"log"
	"strings"
)

// Provides a lightweight middleware which configures CORS
func CORS(h ApiFunc) ApiFunc {
	return func(req Request, resp Response) {
		r := req.Request()
		if r.Method != OPTIONS {
			defer h(req, resp)
		}
		ref := r.Referer()
		if ref == "" {
			if r.TLS == nil {
				ref = "http://" + r.Host
			} else {
				ref = "https://" + r.Host
			}
		}
		ref = strings.TrimRight(ref, "/")

		resp.Recorder().Header().Set(CORS_AccessControlAllowOrigin, ref) // Omit trailing slash
		resp.Recorder().Header().Set(CORS_AccessControlAllowMethods, CORS_DefaultMethods)
		resp.Recorder().Header().Set(CORS_AccessControlAllowHeaders, CORS_DefaultHeaders)
	}
}

const authorization = "Authorization"

func LoadTokens() MiddlewareFunc {
	return func(h ApiFunc) ApiFunc {
		return func(req Request, resp Response) {
			defer h(req, resp)
			t := make(map[string]*j.TokenDef)
			for k, v := range req.Request().Header {
				log.Printf("K: [ %s ] V: [ %s ]", k, v)
				if k == authorization && len(v) == 1 {
					if v[0][:7] == "Bearer" {
						log.Printf("Bearer token")
					}
					continue
				}
				if strings.Contains(k, "token") {
					log.Printf("Header token")
					continue
				}
			}
			log.Printf("Found token(s): [\n%s\n]\n\n", t)
		}
	}
}

/*
func ValidateToken(token_key string, sources ...TokenSource) MiddlewareFunc {
	return func(h ApiFunc) ApiFunc {
		return func(req Request, resp Response) {
			if req.Request().Method != OPTIONS {
				req.t
				if _, err := req.GetToken(token_key, sources...); err != nil {
					resp.Error("invalid session", 40060, 500, ErrInvalidToken)
					return
				}
			}
			h(req, resp)
		}
	}
}




type TokenSource func(Request, string) string

func FromHeader() TokenSource {
	return func(r Request, key string) string {
		return r.Request().Header.Get(fmt.Sprintf("x-%s-token", key))
	}
}

func FromQueryString(param string) TokenSource {
	return func(r Request, key string) string {
		return r.Request().URL.Query().Get(param)
	}
}

// GetToken returns a validated token from the request with key pattern x-{key}-token i.e. x-session-token
func (r *request) GetToken(key string, sources ...TokenSource) (t *j.TokenDef, err error) {
	key = strings.ToLower(key)
	if t, ok := r.tokens[key]; ok {
		//log.Printf("Cached token [ %s ]", key)
		return t, nil
	}
	if len(sources) == 0 { // Default to header
		sources = append(sources, FromHeader())
	}
	err = ErrInvalidToken
	for _, s := range sources {
		if t_str := s(r, key); t_str == "" {
			continue
		} else { // Got some data from the source
			//log.Printf("Got some token data for [ %s ] from source [ %s ]", key, t_str)
			// TODO: Extract configuration of token read somewhere (deps?)
			c := r.deps.Crypto()
			if t, err = c.DecodeToken([]byte(t_str), j.RemoveConstraints(j.CONST_None_Algo)); err != nil {
				//if t, err = j.Decode([]byte(t_str), j.RemoveConstraints(j.None_Algo)); err != nil {
				continue
			}
			t.
			if errs := t.Validate(); errs != nil && len(errs) > 0 {
				err = ErrInvalidToken
				continue
			}
			err = nil // Found a good token, ignore previous errors
			r.tokens[key] = t
		}
	}
	if err != nil {
		return nil, err
	}
	return
}
*/
