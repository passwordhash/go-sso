package main

import (
    "fmt"
    "go-sso/internal/config"
)

func main() {
    cfg := config.MustLoad()

    fmt.Println("Config loaded successfully")
    fmt.Printf("cfg: %#v", cfg)
}
