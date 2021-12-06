package main

import (
	"intery/cmd"
	"intery/initializers"
)

func main() {
	initializers.LoadEnv()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
