package livingkit

import (
	"github.com/gofrs/uuid"
)

// NewUUID4String returns a randomly generated UUID V4 string.
func NewUUID4String() string {
	return uuid.Must(uuid.NewV4()).String()
}

// IsValidUUID checks input string whether valid UUID format or not.
func IsValidUUID(input string) bool {
	_, err := uuid.FromString(input)
	return err == nil
}
