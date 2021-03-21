package models

import (
	"time"
)

// Cache is an abstraction of the caching layer.
type Cache interface {
	Get(key string) []byte
	Set(key string, content []byte, duration time.Duration) error
}
