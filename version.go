package main

import "fmt"

type Version struct {
	Version   string
	GoVersion string
	BuildTime string
}

var (
	v         Version
	version   string
	goVersion string
	buildTime string
)

func init() {
	v = Version{
		Version:   version,
		GoVersion: goVersion,
		BuildTime: buildTime,
	}
}

func (ver Version) String() string {
	return fmt.Sprintf("\ngit hash: %s\nGo version: %s\nBuild time: %s", v.Version, v.GoVersion, v.BuildTime)
}
