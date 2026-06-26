//go:build !wasip1

// This file exists only so the package builds on the host platform. The real
// plugin is in main.go and is compiled with GOOS=wasip1 GOARCH=wasm.
package main

func main() {}
