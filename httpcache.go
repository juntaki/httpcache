package httpcache

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
)

type CachedResponseRecorder struct {
	*httptest.ResponseRecorder
	Body *bytes.Buffer
}

func NewCachedResponseRecorder() *CachedResponseRecorder {
	return &CachedResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		Body:             new(bytes.Buffer),
	}
}

func (t *CachedResponseRecorder) Write(p []byte) (int, error) {
	return io.MultiWriter(t.ResponseRecorder, t.Body).Write(p)
}

type Cache interface {
	ShouldCached(res *http.Response) bool
	Get(req *http.Request) ([]byte, error)
	Set(req *http.Request, res []byte) error
}

func NewMiddleware(cache Cache) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val, err := cache.Get(r); err == nil {
				res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(val)), r)

				for k, v := range res.Header {
					w.Header().Set(k, strings.Join(v, ","))
				}
				w.WriteHeader(res.StatusCode)
				_, err = io.Copy(w, res.Body)
				if err != nil {
					return
				}
				return
			}

			rec := NewCachedResponseRecorder()
			next.ServeHTTP(rec, r)
			rec.Flush()
			res := rec.Result()

			for k, v := range res.Header {
				w.Header().Set(k, strings.Join(v, ","))
			}
			w.WriteHeader(res.StatusCode)
			_, err := io.Copy(w, rec.Body)
			if err != nil {
				return
			}

			if cache.ShouldCached(res) {
				val, err := httputil.DumpResponse(res, true)
				if err != nil {
					return
				}
				cache.Set(r, val)
			}
		})
	}
}
