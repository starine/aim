package container

import (
	"github.com/starine/aim"
	"github.com/starine/aim/wire/pkt"
)

// HashSelector HashSelector
type HashSelector struct {
}

// Lookup a server
func (s *HashSelector) Lookup(header *pkt.Header, srvs []kim.Service) string {
	ll := len(srvs)
	code := HashCode(header.ChannelId)
	return srvs[code%ll].ServiceID()
}
