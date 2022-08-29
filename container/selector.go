package container

import (
	"hash/crc32"

	"github.com/starine/aim"
	"github.com/starine/aim/wire/pkt"
)

// HashCode generated a hash code
func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}

// Selector is used to select a Service
type Selector interface {
	Lookup(*pkt.Header, []aim.Service) string
}
