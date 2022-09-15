package env

import (
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/money/app/log"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func LoadDotEnvFileIfNeeded(loggerImpl log.Logger) {
	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
}

func ParseEnvString(key string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("undefined environment variable %v", key)
	}
	return str, nil
}

func ParseEnvInt(key string, err error) (int, error) {
	s, err := ParseEnvString(key, err)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}
