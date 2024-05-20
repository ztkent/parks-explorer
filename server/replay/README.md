# Replay
Configurable request caching middleware for Go servers.   
Improve API throughput, reduce latency, and save on third-party resources.   

- Eviction policy: [FIFO, LRU]
- Cache Size: [Entries, Memory Usage]
- Filter Options: [URL, Method, Header]
- TTL + Max TTL: [Includes renewals]

## Usage
```go
    import (
        "log"
        "net/http"
        "os"
        "time"

        "github.com/Ztkent/replay"
        "github.com/go-chi/chi/v5"
    )

    func main() {
        r := chi.NewRouter()
        // Initialize a new Cache with given options
        c := replay.NewCache(
            replay.WithMaxSize(100), // Set the maximum number of entries in the cache
            replay.WithMaxMemory(100*1024*1024), // Set the maximum memory usage of the cache
            replay.WithCacheFilters([]string{"URL", "Method"}), // Set the cache filters to use for generating cache keys
            replay.WithEvictionPolicy("LRU"), // Set the eviction policy for the cache [FIFO, LRU]
            replay.WithTTL(5*time.Minute), // Set the time a cache entry can live without being accessed
            replay.WithMaxTTL(30*time.Minute), // Set the maximum time a cache entry can live, including renewals
            replay.WithLogger(log.New(os.Stdout, "replay: ", log.LstdFlags)), // Set the logger to use for cache logging
        )
        r.Get("/test-endpoint", c.Middleware(testHandlerFunc()))
        http.ListenAndServe(os.Getenv("SERVER_PORT"), r)
    }
```