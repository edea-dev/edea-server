package backend

import "embed"

// BootstrapIcons holds the bootstrap-icons svg files to render them directly into the
// templates. it only increases the binary size by ~1MB.
//go:embed static/icons/*
var BootstrapIcons embed.FS
