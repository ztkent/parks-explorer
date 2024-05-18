package replay

import (
	"bytes"
	"container/list"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Cache represents the configurable in-memory request cache
type Cache struct {
	maxSize        int
	maxMemory      uint64
	evictionPolicy string
	evictionTimer  time.Duration
	ttl            time.Duration
	cacheMap       map[string]*list.Element
	cacheList      *list.List
	cacheFilters   []string
	mut            sync.Mutex
	l              *log.Logger
}

// CacheEntry stores a single cache item
type CacheEntry struct {
	key          string
	value        *http.Response
	expiration   time.Time
	lastAccessed time.Time
}

const (
	DefaultMaxSize        = 5
	DefaultMaxMemory      = 10 * 1024 * 1024 // 10 MB
	DefaultEvictionPolicy = "FIFO"
	DefaultEvictionTimer  = 1 * time.Minute
	DefaultTTL            = 5 * time.Minute
	DefaultFilter         = "URL"
)

type CacheOption func(*Cache)

// NewCache initializes a new instance of Cache with given options
func NewCache(options ...CacheOption) *Cache {
	c := &Cache{
		maxSize:        DefaultMaxSize,
		maxMemory:      DefaultMaxMemory,
		evictionPolicy: DefaultEvictionPolicy,
		ttl:            DefaultTTL,
		cacheMap:       make(map[string]*list.Element),
		cacheList:      list.New(),
		cacheFilters:   []string{DefaultFilter},
		evictionTimer:  DefaultEvictionTimer,
		l:              log.New(io.Discard, "", 0),
	}
	for _, option := range options {
		option(c)
	}
	go c.clearExpiredEntries()
	return c
}

func WithMaxSize(maxSize int) CacheOption {
	return func(c *Cache) {
		if maxSize != 0 {
			c.maxSize = maxSize
		}
	}
}

func WithMaxMemory(maxMemory uint64) CacheOption {
	return func(c *Cache) {
		if maxMemory != 0 {
			c.maxMemory = maxMemory
		}
	}
}

func WithEvictionPolicy(evictionPolicy string) CacheOption {
	return func(c *Cache) {
		if evictionPolicy != "" {
			c.evictionPolicy = evictionPolicy
		}
	}
}

func WithEvictionTimer(evictionTimer time.Duration) CacheOption {
	return func(c *Cache) {
		if evictionTimer > time.Minute {
			c.evictionTimer = evictionTimer
		}
	}
}

func WithTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) {
		if ttl != 0 {
			c.ttl = ttl
		}
	}
}

func WithCacheFilters(cacheFilters []string) CacheOption {
	return func(c *Cache) {
		if len(cacheFilters) != 0 {
			c.cacheFilters = cacheFilters
		}
	}
}

func WithLogger(l *log.Logger) CacheOption {
	return func(c *Cache) {
		c.l = l
	}
}

// Middleware function to intercept HTTP requests and interact with the cache
func (c *Cache) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		key := c.generateKey(r)
		wasCached := false
		defer func() {
			c.l.Printf("Request: %s, Cached: %v, Duration: %v", key, wasCached, time.Since(start))
		}()

		c.mut.Lock()
		if ele, found := c.cacheMap[key]; found {
			entry := ele.Value.(*CacheEntry)
			if entry.expiration.After(time.Now()) {
				// Serve from cache
				c.l.Printf("Serving from cache: %s", key)
				c.serveFromCache(w, entry)
				c.mut.Unlock()
				wasCached = true
				return
			}
			// Expired, remove the item
			c.l.Printf("Cache entry expired: %s, removing from cache", key)
			c.cacheList.Remove(ele)
			delete(c.cacheMap, key)
		}
		c.mut.Unlock()
		// Not in cache or expired, serve from next handler and cache the response
		c.l.Printf("Cache miss: %s", key)
		responseRecorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(responseRecorder, r)
		if responseRecorder.statusCode == http.StatusOK { // Cache only successful responses
			c.mut.Lock()
			c.l.Printf("Adding to cache: %s", key)
			c.addToCache(key, responseRecorder.Result())
			c.mut.Unlock()
		}
	})
}

func (c *Cache) clearExpiredEntries() {
	timer := time.NewTicker(c.evictionTimer)
	for range timer.C {
		c.mut.Lock()
		for ele := c.cacheList.Front(); ele != nil; ele = ele.Next() {
			entry := ele.Value.(*CacheEntry)
			if entry.expiration.Before(time.Now()) {
				c.l.Printf("Cache entry expired: %s, removing from cache", entry.key)
				delete(c.cacheMap, entry.key)
				c.cacheList.Remove(ele)
			}
		}
		c.mut.Unlock()
	}
}

// Creates the cache key based on the request and cache filters
func (c *Cache) generateKey(r *http.Request) string {
	keyParts := []string{r.URL.String()}
	for _, filter := range c.cacheFilters {
		switch filter {
		case "Method":
			keyParts = append(keyParts, r.Method)
		case "Header":
			for k, v := range r.Header {
				keyParts = append(keyParts, k+"="+v[0])
			}
		}
	}
	return strings.Join(keyParts, "|")
}

// Serve the response from cache
func (c *Cache) serveFromCache(w http.ResponseWriter, entry *CacheEntry) {
	// Update last accessed time
	entry.lastAccessed = time.Now()

	// Write cached response to the client
	for k, v := range entry.value.Header {
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(entry.value.StatusCode)
	io.Copy(w, io.NopCloser(entry.value.Body))
}

// Add a new response to the cache
func (c *Cache) addToCache(key string, resp *http.Response) {
	c.checkNecessaryEvictions()
	entry := &CacheEntry{
		key:          key,
		value:        cloneResponse(resp),
		expiration:   time.Now().Add(c.ttl),
		lastAccessed: time.Now(),
	}
	c.cacheList.PushFront(entry)
	c.cacheMap[key] = c.cacheList.Front()
}

// stay in bounds of cache size and memory limits
func (c *Cache) checkNecessaryEvictions() {
	for c.cacheList.Len() >= c.maxSize {
		c.l.Printf("Cache is full, evicting an item")
		c.evict()
	}
	c.checkMemoryLimit()
}

// checkMemoryLimit ensures cache memory usage is within the specified maxMemory limit
func (c *Cache) checkMemoryLimit() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := m.Alloc

	for currentMemory > c.maxMemory {
		c.l.Printf("Cache exceeds max memory, evicting an item")
		c.evict()
		runtime.ReadMemStats(&m)
		currentMemory = m.Alloc
	}
}

// Evict based on policy
func (c *Cache) evict() {
	var ele *list.Element
	if c.evictionPolicy == "FIFO" {
		ele = c.cacheList.Back()
	} else if c.evictionPolicy == "LRU" {
		var oldest time.Time
		for e := c.cacheList.Front(); e != nil; e = e.Next() {
			entry := e.Value.(*CacheEntry)
			if oldest.IsZero() || entry.lastAccessed.Before(oldest) {
				oldest = entry.lastAccessed
				ele = e
			}
		}
	}
	if ele != nil {
		entry := ele.Value.(*CacheEntry)
		c.l.Printf("Evicting: %v", entry.key)
		c.cacheList.Remove(ele)
		delete(c.cacheMap, entry.key)
	}
}

// Clone an HTTP response for caching
func cloneResponse(resp *http.Response) *http.Response {
	var buf bytes.Buffer
	if resp.Body != nil {
		io.Copy(&buf, resp.Body)
		resp.Body.Close()
	}
	return &http.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Header:        resp.Header,
		Body:          io.NopCloser(&buf),
		ContentLength: int64(buf.Len()),
	}
}

// Capture our response on the fly, lets us cache it.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       io.ReadWriter
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.body == nil {
		r.body = new(bytes.Buffer)
	}
	r.body.Write(body)
	return r.ResponseWriter.Write(body)
}

func (r *responseRecorder) Header() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

func (r *responseRecorder) Result() *http.Response {
	return &http.Response{
		StatusCode:    r.statusCode,
		Header:        r.headers,
		Body:          io.NopCloser(bytes.NewBuffer(r.body.(*bytes.Buffer).Bytes())),
		ContentLength: int64(r.body.(*bytes.Buffer).Len()),
	}
}
