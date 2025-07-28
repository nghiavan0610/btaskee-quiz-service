package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/samber/lo"
)

func Load() {
	envFile := environmentFile()
	if envFile != "" {
		if _, err := os.Stat(envFile); err == nil {
			lo.Must0(godotenv.Load(envFile))
		}
	}
}

func environmentFile() string {
	v := os.Getenv("ENV_FILE")
	if v == "" {
		v = ".env"
	}
	fmt.Printf("Loading env file [%s] ", v)
	return v
}
