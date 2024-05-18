package replay

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function to create an HTTP request with headers
func createRequest(url, method string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req
}

// Helper function to create an HTTP response
func createResponse(statusCode int, body string, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode:    statusCode,
		Header:        http.Header{},
		Body:          io.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
	}
	for k, v := range headers {
		resp.Header.Add(k, v)
	}
	return resp
}

// Test initialization with default and custom configurations
func TestNewCache(t *testing.T) {
	cache := NewCache()
	assert.Equal(t, DefaultMaxSize, cache.maxSize)
	assert.Equal(t, DefaultEvictionPolicy, cache.evictionPolicy)
	assert.Equal(t, DefaultTTL, cache.ttl)
	assert.Equal(t, []string{DefaultFilter}, cache.cacheFilters)

	customCache := NewCache(
		WithMaxSize(50),
		WithEvictionPolicy("LRU"),
		WithTTL(2*time.Minute),
		WithCacheFilters([]string{"URL", "Method"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)
	assert.Equal(t, 50, customCache.maxSize)
	assert.Equal(t, "LRU", customCache.evictionPolicy)
	assert.Equal(t, 2*time.Minute, customCache.ttl)
	assert.Equal(t, []string{"URL", "Method"}, customCache.cacheFilters)
}

// Test key generation
func TestGenerateKey(t *testing.T) {
	cache := NewCache(WithCacheFilters([]string{"URL", "Method", "Header"}))
	req := createRequest("http://example.com", "GET", map[string]string{"X-Test-Header": "test"})
	expectedKey := "http://example.com|GET|X-Test-Header=test"
	generatedKey := cache.generateKey(req)
	assert.Equal(t, expectedKey, generatedKey)
}

// Test caching and retrieving responses
func TestCacheMiddleware(t *testing.T) {
	cache := NewCache(
		WithMaxSize(2),
		WithEvictionPolicy("FIFO"),
		WithTTL(1*time.Minute),
		WithCacheFilters([]string{"URL"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	// First request should generate a cache entry
	req := httptest.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "response", rr.Body.String())

	// Second request should be served from cache
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, "response", rr2.Body.String())
}

// Test cache expiration
func TestCacheExpiration(t *testing.T) {
	cache := NewCache(
		WithMaxSize(1),
		WithEvictionPolicy("FIFO"),
		WithTTL(1*time.Second),
		WithCacheFilters([]string{"URL"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()
	// First request should generate a cache entry
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Sleep to let the cache expire
	time.Sleep(2 * time.Second)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

// Test FIFO eviction policy
func TestFIFOCacheEviction(t *testing.T) {
	cache := NewCache(
		WithMaxSize(2),
		WithEvictionPolicy("FIFO"),
		WithTTL(1*time.Minute),
		WithCacheFilters([]string{"URL"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response-" + r.URL.Path))
	}))

	// Add two entries to the cache
	req1 := httptest.NewRequest("GET", "http://example.com/1", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	req2 := httptest.NewRequest("GET", "http://example.com/2", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// Add third entry, forcing eviction of the first one (FIFO)
	req3 := httptest.NewRequest("GET", "http://example.com/3", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)

	// First entry should be evicted
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req1)
	assert.Equal(t, http.StatusOK, rr4.Code)
	assert.Equal(t, "response-1", rr4.Body.String()) // Fresh response as it should be evicted
}

// Test LRU eviction policy
func TestLRUCacheEviction(t *testing.T) {
	cache := NewCache(
		WithMaxSize(2),
		WithEvictionPolicy("LRU"),
		WithTTL(1*time.Minute),
		WithCacheFilters([]string{"URL"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response-" + r.URL.Path))
	}))

	// Add two entries to the cache
	req1 := httptest.NewRequest("GET", "http://example.com/1", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	req2 := httptest.NewRequest("GET", "http://example.com/2", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// Access the first entry to make it recently used
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req1)
	assert.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "response-/1", rr3.Body.String())

	// Add third entry, forcing eviction of the least recently used (req2)
	req3 := httptest.NewRequest("GET", "http://example.com/3", nil)
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req3)
	assert.Equal(t, http.StatusOK, rr4.Code)

	// Second entry should be evicted
	rr5 := httptest.NewRecorder()
	handler.ServeHTTP(rr5, req2)
	assert.Equal(t, http.StatusOK, rr5.Code)
	assert.Equal(t, "response-/2", rr5.Body.String()) // Fresh response as it should be evicted
}

// Test concurrency: Ensure the cache operates correctly under concurrent access
func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(
		WithMaxSize(100),
		WithEvictionPolicy("LRU"),
		WithTTL(1*time.Minute),
		WithCacheFilters([]string{"URL"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "http://example.com/"+fmt.Sprintf("%d", n), nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}(i)
	}
	wg.Wait()
}

// Test cache with different headers, methods, and URL params
func TestCacheWithDifferentRequests(t *testing.T) {
	cache := NewCache(
		WithMaxSize(100),
		WithEvictionPolicy("LRU"),
		WithTTL(1*time.Minute),
		WithCacheFilters([]string{"Method", "Header"}),
		WithLogger(log.New(os.Stdout, "replay-test: ", log.LstdFlags)),
	)

	handler := cache.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	req := createRequest("http://example.com", "GET", map[string]string{"Content-Type": "application/json"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "response", rr.Body.String())

	req2 := createRequest("http://example.com", "POST", map[string]string{"Content-Type": "application/json"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, "response", rr2.Body.String())

	req3 := createRequest("http://example.com", "GET", map[string]string{"Content-Type": "text/plain"})
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "response", rr3.Body.String())

	req4 := createRequest("http://example.com/different", "GET", map[string]string{"Content-Type": "application/json"})
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)
	assert.Equal(t, http.StatusOK, rr4.Code)
	assert.Equal(t, "response", rr4.Body.String())
}
