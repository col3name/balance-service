package env

import (
	"flag"
	"fmt"
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func LoadDotEnvFileIfNeeded(loggerImpl logger.Logger) {
	ok := flag.Bool("load", false, "is need load .env file")
	flag.Parse()
	if *ok {
		err := godotenv.Load()
		if err != nil {
			loggerImpl.Fatal("Error loading .env file")
		}
	}
}

func ParseEnvString(key string, err error, defaultValue string) (string, error) {
	if err != nil {
		return "", err
	}
	str, ok := os.LookupEnv(key)
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue, nil
		}
		return "", fmt.Errorf("undefined environment variable %v", key)
	}
	return str, nil
}

func ParseEnvInt(key string, err error, defaultValue int) (int, error) {
	s, err := ParseEnvString(key, err, strconv.Itoa(defaultValue))
	if err != nil {
		if defaultValue != 0 {
			return defaultValue, nil
		}
		return 0, err
	}
	return strconv.Atoi(s)
}
