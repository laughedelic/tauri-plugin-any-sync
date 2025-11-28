//go:build tools
// +build tools

package mobile

// This file is used to ensure golang.org/x/mobile is in go.mod
// It's not compiled, but helps go mod tidy keep the dependency
import _ "golang.org/x/mobile/bind"
