package httpcache

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

type DummyCache struct {
	cache map[string][]byte
}

func (d *DummyCache) ShouldCached(res *http.Response) bool {
	return true
}

func (d *DummyCache) Get(req *http.Request) ([]byte, error) {
	if val, ok := d.cache[req.URL.Path]; ok {
		return val, nil
	}
	return nil, errors.New("cache miss")
}

func (d *DummyCache) Set(req *http.Request, res []byte) error {
	d.cache[req.URL.Path] = res
	return nil
}

func NewDummyCache() Cache {
	return &DummyCache{cache: make(map[string][]byte)}
}

func TestRouter(t *testing.T) {
	r := chi.NewRouter()
	c := NewDummyCache()
	r.Use(NewMiddleware(c))
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "hoge")
		w.Write([]byte("welcome"))
	})
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	req, _ := http.NewRequest("GET", testServer.URL+"/hello", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := httputil.DumpResponse(resp, true)
	fmt.Println("test1", string(result))

	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	result2, _ := httputil.DumpResponse(resp2, true)
	fmt.Println("test2", string(result2))
}
