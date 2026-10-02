# Changelog

## [0.1.5]

### Fixed

- The seeder reported version 0.1.2 on its status page and advertised `0.1` as
  its user agent version to every peer it crawled -- two separate hard-coded
  strings, both well behind the released version. Both now come from one value
  taken from the release tag at build time.

### Changed

- Built with Go 1.26.5. (#3)

## [0.1.4-beta]

### Fixes

- **`http.go`**: 7 `fmt.Fprintf(w, tN)` calls passed a fixed template string directly as the format string instead of the data. Switched to `fmt.Fprint`.
- **`cloudflare_test.go`**: removed a hardcoded, live Cloudflare API token from the source (already rotated and purged from git history in a prior, separate step). Now reads from `CLOUDFLARE_API_TOKEN` and skips the test when unset.
- **`crawler_test.go`**: removed a bare `select {}` at the end of `TestCrawlIP` (an unconditional infinite block); switched `log.Fatal`/`log.Fatalf` to `t.Fatal`/`t.Fatalf` in `TestCrawler` so a failure doesn't kill the whole test binary. Both tests are gated under `testing.Short()` since they dial a hardcoded external node IP.

### CI

- Added `.github/workflows/ci.yaml`: runs `go build`, `go vet`, and `go test -short` on push to `main` and on pull requests.

## [0.1.3-beta]

### Dependency Updates

- Updated dependencies to align with `go-flokicoin v0.25.13-alpha`.
- Routine `go mod tidy` cleanup.
