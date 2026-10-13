package terminal_session

import (
	"sync"
	"time"
)

type RevokeFunc func(scope, userID, authSessionID string) error

type resumableRevocation struct {
	err   error
	retry RevokeFunc
}

type revokeTarget struct{ scope, userID, authSessionID string }
type revokeRetry struct {
	target   revokeTarget
	revoke   RevokeFunc
	next     time.Time
	deadline time.Time
	attempt  int
	version  uint64
}

type revocationQueue struct {
	mu      sync.Mutex
	pending map[revokeTarget]revokeRetry
	limit   int
	version uint64
}

var terminalRevocations = revocationQueue{pending: make(map[revokeTarget]revokeRetry), limit: 256}
var startRevocationWorker sync.Once
