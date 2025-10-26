package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes clientset with additional context
type Client struct {
	clientset      *kubernetes.Clientset
	config         *rest.Config
	namespace      string
	currentContext string
	contexts       []string
}

// NewClient creates a new Kubernetes client
// It attempts to load configuration in this order:
// 1. In-cluster config
// 2. KUBECONFIG environment variable
// 3. ~/.kube/config
func NewClient(kubeconfigPath string, context string, namespace string) (*Client, error) {
	var config *rest.Config
	var err error

	// Try to load kubeconfig
	if kubeconfigPath == "" {
		// Check environment variable
		if envConfig := os.Getenv("KUBECONFIG"); envConfig != "" {
			kubeconfigPath = envConfig
		} else if home := homedir.HomeDir(); home != "" {
			// Use default location
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Load available contexts
	contexts, currentContext, err := loadContexts(kubeconfigPath)
	if err != nil {
		// Non-fatal, just use empty contexts
		contexts = []string{}
		currentContext = "unknown"
	}

	// Use provided context or current context
	if context != "" {
		currentContext = context
	}

	// Default namespace if not specified
	if namespace == "" {
		namespace = "default"
	}

	return &Client{
		clientset:      clientset,
		config:         config,
		namespace:      namespace,
		currentContext: currentContext,
		contexts:       contexts,
	}, nil
}

// loadContexts reads available contexts from kubeconfig
func loadContexts(kubeconfigPath string) ([]string, string, error) {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, "", err
	}

	contexts := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, config.CurrentContext, nil
}

// GetPods retrieves pods from the specified namespace
func (c *Client) GetPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods, nil
}

// GetAllPods retrieves pods from all namespaces
func (c *Client) GetAllPods(ctx context.Context) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all pods: %w", err)
	}

	return pods, nil
}

// GetNamespaces retrieves all namespaces
func (c *Client) GetNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return namespaces, nil
}

// GetPodLogs retrieves logs for a specific pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	opts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return string(logs), nil
}

// GetPod retrieves a specific pod
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return pod, nil
}

// SetNamespace changes the current namespace
func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
}

// GetNamespace returns the current namespace
func (c *Client) GetNamespace() string {
	return c.namespace
}

// GetCurrentContext returns the current context
func (c *Client) GetCurrentContext() string {
	return c.currentContext
}

// GetContexts returns all available contexts
func (c *Client) GetContexts() []string {
	return c.contexts
}

// TestConnection verifies the connection to the Kubernetes cluster
func (c *Client) TestConnection(parentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	_, err := c.clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Use ctx if we add more API calls later
	_ = ctx

	return nil
}
