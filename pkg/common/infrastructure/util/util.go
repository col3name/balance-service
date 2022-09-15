package util

import b64 "encoding/base64"

func B64encode(value string) string {
	return b64.StdEncoding.EncodeToString([]byte(value))
}

func B64decode(value string) ([]byte, error) {
	return b64.StdEncoding.DecodeString(value)
}
