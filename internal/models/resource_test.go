package models

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPodInfo(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		wantName string
		wantNS   string
	}{
		{
			name: "basic pod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "10.0.0.1",
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "nginx",
							Ready:        true,
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			wantName: "test-pod",
			wantNS:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewPodInfo(tt.pod)

			if info.Name != tt.wantName {
				t.Errorf("NewPodInfo().Name = %v, want %v", info.Name, tt.wantName)
			}

			if info.Namespace != tt.wantNS {
				t.Errorf("NewPodInfo().Namespace = %v, want %v", info.Namespace, tt.wantNS)
			}

			if info.Status != "Running" {
				t.Errorf("NewPodInfo().Status = %v, want Running", info.Status)
			}

			if info.Ready != "1/1" {
				t.Errorf("NewPodInfo().Ready = %v, want 1/1", info.Ready)
			}

			if info.IP != "10.0.0.1" {
				t.Errorf("NewPodInfo().IP = %v, want 10.0.0.1", info.IP)
			}
		})
	}
}

func TestPodInfo_GetStatusSymbol(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		ready      string
		containers int
		want       string
	}{
		{
			name:       "running and ready",
			status:     "Running",
			ready:      "1/1",
			containers: 1,
			want:       "●",
		},
		{
			name:       "succeeded",
			status:     "Succeeded",
			ready:      "0/1",
			containers: 1,
			want:       "✔",
		},
		{
			name:       "failed",
			status:     "Failed",
			ready:      "0/1",
			containers: 1,
			want:       "✖",
		},
		{
			name:       "error",
			status:     "Error",
			ready:      "0/1",
			containers: 1,
			want:       "✖",
		},
		{
			name:       "pending",
			status:     "Pending",
			ready:      "0/1",
			containers: 1,
			want:       "◐",
		},
		{
			name:       "terminating",
			status:     "Terminating",
			ready:      "0/1",
			containers: 1,
			want:       "◌",
		},
		{
			name:       "not ready",
			status:     "NotReady",
			ready:      "1/2",
			containers: 2,
			want:       "◑",
		},
		{
			name:       "unknown",
			status:     "Unknown",
			ready:      "0/1",
			containers: 1,
			want:       "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PodInfo{
				Status:     tt.status,
				Ready:      tt.ready,
				Containers: make([]ContainerInfo, tt.containers),
			}

			got := p.GetStatusSymbol()
			if got != tt.want {
				t.Errorf("GetStatusSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodInfo_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		ready      string
		containers int
		want       bool
	}{
		{
			name:       "healthy pod",
			status:     "Running",
			ready:      "1/1",
			containers: 1,
			want:       true,
		},
		{
			name:       "not ready",
			status:     "Running",
			ready:      "0/1",
			containers: 1,
			want:       false,
		},
		{
			name:       "pending",
			status:     "Pending",
			ready:      "0/1",
			containers: 1,
			want:       false,
		},
		{
			name:       "multi-container healthy",
			status:     "Running",
			ready:      "2/2",
			containers: 2,
			want:       true,
		},
		{
			name:       "multi-container partial",
			status:     "Running",
			ready:      "1/2",
			containers: 2,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PodInfo{
				Status:     tt.status,
				Ready:      tt.ready,
				Containers: make([]ContainerInfo, tt.containers),
			}

			got := p.IsHealthy()
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name      string
		timestamp metav1.Time
		want      string
	}{
		{
			name: "seconds",
			timestamp: metav1.Time{
				Time: time.Now().Add(-30 * time.Second),
			},
			want: "30s",
		},
		{
			name: "minutes",
			timestamp: metav1.Time{
				Time: time.Now().Add(-5 * time.Minute),
			},
			want: "5m",
		},
		{
			name: "hours",
			timestamp: metav1.Time{
				Time: time.Now().Add(-3 * time.Hour),
			},
			want: "3h",
		},
		{
			name: "days",
			timestamp: metav1.Time{
				Time: time.Now().Add(-48 * time.Hour),
			},
			want: "2d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.timestamp)
			if got != tt.want {
				t.Errorf("formatAge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewServiceInfo(t *testing.T) {
	tests := []struct {
		name    string
		service *corev1.Service
		want    ServiceInfo
	}{
		{
			name: "ClusterIP service",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "10.96.0.1",
					Ports: []corev1.ServicePort{
						{
							Port:     80,
							Protocol: corev1.ProtocolTCP,
						},
					},
					Selector: map[string]string{
						"app": "nginx",
					},
				},
			},
			want: ServiceInfo{
				Name:       "test-service",
				Namespace:  "default",
				Type:       "ClusterIP",
				ClusterIP:  "10.96.0.1",
				ExternalIP: "<none>",
				Ports:      "80/TCP",
				Age:        "10m",
			},
		},
		{
			name: "NodePort service",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodeport-service",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeNodePort,
					ClusterIP: "10.96.0.2",
					Ports: []corev1.ServicePort{
						{
							Port:     80,
							NodePort: 30080,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
			want: ServiceInfo{
				Name:       "nodeport-service",
				Namespace:  "default",
				Type:       "NodePort",
				ClusterIP:  "10.96.0.2",
				ExternalIP: "<none>",
				Ports:      "80:30080/TCP",
				Age:        "1h",
			},
		},
		{
			name: "LoadBalancer service with external IP",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb-service",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-2 * time.Hour),
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeLoadBalancer,
					ClusterIP: "10.96.0.3",
					Ports: []corev1.ServicePort{
						{
							Port:     443,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "203.0.113.1"},
						},
					},
				},
			},
			want: ServiceInfo{
				Name:       "lb-service",
				Namespace:  "default",
				Type:       "LoadBalancer",
				ClusterIP:  "10.96.0.3",
				ExternalIP: "203.0.113.1",
				Ports:      "443/TCP",
				Age:        "2h",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServiceInfo(tt.service)

			if got.Name != tt.want.Name {
				t.Errorf("NewServiceInfo().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("NewServiceInfo().Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if got.Type != tt.want.Type {
				t.Errorf("NewServiceInfo().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.ClusterIP != tt.want.ClusterIP {
				t.Errorf("NewServiceInfo().ClusterIP = %v, want %v", got.ClusterIP, tt.want.ClusterIP)
			}
			if got.ExternalIP != tt.want.ExternalIP {
				t.Errorf("NewServiceInfo().ExternalIP = %v, want %v", got.ExternalIP, tt.want.ExternalIP)
			}
			if got.Ports != tt.want.Ports {
				t.Errorf("NewServiceInfo().Ports = %v, want %v", got.Ports, tt.want.Ports)
			}
			if got.Age != tt.want.Age {
				t.Errorf("NewServiceInfo().Age = %v, want %v", got.Age, tt.want.Age)
			}
		})
	}
}

func TestServiceInfo_GetStatusSymbol(t *testing.T) {
	tests := []struct {
		name    string
		service ServiceInfo
		want    string
	}{
		{
			name: "ClusterIP service",
			service: ServiceInfo{
				Type:       "ClusterIP",
				ExternalIP: "<none>",
			},
			want: "●",
		},
		{
			name: "LoadBalancer with external IP",
			service: ServiceInfo{
				Type:       "LoadBalancer",
				ExternalIP: "203.0.113.1",
			},
			want: "●",
		},
		{
			name: "LoadBalancer pending external IP",
			service: ServiceInfo{
				Type:       "LoadBalancer",
				ExternalIP: "<none>",
			},
			want: "◐",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.service.GetStatusSymbol()
			if got != tt.want {
				t.Errorf("GetStatusSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDeploymentInfo(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		want       DeploymentInfo
	}{
		{
			name: "healthy deployment",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-deployment",
					Namespace: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
					},
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:     3,
					UpdatedReplicas:   3,
					AvailableReplicas: 3,
				},
			},
			want: DeploymentInfo{
				Name:      "nginx-deployment",
				Namespace: "default",
				Ready:     "3/3",
				UpToDate:  3,
				Available: 3,
				Age:       "30m",
				Replicas:  3,
				Strategy:  "RollingUpdate",
			},
		},
		{
			name: "partially ready deployment",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app-deployment",
					Namespace: "production",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:     1,
					UpdatedReplicas:   2,
					AvailableReplicas: 1,
				},
			},
			want: DeploymentInfo{
				Name:      "app-deployment",
				Namespace: "production",
				Ready:     "1/3",
				UpToDate:  2,
				Available: 1,
				Age:       "1h",
				Replicas:  3,
				Strategy:  "Recreate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDeploymentInfo(tt.deployment)

			if got.Name != tt.want.Name {
				t.Errorf("NewDeploymentInfo().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("NewDeploymentInfo().Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if got.Ready != tt.want.Ready {
				t.Errorf("NewDeploymentInfo().Ready = %v, want %v", got.Ready, tt.want.Ready)
			}
			if got.UpToDate != tt.want.UpToDate {
				t.Errorf("NewDeploymentInfo().UpToDate = %v, want %v", got.UpToDate, tt.want.UpToDate)
			}
			if got.Available != tt.want.Available {
				t.Errorf("NewDeploymentInfo().Available = %v, want %v", got.Available, tt.want.Available)
			}
			if got.Replicas != tt.want.Replicas {
				t.Errorf("NewDeploymentInfo().Replicas = %v, want %v", got.Replicas, tt.want.Replicas)
			}
			if got.Strategy != tt.want.Strategy {
				t.Errorf("NewDeploymentInfo().Strategy = %v, want %v", got.Strategy, tt.want.Strategy)
			}
		})
	}
}

func TestDeploymentInfo_GetStatusSymbol(t *testing.T) {
	tests := []struct {
		name       string
		deployment DeploymentInfo
		want       string
	}{
		{
			name: "healthy deployment",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 3,
				UpToDate:  3,
			},
			want: "●",
		},
		{
			name: "partially ready",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 1,
				UpToDate:  3,
			},
			want: "◑",
		},
		{
			name: "not available",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 0,
				UpToDate:  0,
			},
			want: "✖",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deployment.GetStatusSymbol()
			if got != tt.want {
				t.Errorf("GetStatusSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentInfo_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		deployment DeploymentInfo
		want       bool
	}{
		{
			name: "healthy",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 3,
				UpToDate:  3,
			},
			want: true,
		},
		{
			name: "not up to date",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 3,
				UpToDate:  2,
			},
			want: false,
		},
		{
			name: "not available",
			deployment: DeploymentInfo{
				Replicas:  3,
				Available: 2,
				UpToDate:  3,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deployment.IsHealthy()
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStatefulSetInfo(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name        string
		statefulSet *appsv1.StatefulSet
		want        StatefulSetInfo
	}{
		{
			name: "healthy statefulset",
			statefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgres-sts",
					Namespace: "database",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-2 * time.Hour),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.RollingUpdateStatefulSetStrategyType,
					},
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 3,
				},
			},
			want: StatefulSetInfo{
				Name:      "postgres-sts",
				Namespace: "database",
				Ready:     "3/3",
				Age:       "2h",
				Replicas:  3,
				Strategy:  "RollingUpdate",
			},
		},
		{
			name: "partially ready statefulset",
			statefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "redis-sts",
					Namespace: "cache",
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-45 * time.Minute),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
						Type: appsv1.OnDeleteStatefulSetStrategyType,
					},
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 2,
				},
			},
			want: StatefulSetInfo{
				Name:      "redis-sts",
				Namespace: "cache",
				Ready:     "2/3",
				Age:       "45m",
				Replicas:  3,
				Strategy:  "OnDelete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStatefulSetInfo(tt.statefulSet)

			if got.Name != tt.want.Name {
				t.Errorf("NewStatefulSetInfo().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("NewStatefulSetInfo().Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if got.Ready != tt.want.Ready {
				t.Errorf("NewStatefulSetInfo().Ready = %v, want %v", got.Ready, tt.want.Ready)
			}
			if got.Replicas != tt.want.Replicas {
				t.Errorf("NewStatefulSetInfo().Replicas = %v, want %v", got.Replicas, tt.want.Replicas)
			}
			if got.Strategy != tt.want.Strategy {
				t.Errorf("NewStatefulSetInfo().Strategy = %v, want %v", got.Strategy, tt.want.Strategy)
			}
		})
	}
}

func TestStatefulSetInfo_GetStatusSymbol(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name        string
		statefulSet StatefulSetInfo
		want        string
	}{
		{
			name: "healthy",
			statefulSet: StatefulSetInfo{
				Replicas: 3,
				StatefulSet: &appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 3,
					},
				},
			},
			want: "●",
		},
		{
			name: "partially ready",
			statefulSet: StatefulSetInfo{
				Replicas: 3,
				StatefulSet: &appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 2,
					},
				},
			},
			want: "◑",
		},
		{
			name: "not ready",
			statefulSet: StatefulSetInfo{
				Replicas: replicas,
				StatefulSet: &appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				},
			},
			want: "✖",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.statefulSet.GetStatusSymbol()
			if got != tt.want {
				t.Errorf("GetStatusSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatefulSetInfo_IsHealthy(t *testing.T) {
	tests := []struct {
		name        string
		statefulSet StatefulSetInfo
		want        bool
	}{
		{
			name: "healthy",
			statefulSet: StatefulSetInfo{
				Replicas: 3,
				StatefulSet: &appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 3,
					},
				},
			},
			want: true,
		},
		{
			name: "not healthy",
			statefulSet: StatefulSetInfo{
				Replicas: 3,
				StatefulSet: &appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 2,
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.statefulSet.IsHealthy()
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
