package channel

import (
	"sync"
)

// ConsSync is a synchronization handle owned by consumers that is used to
// shutdown goroutine producers.
type ConsSync struct {
	mu         sync.Mutex
	signalSent bool           // Signal channel closed?
	signalChan chan struct{}  // Channel that is closed to signal shutdown to producers
	wg         sync.WaitGroup // Waitgroup to synchronize shutdown of producers
}

// NewConsSync returns an initialized consumer shutdown handle.
func NewConsSync() *ConsSync {
	return &ConsSync{signalChan: make(chan struct{})}
}

// Add producers to wait for when shutting down.
func (cons *ConsSync) Add(delta int) {
	cons.wg.Add(delta)
}

// Shutdown shuts down goroutine producers. It returns a synchronization
// function that can be executed by the caller if shutdown was initiated for the
// first time.
func (cons *ConsSync) Shutdown() (wait func()) {
	cons.mu.Lock()
	defer cons.mu.Unlock()

	if !cons.signalSent {
		close(cons.signalChan)
		cons.signalSent = true
		wait = func() { cons.wg.Wait() }
	}

	return
}

// ProdSync returns a handle used by producers to honor the
// shutdown synchronization contract.
func (cons *ConsSync) ProdSync() ProdSync {
	return ProdSync{
		SignalChan: cons.signalChan,
		wg:         &cons.wg,
	}
}

// ProdSync is a synchronization handle used by producers to be notified
// when to perform shutdown and to signal that shutdown has been performed.
type ProdSync struct {
	SignalChan <-chan struct{} // Channel that will be closed to signal that shutdown should be performed.
	wg         *sync.WaitGroup // Waitgroup to synchronize shutdown of producers
}

// Done signals that the producer has completed shutdown.
func (prod *ProdSync) Done() {
	prod.wg.Done()
}
