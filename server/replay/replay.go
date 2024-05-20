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

/*
Configurable request caching middleware for Go servers.
*/

type Cache struct {
	maxSize        int
	maxMemory      uint64
	evictionPolicy string
	evictionTimer  time.Duration
	ttl            time.Duration
	maxTtl         time.Duration
	cacheMap       map[string]*list.Element
	cacheList      *list.List
	cacheFilters   []string
	mut            sync.Mutex
	l              *log.Logger
}

// CacheEntry stores a single cache item
type CacheEntry struct {
	key          string
	value        *CacheResponse
	created      time.Time
	lastAccessed time.Time
}

// CacheResponse stores the response info to be cached
type CacheResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

const (
	DefaultMaxSize        = 25                // Maximum number of entries in the cache
	DefaultMaxMemory      = 100 * 1024 * 1024 // 100 MB
	DefaultEvictionPolicy = "FIFO"            // First In First Out
	DefaultEvictionTimer  = 1 * time.Minute   // The time between cache eviction checks
	DefaultTTL            = 5 * time.Minute   // The time a cache entry can live without being accessed
	DefaultMaxTTL         = 10 * time.Minute  // The maximum time a cache entry can live, including renewals
	DefaultFilter         = "URL"             // Cache requests based on URL
)

type CacheOption func(*Cache)

// NewCache initializes a new instance of Cache with given options
func NewCache(options ...CacheOption) *Cache {
	c := &Cache{
		maxSize:        DefaultMaxSize,
		maxMemory:      DefaultMaxMemory,
		evictionPolicy: DefaultEvictionPolicy,
		ttl:            DefaultTTL,
		maxTtl:         DefaultMaxTTL,
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

// Set the maximum number of entries in the cache
func WithMaxSize(maxSize int) CacheOption {
	return func(c *Cache) {
		if maxSize != 0 {
			c.maxSize = maxSize
		}
	}
}

// Set the maximum memory usage of the cache
func WithMaxMemory(maxMemory uint64) CacheOption {
	return func(c *Cache) {
		if maxMemory != 0 {
			c.maxMemory = maxMemory
		}
	}
}

// Set the eviction policy for the cache [FIFO, LRU]
func WithEvictionPolicy(evictionPolicy string) CacheOption {
	return func(c *Cache) {
		if evictionPolicy != "" {
			c.evictionPolicy = evictionPolicy
		}
	}
}

// Set the time between cache eviction checks
func WithEvictionTimer(evictionTimer time.Duration) CacheOption {
	return func(c *Cache) {
		if evictionTimer > time.Minute {
			c.evictionTimer = evictionTimer
		}
	}
}

// Set the time a cache entry can live without being accessed
func WithTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) {
		if ttl > 0 {
			c.ttl = ttl
		}
	}
}

// Set the maximum time a cache entry can live, including renewals
func WithMaxTTL(maxTtl time.Duration) CacheOption {
	return func(c *Cache) {
		if maxTtl > c.ttl {
			c.maxTtl = maxTtl
		}
	}
}

// Set the cache filters to use for generating cache keys
func WithCacheFilters(cacheFilters []string) CacheOption {
	return func(c *Cache) {
		if len(cacheFilters) != 0 {
			c.cacheFilters = cacheFilters
		}
	}
}

// Set the logger to use for cache logging
func WithLogger(l *log.Logger) CacheOption {
	return func(c *Cache) {
		c.l = l
	}
}

// Middleware function to intercept HTTP requests and interact with the cache
func (c *Cache) Middleware(next http.HandlerFunc) http.HandlerFunc {
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
			if entry.lastAccessed.Add(c.ttl).After(time.Now()) {
				// valid, not expired
				c.l.Printf("Serving from cache: %s", key)
				c.serveFromCache(w, entry)
				c.mut.Unlock()
				wasCached = true
				return
			}
			// accessed expired item, remove the item
			c.l.Printf("Cache entry expired: %s, removing from cache", key)
			c.cacheList.Remove(ele)
			delete(c.cacheMap, key)
		} else {
			c.l.Printf("Cache miss: %s", key)
		}
		c.mut.Unlock()
		// not in cache or expired, serve then cache the response
		wr := &writerRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next(wr, r)
		if wr.statusCode == http.StatusOK {
			// cache only the successful responses to prevent chaos
			go c.addToCache(key, wr.Result())
		}
	})
}

// clear expired cache entries that have not been accessed
func (c *Cache) clearExpiredEntries() {
	timer := time.NewTicker(c.evictionTimer)
	for range timer.C {
		c.mut.Lock()
		for ele := c.cacheList.Front(); ele != nil; ele = ele.Next() {
			entry := ele.Value.(*CacheEntry)
			if entry.lastAccessed.Add(c.ttl).Before(time.Now()) ||
				entry.created.Add(c.maxTtl).Before(time.Now()) {
				c.l.Printf("Cache entry expired: %s, removing from cache", entry.key)
				delete(c.cacheMap, entry.key)
				c.cacheList.Remove(ele)
			}
		}
		c.mut.Unlock()
	}
}

// generate cache key based on request + selected filters
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

// serve the response from cache
func (c *Cache) serveFromCache(w http.ResponseWriter, entry *CacheEntry) {
	entry.lastAccessed = time.Now()
	for k, v := range entry.value.Header {
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(entry.value.StatusCode)
	w.Write(entry.value.Body)
}

// add a new response to the cache
func (c *Cache) addToCache(key string, resp *http.Response) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.checkNecessaryEvictions()
	entry := &CacheEntry{
		key:          key,
		value:        cloneResponse(resp),
		created:      time.Now(),
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

// validate cache memory usage is within the specified maxMemory limit
func (c *Cache) checkMemoryLimit() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := m.Alloc

	for currentMemory > c.maxMemory && c.cacheList.Len() != 0 {
		c.l.Printf("Cache exceeds max memory, evicting an item")
		c.evict()
		runtime.ReadMemStats(&m)
		currentMemory = m.Alloc
	}
}

// evict from cache based on policy
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

// copy the response for caching
func cloneResponse(resp *http.Response) *CacheResponse {
	var buf bytes.Buffer
	if resp.Body != nil {
		io.Copy(&buf, resp.Body)
		resp.Body.Close()
	}
	return &CacheResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       buf.Bytes(),
	}
}

// Capture our response on the fly, lets us cache it.
type writerRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       io.ReadWriter
}

func (wr *writerRecorder) WriteHeader(statusCode int) {
	wr.statusCode = statusCode
	wr.ResponseWriter.WriteHeader(statusCode)
}

func (wr *writerRecorder) Write(body []byte) (int, error) {
	if wr.body == nil {
		wr.body = new(bytes.Buffer)
	}
	wr.body.Write(body)
	return wr.ResponseWriter.Write(body)
}

func (wr *writerRecorder) Header() http.Header {
	if wr.headers == nil {
		wr.headers = make(http.Header)
	}
	return wr.headers
}

func (wr *writerRecorder) Result() *http.Response {
	return &http.Response{
		StatusCode:    wr.statusCode,
		Header:        wr.headers,
		Body:          io.NopCloser(bytes.NewBuffer(wr.body.(*bytes.Buffer).Bytes())),
		ContentLength: int64(wr.body.(*bytes.Buffer).Len()),
	}
}
