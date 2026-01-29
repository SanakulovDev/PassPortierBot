package vault

import (
	"sync"
	"time"
)

var (
	sessions = make(map[int64][]byte)
	mu       sync.RWMutex
)

func SetKey(userID int64, key []byte) {
	mu.Lock()
	defer mu.Unlock()
	sessions[userID] = key

	// 30 daqiqadan keyin avtomatik o'chirish
	time.AfterFunc(30*time.Minute, func() {
		mu.Lock()
		delete(sessions, userID)
		mu.Unlock()
	})
}

func GetKey(userID int64) ([]byte, bool) {
	mu.RLock()
	defer mu.RUnlock()
	key, ok := sessions[userID]
	return key, ok
}