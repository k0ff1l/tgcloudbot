package main

import (
	"fmt"

	"github.com/k0ff1l/tgcloudbot/internal/config"
)

func main() {
	cfg := config.New()

	fmt.Println(cfg)
}
