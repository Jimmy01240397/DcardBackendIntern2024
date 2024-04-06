// +build !release

package config

import (
    "github.com/joho/godotenv"
)

func loadenv() {
    _ = godotenv.Load()
}
