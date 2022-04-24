package utils

import "sync"

type Locker struct {
	lock bool
	mx   sync.RWMutex
}

func (l *Locker) Lock() {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.lock = true
}

func (l *Locker) Unlock() {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.lock = false
}

func (l *Locker) IsLocked() bool {
	return l.lock
}
