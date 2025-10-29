package k8s

import (
	"context"
	"fmt"
	"sync"
)

// WatchManager orchestrates multiple resource watchers and provides a unified interface
type WatchManager struct {
	client      *Client
	watchers    map[ResourceType]*ResourceWatcher
	cancelFuncs map[ResourceType]context.CancelFunc
	eventChan   chan WatchEvent
	errorChan   chan WatchError
	mu          sync.RWMutex
	ctx         context.Context
	cancelAll   context.CancelFunc
	debugMode   bool
}

// NewWatchManager creates a new watch manager
func NewWatchManager(client *Client) *WatchManager {
	return &WatchManager{
		client:      client,
		watchers:    make(map[ResourceType]*ResourceWatcher),
		cancelFuncs: make(map[ResourceType]context.CancelFunc),
		eventChan:   make(chan WatchEvent, 100), // Buffer to avoid blocking watchers
		errorChan:   make(chan WatchError, 100),
		debugMode:   false,
	}
}

// SetDebugMode enables or disables debug logging for all watchers
func (wm *WatchManager) SetDebugMode(enabled bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.debugMode = enabled

	// Propagate to existing watchers
	for _, w := range wm.watchers {
		w.SetDebugMode(enabled)
	}
}

// Start begins watching the specified resource types
// This spawns a goroutine for each resource type
func (wm *WatchManager) Start(ctx context.Context, resourceTypes []ResourceType) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Create cancellable context for all watchers
	wm.ctx, wm.cancelAll = context.WithCancel(ctx)

	// Start a watcher for each resource type
	for _, rt := range resourceTypes {
		if err := wm.startWatcherLocked(rt); err != nil {
			// If any watcher fails to start, stop all and return error
			wm.stopAllLocked()
			return fmt.Errorf("failed to start watcher for %s: %w", rt, err)
		}
	}

	return nil
}

// startWatcherLocked starts a single watcher (must be called with lock held)
func (wm *WatchManager) startWatcherLocked(rt ResourceType) error {
	// Check if already watching
	if _, exists := wm.watchers[rt]; exists {
		return fmt.Errorf("already watching %s", rt)
	}

	// Create watcher
	watcher := NewResourceWatcher(wm.client, rt, wm.client.GetNamespace())
	watcher.SetDebugMode(wm.debugMode)

	// Create context for this specific watcher
	watcherCtx, watcherCancel := context.WithCancel(wm.ctx)

	// Store watcher and cancel function
	wm.watchers[rt] = watcher
	wm.cancelFuncs[rt] = watcherCancel

	// Start the watcher
	watcher.Start(watcherCtx, wm.eventChan, wm.errorChan)

	return nil
}

// Stop gracefully stops all watchers
func (wm *WatchManager) Stop() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.stopAllLocked()
}

// stopAllLocked stops all watchers (must be called with lock held)
func (wm *WatchManager) stopAllLocked() {
	// Cancel all watchers
	if wm.cancelAll != nil {
		wm.cancelAll()
	}

	// Stop each watcher individually
	for _, watcher := range wm.watchers {
		watcher.Stop()
	}

	// Clear maps
	wm.watchers = make(map[ResourceType]*ResourceWatcher)
	wm.cancelFuncs = make(map[ResourceType]context.CancelFunc)
}

// RestartWatcher restarts a specific resource type watcher
// Useful for manual recovery or namespace switching
func (wm *WatchManager) RestartWatcher(resourceType ResourceType) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Stop existing watcher if present
	if watcher, exists := wm.watchers[resourceType]; exists {
		watcher.Stop()
		if cancel, ok := wm.cancelFuncs[resourceType]; ok {
			cancel()
		}
		delete(wm.watchers, resourceType)
		delete(wm.cancelFuncs, resourceType)
	}

	// Start new watcher
	return wm.startWatcherLocked(resourceType)
}

// RestartAll restarts all watchers (e.g., after namespace change)
func (wm *WatchManager) RestartAll() error {
	wm.mu.Lock()

	// Get current resource types
	resourceTypes := make([]ResourceType, 0, len(wm.watchers))
	for rt := range wm.watchers {
		resourceTypes = append(resourceTypes, rt)
	}

	// Stop all watchers
	wm.stopAllLocked()
	wm.mu.Unlock()

	// Restart with new namespace
	if wm.ctx != nil {
		return wm.Start(wm.ctx, resourceTypes)
	}

	return fmt.Errorf("watch manager not started")
}

// GetEventChannel returns the channel for receiving watch events
func (wm *WatchManager) GetEventChannel() <-chan WatchEvent {
	return wm.eventChan
}

// GetErrorChannel returns the channel for receiving watch errors
func (wm *WatchManager) GetErrorChannel() <-chan WatchError {
	return wm.errorChan
}

// GetConnectionStates returns the current connection state of all watchers
func (wm *WatchManager) GetConnectionStates() map[ResourceType]ConnectionState {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	states := make(map[ResourceType]ConnectionState)
	for rt, watcher := range wm.watchers {
		states[rt] = watcher.GetState()
	}
	return states
}

// GetOverallConnectionState returns an aggregate connection state
// If any watcher is in error state, returns StateError
// If any watcher is reconnecting, returns StateReconnecting
// If all watchers are connected, returns StateConnected
// Otherwise returns StateDisconnected
func (wm *WatchManager) GetOverallConnectionState() ConnectionState {
	states := wm.GetConnectionStates()

	if len(states) == 0 {
		return StateDisconnected
	}

	hasError := false
	hasReconnecting := false
	hasConnecting := false
	connectedCount := 0

	for _, state := range states {
		switch state {
		case StateError:
			hasError = true
		case StateReconnecting:
			hasReconnecting = true
		case StateConnecting:
			hasConnecting = true
		case StateConnected:
			connectedCount++
		}
	}

	// Priority: Error > Reconnecting > Connecting > Connected > Disconnected
	if hasError {
		return StateError
	}
	if hasReconnecting {
		return StateReconnecting
	}
	if hasConnecting {
		return StateConnecting
	}
	if connectedCount == len(states) {
		return StateConnected
	}

	return StateDisconnected
}

// GetWatcherCount returns the number of active watchers
func (wm *WatchManager) GetWatcherCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.watchers)
}

// IsWatching returns true if watching the specified resource type
func (wm *WatchManager) IsWatching(resourceType ResourceType) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	_, exists := wm.watchers[resourceType]
	return exists
}

// UpdateNamespace updates the namespace for all watchers and restarts them
func (wm *WatchManager) UpdateNamespace(namespace string) error {
	wm.mu.Lock()
	wm.client.SetNamespace(namespace)
	wm.mu.Unlock()

	// Restart all watchers with new namespace
	return wm.RestartAll()
}

// GetResourceVersions returns the current resource version for each watcher
func (wm *WatchManager) GetResourceVersions() map[ResourceType]string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	versions := make(map[ResourceType]string)
	for rt, watcher := range wm.watchers {
		versions[rt] = watcher.GetResourceVersion()
	}
	return versions
}
