package k8s

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// ResourceType represents the type of Kubernetes resource being watched
type ResourceType int

const (
	ResourceTypePod ResourceType = iota
	ResourceTypeService
	ResourceTypeDeployment
	ResourceTypeStatefulSet
	ResourceTypeEvent
)

// String returns the string representation of the resource type
func (rt ResourceType) String() string {
	switch rt {
	case ResourceTypePod:
		return "Pod"
	case ResourceTypeService:
		return "Service"
	case ResourceTypeDeployment:
		return "Deployment"
	case ResourceTypeStatefulSet:
		return "StatefulSet"
	case ResourceTypeEvent:
		return "Event"
	default:
		return "Unknown"
	}
}

// ConnectionState represents the connection state of a watcher
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateError
)

// String returns the string representation of the connection state
func (cs ConnectionState) String() string {
	switch cs {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// WatchEvent represents a watch event from Kubernetes
type WatchEvent struct {
	ResourceType ResourceType
	EventType    watch.EventType
	Object       runtime.Object
}

// WatchError represents an error from the watch stream
type WatchError struct {
	ResourceType ResourceType
	Err          error
	Fatal        bool
}

// ResourceWatcher handles watching a single resource type
type ResourceWatcher struct {
	client          *Client
	resourceType    ResourceType
	namespace       string
	resourceVersion string
	state           ConnectionState
	backoff         *ExponentialBackoff
	debugMode       bool
	mu              sync.RWMutex
	cancelFunc      context.CancelFunc
}

// NewResourceWatcher creates a new resource watcher
func NewResourceWatcher(client *Client, resourceType ResourceType, namespace string) *ResourceWatcher {
	return &ResourceWatcher{
		client:       client,
		resourceType: resourceType,
		namespace:    namespace,
		state:        StateDisconnected,
		backoff:      NewExponentialBackoff(),
		debugMode:    false,
	}
}

// SetDebugMode enables or disables debug logging
func (rw *ResourceWatcher) SetDebugMode(enabled bool) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.debugMode = enabled
}

// Start begins watching the resource type and sends events to the provided channels
func (rw *ResourceWatcher) Start(ctx context.Context, eventChan chan<- WatchEvent, errorChan chan<- WatchError) {
	watchCtx, cancel := context.WithCancel(ctx)
	rw.mu.Lock()
	rw.cancelFunc = cancel
	rw.mu.Unlock()

	go rw.watchLoop(watchCtx, eventChan, errorChan)
}

// Stop stops the watcher
func (rw *ResourceWatcher) Stop() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.cancelFunc != nil {
		rw.cancelFunc()
		rw.cancelFunc = nil
	}
	rw.state = StateDisconnected
}

// GetState returns the current connection state
func (rw *ResourceWatcher) GetState() ConnectionState {
	rw.mu.RLock()
	defer rw.mu.RUnlock()
	return rw.state
}

// GetResourceVersion returns the current resource version
func (rw *ResourceWatcher) GetResourceVersion() string {
	rw.mu.RLock()
	defer rw.mu.RUnlock()
	return rw.resourceVersion
}

// setState updates the connection state
func (rw *ResourceWatcher) setState(state ConnectionState) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.state = state
}

// setResourceVersion updates the tracked resource version
func (rw *ResourceWatcher) setResourceVersion(version string) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.resourceVersion = version
}

// debugLog logs a message if debug mode is enabled
func (rw *ResourceWatcher) debugLog(format string, args ...interface{}) {
	rw.mu.RLock()
	debug := rw.debugMode
	rw.mu.RUnlock()

	if debug {
		msg := fmt.Sprintf("[WATCH] %s: ", rw.resourceType)
		msg += fmt.Sprintf(format, args...)
		log.Println(msg)
	}
}

// watchLoop is the main watch loop that handles connection, reconnection, and event processing
func (rw *ResourceWatcher) watchLoop(ctx context.Context, eventChan chan<- WatchEvent, errorChan chan<- WatchError) {
	for {
		select {
		case <-ctx.Done():
			rw.debugLog("Context cancelled, stopping watch")
			rw.setState(StateDisconnected)
			return
		default:
			// Attempt to connect and watch
			if err := rw.connectAndWatch(ctx, eventChan, errorChan); err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					rw.setState(StateDisconnected)
					return
				}

				// Determine if error is fatal
				fatal := isWatchErrorFatal(err)
				errorChan <- WatchError{
					ResourceType: rw.resourceType,
					Err:          err,
					Fatal:        fatal,
				}

				if fatal {
					rw.setState(StateError)
					rw.debugLog("Fatal error, stopping watch: %v", err)
					return
				}

				// Non-fatal error, attempt reconnection with backoff
				rw.setState(StateReconnecting)
				delay := rw.backoff.Next()
				rw.debugLog("Reconnecting in %v (attempt %d)", delay, rw.backoff.Attempts())

				select {
				case <-ctx.Done():
					rw.setState(StateDisconnected)
					return
				case <-time.After(delay):
					// Continue to next iteration
				}
			} else {
				// Connection closed cleanly, reconnect immediately
				rw.debugLog("Watch connection closed cleanly, reconnecting")
				rw.setState(StateConnecting)
			}
		}
	}
}

// connectAndWatch establishes the watch connection and processes events
func (rw *ResourceWatcher) connectAndWatch(ctx context.Context, eventChan chan<- WatchEvent, errorChan chan<- WatchError) error {
	// Perform initial list if we don't have a resource version
	if rw.GetResourceVersion() == "" {
		if err := rw.performInitialList(ctx, eventChan); err != nil {
			return fmt.Errorf("initial list failed: %w", err)
		}
	}

	// Start watch from current resource version
	rw.setState(StateConnecting)
	watcher, err := rw.createWatcher(ctx)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Stop()

	rw.setState(StateConnected)
	rw.backoff.Reset()
	rw.debugLog("Watch started with resourceVersion=%s", rw.GetResourceVersion())

	// Process events from the watch stream
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				// Channel closed, connection terminated
				rw.debugLog("Watch channel closed")
				return nil
			}

			if err := rw.handleWatchEvent(event, eventChan, errorChan); err != nil {
				return err
			}
		}
	}
}

// performInitialList performs a list operation to get the current resource version
func (rw *ResourceWatcher) performInitialList(ctx context.Context, eventChan chan<- WatchEvent) error {
	rw.debugLog("Performing initial list")

	switch rw.resourceType {
	case ResourceTypePod:
		list, err := rw.client.GetPods(ctx, rw.namespace)
		if err != nil {
			return err
		}
		rw.setResourceVersion(list.ResourceVersion)
		rw.debugLog("Initial list returned %d items, resourceVersion=%s", len(list.Items), list.ResourceVersion)

		// Send ADDED events for existing items
		for i := range list.Items {
			eventChan <- WatchEvent{
				ResourceType: rw.resourceType,
				EventType:    watch.Added,
				Object:       &list.Items[i],
			}
		}

	case ResourceTypeService:
		list, err := rw.client.GetServices(ctx, rw.namespace)
		if err != nil {
			return err
		}
		rw.setResourceVersion(list.ResourceVersion)
		rw.debugLog("Initial list returned %d items, resourceVersion=%s", len(list.Items), list.ResourceVersion)

		for i := range list.Items {
			eventChan <- WatchEvent{
				ResourceType: rw.resourceType,
				EventType:    watch.Added,
				Object:       &list.Items[i],
			}
		}

	case ResourceTypeDeployment:
		list, err := rw.client.GetDeployments(ctx, rw.namespace)
		if err != nil {
			return err
		}
		rw.setResourceVersion(list.ResourceVersion)
		rw.debugLog("Initial list returned %d items, resourceVersion=%s", len(list.Items), list.ResourceVersion)

		for i := range list.Items {
			eventChan <- WatchEvent{
				ResourceType: rw.resourceType,
				EventType:    watch.Added,
				Object:       &list.Items[i],
			}
		}

	case ResourceTypeStatefulSet:
		list, err := rw.client.GetStatefulSets(ctx, rw.namespace)
		if err != nil {
			return err
		}
		rw.setResourceVersion(list.ResourceVersion)
		rw.debugLog("Initial list returned %d items, resourceVersion=%s", len(list.Items), list.ResourceVersion)

		for i := range list.Items {
			eventChan <- WatchEvent{
				ResourceType: rw.resourceType,
				EventType:    watch.Added,
				Object:       &list.Items[i],
			}
		}

	case ResourceTypeEvent:
		list, err := rw.client.clientset.CoreV1().Events(rw.namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		rw.setResourceVersion(list.ResourceVersion)
		rw.debugLog("Initial list returned %d items, resourceVersion=%s", len(list.Items), list.ResourceVersion)

		for i := range list.Items {
			eventChan <- WatchEvent{
				ResourceType: rw.resourceType,
				EventType:    watch.Added,
				Object:       &list.Items[i],
			}
		}

	default:
		return fmt.Errorf("unsupported resource type: %v", rw.resourceType)
	}

	return nil
}

// createWatcher creates the appropriate watcher based on resource type
func (rw *ResourceWatcher) createWatcher(ctx context.Context) (watch.Interface, error) {
	rv := rw.GetResourceVersion()

	switch rw.resourceType {
	case ResourceTypePod:
		return rw.client.WatchPods(ctx, rw.namespace, rv)
	case ResourceTypeService:
		return rw.client.WatchServices(ctx, rw.namespace, rv)
	case ResourceTypeDeployment:
		return rw.client.WatchDeployments(ctx, rw.namespace, rv)
	case ResourceTypeStatefulSet:
		return rw.client.WatchStatefulSets(ctx, rw.namespace, rv)
	case ResourceTypeEvent:
		return rw.client.WatchEvents(ctx, rw.namespace, rv)
	default:
		return nil, fmt.Errorf("unsupported resource type: %v", rw.resourceType)
	}
}

// handleWatchEvent processes a single watch event
func (rw *ResourceWatcher) handleWatchEvent(event watch.Event, eventChan chan<- WatchEvent, errorChan chan<- WatchError) error {
	switch event.Type {
	case watch.Added, watch.Modified, watch.Deleted:
		// Update resource version from object metadata
		if err := rw.updateResourceVersionFromObject(event.Object); err != nil {
			rw.debugLog("Warning: failed to update resource version: %v", err)
		}

		rw.debugLog("Event %s - %s", event.Type, getObjectName(event.Object))

		eventChan <- WatchEvent{
			ResourceType: rw.resourceType,
			EventType:    event.Type,
			Object:       event.Object,
		}

	case watch.Error:
		// Handle error events (e.g., 410 Gone)
		statusErr, ok := event.Object.(*metav1.Status)
		if !ok {
			return fmt.Errorf("received error event with unexpected type: %T", event.Object)
		}

		// Check for 410 Gone (resourceVersion too old)
		if statusErr.Code == 410 {
			rw.debugLog("ResourceVersion expired (410 Gone), performing full re-list")
			rw.setResourceVersion("")
			return fmt.Errorf("resource version expired")
		}

		return fmt.Errorf("watch error: %s", statusErr.Message)

	case watch.Bookmark:
		// Bookmark events contain updated resource version
		if err := rw.updateResourceVersionFromObject(event.Object); err == nil {
			rw.debugLog("Bookmark received, resourceVersion=%s", rw.GetResourceVersion())
		}

	default:
		rw.debugLog("Unknown event type: %v", event.Type)
	}

	return nil
}

// updateResourceVersionFromObject extracts and updates the resource version from an object
func (rw *ResourceWatcher) updateResourceVersionFromObject(obj runtime.Object) error {
	var rv string

	switch o := obj.(type) {
	case *corev1.Pod:
		rv = o.ResourceVersion
	case *corev1.Service:
		rv = o.ResourceVersion
	case *appsv1.Deployment:
		rv = o.ResourceVersion
	case *appsv1.StatefulSet:
		rv = o.ResourceVersion
	case *corev1.Event:
		rv = o.ResourceVersion
	default:
		return fmt.Errorf("unsupported object type: %T", obj)
	}

	if rv != "" {
		rw.setResourceVersion(rv)
	}

	return nil
}

// getObjectName extracts the name from a runtime.Object for logging
func getObjectName(obj runtime.Object) string {
	switch o := obj.(type) {
	case *corev1.Pod:
		return fmt.Sprintf("%s/%s", o.Namespace, o.Name)
	case *corev1.Service:
		return fmt.Sprintf("%s/%s", o.Namespace, o.Name)
	case *appsv1.Deployment:
		return fmt.Sprintf("%s/%s", o.Namespace, o.Name)
	case *appsv1.StatefulSet:
		return fmt.Sprintf("%s/%s", o.Namespace, o.Name)
	case *corev1.Event:
		return fmt.Sprintf("%s/%s", o.Namespace, o.Name)
	default:
		return fmt.Sprintf("unknown(%T)", obj)
	}
}

// isWatchErrorFatal determines if an error should cause the watcher to stop permanently
func isWatchErrorFatal(err error) bool {
	// 401/403 are fatal - authentication/authorization errors
	if apierrors.IsUnauthorized(err) || apierrors.IsForbidden(err) {
		return true
	}

	// 404 might indicate namespace was deleted
	if apierrors.IsNotFound(err) {
		return true
	}

	// Most other errors are transient
	return false
}
