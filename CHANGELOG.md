# Changelog

## [0.1.5]

### Changed

- Built with Go 1.26.8, up from 1.26.5, which closes four reachable stdlib
  vulnerabilities (GO-2026-6218 `net/url`, GO-2026-6090 `crypto/tls`,
  GO-2026-5972 `encoding/asn1`, GO-2026-5026 `net/http`).
- Updated every flokiorg dependency to its current release. `golang.org/x/text` and `golang.org/x/net`
  moved to current releases, taking govulncheck from two reachable findings to
  none.
- The release workflow now runs `go test -short`, the same command CI runs; it
  had been running the full suite, which cannot pass on a runner.
- The release now publishes a multi-arch container image to
  `ghcr.io/flokiorg/flokicoin-go-seeder`, and every push to the default branch publishes an
  `:edge` image.

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
