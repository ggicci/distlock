package distlock

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runBasicLockTests(t *testing.T, provider Provider) {
	factory := New(provider, WithLockLifetime(1*time.Second))
	m := factory.New("johndoe", WithNamespace("questions"))
	expectedMutexDisplayName := fmt.Sprintf("Mutex(%s:questions:johndoe)", provider.Name())
	gotMutexDisplayName := m.String()
	if gotMutexDisplayName != expectedMutexDisplayName {
		t.Errorf("Mutex display name incorrect. Expected: %s, got: %s",
			expectedMutexDisplayName,
			gotMutexDisplayName,
		)
	}

	testLockAndUnlockInTime(t, m)
	testUnlockAfterAnExpiredLock(t, m)
	testLockContention(t, m)

	m1 := factory.New("apple", WithLockLifetime(10*time.Millisecond))
	m2 := factory.New("apple", WithLockLifetime(100*time.Millisecond))
	testUnlockAfterOwnerChange(t, m1, m2)
}

func testLockAndUnlockInTime(t *testing.T, m Mutex) {
	assert.NoError(t, m.Lock())
	assert.NoError(t, m.Unlock())
	// expectLocked(t, m.Lock())
	// expectUnlocked(t, m.Unlock())
}

func testUnlockAfterAnExpiredLock(t *testing.T, m Mutex) {
	assert.NoError(t, m.Lock())
	time.Sleep(1200 * time.Millisecond) // expired (released by system)
	assert.ErrorIs(t, m.Unlock(), ErrNotLocked)
}

func testLockContention(t *testing.T, m Mutex) {
	assert.NoError(t, m.Lock())
	assert.ErrorIs(t, m.Lock(), ErrAlreadyLocked)
	assert.NoError(t, m.Unlock())
}

func testUnlockAfterOwnerChange(t *testing.T, m1, m2 Mutex) {
	assert.NoError(t, m1.Lock())
	assert.ErrorIs(t, m2.Lock(), ErrAlreadyLocked)
	time.Sleep(10 * time.Millisecond) // m1 expired (released by system)
	assert.NoError(t, m2.Lock())      // m2 can obtain the lock, since m1 is expired
	assert.ErrorIs(t, m1.Unlock(), ErrNotLocked)
}
