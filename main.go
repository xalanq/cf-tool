package main

import (
	"cf-tool/client"
	"cf-tool/config"
	"fmt"
)

func gg() {
	cln := client.New(config.ConfigPath)
	handles, err := cln.ParseHandles()
	if err != nil {
		panic(err)
	}
	for _, handle := range handles {
		fmt.Println(handle.Color, handle.Handle)
	}
}
