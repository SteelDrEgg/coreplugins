//go:build wasip1

package main

import arupawasm "github.com/SteelDrEgg/arupa-sdk/golang/wasm"

func main() {}

func init() {
	manager := newSecretManagerPlugin()
	arupawasm.RegisterService(manager.sdk)
}
