//go:build wasip1

package main

import panel "github.com/SteelDrEgg/coreplugins/pluginsdk/wasm/proto"

func main() {}

func init() {
	panel.RegisterPlugin(&secretManagerPlugin{})
}
