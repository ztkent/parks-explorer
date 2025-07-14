module github.com/ztkent/nps-dashboard

go 1.24

toolchain go1.24.3

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/httprate v0.9.0
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/rs/cors v1.11.0
	github.com/ztkent/go-nps v1.0.1
	github.com/ztkent/replay v1.0.2
	golang.org/x/oauth2 v0.30.0
)

require (
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/ztkent/go-nps => ../go-nps
