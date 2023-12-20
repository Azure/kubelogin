package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// gitTag provides the git tag used to build this binary.
// This is set via ldflags at build time, which normally set by the release pipeline.
// For go install binary, this value stays empty.
var gitTag string

type Version struct {
	Version   string
	GoVersion string
	BuildTime string
	Platform  string
}

func loadVersion() Version {
	rv := Version{
		Version:   "unknown",
		GoVersion: "unknown",
		BuildTime: "unknown",
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if gitTag != "" {
		rv.Version = gitTag
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return rv
	}

	rv.GoVersion = buildInfo.GoVersion

	var (
		modified  bool
		revision  string
		buildTime string
	)
	for _, s := range buildInfo.Settings {
		if s.Value == "" {
			continue
		}

		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		case "vcs.time":
			buildTime = s.Value
		}
	}

	// in Go install mode, this is a known issue that vcs information will not be available.
	// ref: https://github.com/golang/go/issues/51279
	// Fallback to use module version and stop here as vcs information is incomplete.
	if revision == "" {
		if buildInfo.Main.Version != "" {
			// fallback to use module version (legacy usage)
			rv.Version = buildInfo.Main.Version
		}

		return rv
	}

	if modified {
		revision += "-dirty"
	}
	if gitTag != "" {
		revision = gitTag + "/" + revision
	}
	rv.Version = revision

	if buildTime != "" {
		rv.BuildTime = buildTime
	}

	return rv
}

func (ver Version) String() string {
	return fmt.Sprintf(
		"\ngit hash: %s\nGo version: %s\nBuild time: %s\nPlatform: %s",
		ver.Version,
		ver.GoVersion,
		ver.BuildTime,
		ver.Platform,
	)
}
