package uuid

import "github.com/gofrs/uuid"

func IsValid(val string) bool {
	_, err := FromString(val)
	return err == nil
}

func FromString(value string) (uuid.UUID, error) {
	return uuid.FromString(value)
}
