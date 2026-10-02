package main

// appVersion is the application version, set at release-build time with
// '-ldflags "-X main.appVersion=0.1.5"' from the release tag. It is the only
// place the version is recorded, so what this binary reports cannot drift from
// what was actually published. A plain `go build` leaves it empty and
// devVersion is reported instead.
var appVersion string

// devVersion is reported by builds that had no version injected. It is
// deliberately not a real version number, so a development build is never
// mistaken for a release.
const devVersion = "0.0.0-dev"

// version returns the effective application version. It is reported on the
// status page and advertised as this seeder's user agent version to the peers
// it crawls, so both come from the same value.
func version() string {
	if appVersion != "" {
		return appVersion
	}

	return devVersion
}
