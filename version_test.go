package main

import "testing"

func Test_loadVersion(t *testing.T) {
	version := loadVersion()
	versionString := version.String()
	if versionString == "" {
		t.Errorf("versionString is empty")
	}
}
