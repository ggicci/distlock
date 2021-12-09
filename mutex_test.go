package distlock

import (
	"fmt"
	"testing"
	"time"
)

func runBasicLockTests(t *testing.T, provider Provider) {
	var (
		err error
		m   = New(provider, WithLockLifetime(1*time.Second)).
			New("johndoe", WithNamespace("questions"))
	)
	expectedMutexDisplayName := fmt.Sprintf("Mutex(%s:questions:johndoe)", provider.Name())
	gotMutexDisplayName := m.String()
	if gotMutexDisplayName != expectedMutexDisplayName {
		t.Errorf("Mutex display name incorrect. Expected: %s, got: %s",
			expectedMutexDisplayName,
			gotMutexDisplayName,
		)
	}

	err = m.Lock()
	if err != nil {
		t.Errorf("Lock failed: name=%s, err=%v", m, err)
	}
	err = m.Unlock()
	if err != nil {
		t.Errorf("Unlock failed: name=%s, err=%v", m, err)
	}

	err = m.Lock()
	if err != nil {
		t.Errorf("Lock failed: name=%s, err=%v", m, err)
	}

	// Here unlock should fail because lock was released by lifetime.
	time.Sleep(1200 * time.Millisecond)
	err = m.Unlock()
	if err == nil {
		t.Errorf("unlock should fail")
	}

	m.Lock()
	// Here lock should fail.
	err = m.Lock()
	if err == nil {
		t.Errorf("lock should fail")
	}
	err = m.Unlock()
	if err != nil {
		t.Errorf("Unlock failed: name=%s, err=%v", m, err)
	}

	m.Lock()
	m.(*mutex).id = "2814"
	err = m.Unlock()
	if err == nil {
		t.Errorf("unlock should fail")
	}
}
