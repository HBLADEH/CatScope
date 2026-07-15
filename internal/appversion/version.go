package appversion

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var rawVersion string

// Version is read from the repository's canonical version file.
var Version = strings.TrimSpace(rawVersion)

const RepositoryURL = "https://github.com/HBLADEH/CatScope"
