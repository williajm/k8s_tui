package components

import (
	"strings"
	"testing"

	"github.com/williajm/k8s-tui/internal/models"
)

func TestNewDescribeViewer(t *testing.T) {
	dv := NewDescribeViewer()

	if dv == nil {
		t.Fatal("NewDescribeViewer() returned nil")
	}

	if dv.format != models.FormatDescribe {
		t.Errorf("format = %v, want %v", dv.format, models.FormatDescribe)
	}

	if dv.width != 80 {
		t.Errorf("width = %d, want 80", dv.width)
	}

	if dv.height != 20 {
		t.Errorf("height = %d, want 20", dv.height)
	}

	if dv.data != nil {
		t.Error("data should be nil initially")
	}
}

func TestDescribeViewer_SetSize(t *testing.T) {
	dv := NewDescribeViewer()

	width := 120
	height := 40

	dv.SetSize(width, height)

	if dv.width != width {
		t.Errorf("width = %d, want %d", dv.width, width)
	}

	if dv.height != height {
		t.Errorf("height = %d, want %d", dv.height, height)
	}

	// Viewport should be adjusted (accounting for border, header, footer)
	expectedVpWidth := width - 4
	expectedVpHeight := height - 6

	if dv.viewport.Width != expectedVpWidth {
		t.Errorf("viewport.Width = %d, want %d", dv.viewport.Width, expectedVpWidth)
	}

	if dv.viewport.Height != expectedVpHeight {
		t.Errorf("viewport.Height = %d, want %d", dv.viewport.Height, expectedVpHeight)
	}
}

func TestDescribeViewer_SetData(t *testing.T) {
	dv := NewDescribeViewer()

	data := models.NewDescribeData("Pod", "nginx-pod", "default")
	metadata := data.AddSection("Metadata")
	metadata.AddField("Name", "nginx-pod", 0)
	metadata.AddField("Namespace", "default", 0)

	dv.SetData(data)

	if dv.data != data {
		t.Error("SetData() did not set data correctly")
	}

	if dv.data.Name != "nginx-pod" {
		t.Errorf("data.Name = %v, want nginx-pod", dv.data.Name)
	}
}

func TestDescribeViewer_SetYAML(t *testing.T) {
	dv := NewDescribeViewer()

	yamlContent := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx"
	dv.SetYAML(yamlContent)

	if dv.yamlCache != yamlContent {
		t.Errorf("yamlCache = %q, want %q", dv.yamlCache, yamlContent)
	}
}

func TestDescribeViewer_SetJSON(t *testing.T) {
	dv := NewDescribeViewer()

	jsonContent := `{"apiVersion": "v1", "kind": "Pod"}`
	dv.SetJSON(jsonContent)

	if dv.jsonCache != jsonContent {
		t.Errorf("jsonCache = %q, want %q", dv.jsonCache, jsonContent)
	}
}

func TestDescribeViewer_SetFormat(t *testing.T) {
	dv := NewDescribeViewer()

	tests := []struct {
		name   string
		format models.DescribeFormat
	}{
		{"describe format", models.FormatDescribe},
		{"yaml format", models.FormatYAML},
		{"json format", models.FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dv.SetFormat(tt.format)

			if dv.format != tt.format {
				t.Errorf("format = %v, want %v", dv.format, tt.format)
			}
		})
	}
}

func TestDescribeViewer_CycleFormat(t *testing.T) {
	dv := NewDescribeViewer()

	// Start at FormatDescribe
	if dv.format != models.FormatDescribe {
		t.Errorf("initial format = %v, want %v", dv.format, models.FormatDescribe)
	}

	// Cycle to YAML
	dv.CycleFormat()
	if dv.format != models.FormatYAML {
		t.Errorf("format after 1st cycle = %v, want %v", dv.format, models.FormatYAML)
	}

	// Cycle to JSON
	dv.CycleFormat()
	if dv.format != models.FormatJSON {
		t.Errorf("format after 2nd cycle = %v, want %v", dv.format, models.FormatJSON)
	}

	// Cycle back to Describe
	dv.CycleFormat()
	if dv.format != models.FormatDescribe {
		t.Errorf("format after 3rd cycle = %v, want %v", dv.format, models.FormatDescribe)
	}
}

func TestDescribeViewer_View(t *testing.T) {
	dv := NewDescribeViewer()
	dv.SetSize(100, 30)

	t.Run("view without data", func(t *testing.T) {
		view := dv.View()
		if view == "" {
			t.Error("View() returned empty string")
		}
	})

	t.Run("view with data", func(t *testing.T) {
		data := models.NewDescribeData("Pod", "test-pod", "default")
		metadata := data.AddSection("Metadata")
		metadata.AddField("Name", "test-pod", 0)

		dv.SetData(data)
		view := dv.View()

		if view == "" {
			t.Error("View() returned empty string")
		}

		// View should contain the pod name
		if !strings.Contains(view, "test-pod") {
			t.Error("View() should contain pod name")
		}
	})
}

func TestDescribeViewer_ViewFormats(t *testing.T) {
	dv := NewDescribeViewer()
	dv.SetSize(100, 30)

	// Set up data
	data := models.NewDescribeData("Pod", "nginx", "default")
	metadata := data.AddSection("Metadata")
	metadata.AddField("Name", "nginx", 0)
	metadata.AddField("Namespace", "default", 0)

	dv.SetData(data)

	t.Run("describe format view", func(t *testing.T) {
		dv.SetFormat(models.FormatDescribe)
		view := dv.View()

		if view == "" {
			t.Error("View() returned empty for Describe format")
		}

		if !strings.Contains(view, "nginx") {
			t.Error("Describe view should contain resource name")
		}
	})

	t.Run("yaml format view", func(t *testing.T) {
		dv.SetFormat(models.FormatYAML)
		yamlContent := "apiVersion: v1\nkind: Pod"
		dv.SetYAML(yamlContent)

		view := dv.View()

		if view == "" {
			t.Error("View() returned empty for YAML format")
		}
	})

	t.Run("json format view", func(t *testing.T) {
		dv.SetFormat(models.FormatJSON)
		jsonContent := `{"kind": "Pod"}`
		dv.SetJSON(jsonContent)

		view := dv.View()

		if view == "" {
			t.Error("View() returned empty for JSON format")
		}
	})
}

func TestDescribeViewer_RenderDescribeFormat(t *testing.T) {
	dv := NewDescribeViewer()

	t.Run("with nil data", func(t *testing.T) {
		dv.data = nil
		content := dv.renderDescribeFormat()

		if content != "No data available" {
			t.Errorf("renderDescribeFormat() with nil data = %q, want 'No data available'", content)
		}
	})

	t.Run("with valid data", func(t *testing.T) {
		data := models.NewDescribeData("Service", "api-service", "production")

		metadata := data.AddSection("Metadata")
		metadata.AddField("Name", "api-service", 0)
		metadata.AddField("Namespace", "production", 0)

		spec := data.AddSection("Spec")
		spec.AddField("Type", "ClusterIP", 0)
		spec.AddField("ClusterIP", "10.96.0.1", 0)

		dv.data = data
		content := dv.renderDescribeFormat()

		if !strings.Contains(content, "api-service") {
			t.Error("renderDescribeFormat() should contain service name")
		}

		if !strings.Contains(content, "production") {
			t.Error("renderDescribeFormat() should contain namespace")
		}

		if !strings.Contains(content, "Metadata") {
			t.Error("renderDescribeFormat() should contain section title")
		}
	})

	t.Run("with indented fields", func(t *testing.T) {
		data := models.NewDescribeData("Pod", "nginx", "default")
		containers := data.AddSection("Containers")
		containers.AddField("nginx", "", 0)
		containers.AddField("Image", "nginx:1.21", 1)
		containers.AddField("Port", "80/TCP", 2)

		dv.data = data
		content := dv.renderDescribeFormat()

		if !strings.Contains(content, "nginx") {
			t.Error("renderDescribeFormat() should contain container info")
		}
	})
}

func TestDescribeViewer_RenderYAMLFormat(t *testing.T) {
	dv := NewDescribeViewer()

	t.Run("without cached YAML", func(t *testing.T) {
		content := dv.renderYAMLFormat()

		if content != "Loading YAML..." {
			t.Errorf("renderYAMLFormat() without cache = %q, want 'Loading YAML...'", content)
		}
	})

	t.Run("with cached YAML", func(t *testing.T) {
		yamlContent := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx"
		dv.yamlCache = yamlContent

		content := dv.renderYAMLFormat()

		if content != yamlContent {
			t.Errorf("renderYAMLFormat() = %q, want %q", content, yamlContent)
		}
	})
}

func TestDescribeViewer_RenderJSONFormat(t *testing.T) {
	dv := NewDescribeViewer()

	t.Run("without cached JSON", func(t *testing.T) {
		content := dv.renderJSONFormat()

		if content != "Loading JSON..." {
			t.Errorf("renderJSONFormat() without cache = %q, want 'Loading JSON...'", content)
		}
	})

	t.Run("with cached JSON", func(t *testing.T) {
		jsonContent := `{"apiVersion": "v1", "kind": "Pod"}`
		dv.jsonCache = jsonContent

		content := dv.renderJSONFormat()

		if content != jsonContent {
			t.Errorf("renderJSONFormat() = %q, want %q", content, jsonContent)
		}
	})
}

func TestDescribeViewer_GetViewport(t *testing.T) {
	dv := NewDescribeViewer()

	vp := dv.GetViewport()

	if vp == nil {
		t.Error("GetViewport() returned nil")
	}

	if vp != &dv.viewport {
		t.Error("GetViewport() should return pointer to internal viewport")
	}
}

func TestDescribeViewer_RenderHeader(t *testing.T) {
	dv := NewDescribeViewer()
	dv.SetSize(100, 30)

	t.Run("without data", func(t *testing.T) {
		header := dv.renderHeader()

		if !strings.Contains(header, "Resource Description") {
			t.Error("renderHeader() without data should contain default title")
		}
	})

	t.Run("with data", func(t *testing.T) {
		data := models.NewDescribeData("Deployment", "web-app", "production")
		dv.SetData(data)
		dv.SetFormat(models.FormatYAML)

		header := dv.renderHeader()

		if !strings.Contains(header, "Deployment") {
			t.Error("renderHeader() should contain resource kind")
		}

		if !strings.Contains(header, "web-app") {
			t.Error("renderHeader() should contain resource name")
		}

		if !strings.Contains(header, "production") {
			t.Error("renderHeader() should contain namespace")
		}

		if !strings.Contains(header, "YAML") {
			t.Error("renderHeader() should contain format name")
		}
	})
}

func TestDescribeViewer_RenderFooter(t *testing.T) {
	dv := NewDescribeViewer()
	dv.SetSize(100, 30)

	footer := dv.renderFooter()

	// Footer should contain keyboard hints
	if !strings.Contains(footer, "YAML") {
		t.Error("renderFooter() should contain YAML hint")
	}

	if !strings.Contains(footer, "JSON") {
		t.Error("renderFooter() should contain JSON hint")
	}

	if !strings.Contains(footer, "Describe") {
		t.Error("renderFooter() should contain Describe hint")
	}

	if !strings.Contains(footer, "Scroll") && !strings.Contains(footer, "scroll") {
		t.Error("renderFooter() should contain scroll hint")
	}
}

func TestDescribeViewer_CompleteWorkflow(t *testing.T) {
	// Test a complete workflow
	dv := NewDescribeViewer()
	dv.SetSize(120, 40)

	// Create describe data
	data := models.NewDescribeData("StatefulSet", "database", "production")

	metadata := data.AddSection("Metadata")
	metadata.AddField("Name", "database", 0)
	metadata.AddField("Namespace", "production", 0)
	metadata.AddField("Labels", "app=db,tier=data", 0)

	strategy := data.AddSection("Update Strategy")
	strategy.AddField("Type", "RollingUpdate", 0)

	replicas := data.AddSection("Replicas")
	replicas.AddField("Desired", "5", 0)
	replicas.AddField("Current", "5", 0)
	replicas.AddField("Ready", "5", 0)

	// Set data
	dv.SetData(data)

	// Set YAML and JSON
	dv.SetYAML("apiVersion: apps/v1\nkind: StatefulSet")
	dv.SetJSON(`{"kind": "StatefulSet", "apiVersion": "apps/v1"}`)

	// Cycle through formats and verify views
	formats := []models.DescribeFormat{
		models.FormatDescribe,
		models.FormatYAML,
		models.FormatJSON,
	}

	for _, format := range formats {
		dv.SetFormat(format)
		view := dv.View()

		if view == "" {
			t.Errorf("View() returned empty for format %v", format)
		}
	}
}

func TestDescribeViewer_RenderDescribeField(t *testing.T) {
	dv := NewDescribeViewer()

	tests := []struct {
		name  string
		field models.DescribeField
	}{
		{
			name: "zero indent",
			field: models.DescribeField{
				Key:    "Name",
				Value:  "nginx",
				Indent: 0,
			},
		},
		{
			name: "one indent",
			field: models.DescribeField{
				Key:    "Image",
				Value:  "nginx:1.21",
				Indent: 1,
			},
		},
		{
			name: "two indent",
			field: models.DescribeField{
				Key:    "Port",
				Value:  "80/TCP",
				Indent: 2,
			},
		},
		{
			name: "value only field",
			field: models.DescribeField{
				Key:    "",
				Value:  "list item",
				Indent: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dv.renderDescribeField(tt.field)

			if result == "" {
				t.Error("renderDescribeField() returned empty string")
			}

			// For value-only fields, should not contain colon
			if tt.field.Key == "" {
				if strings.Contains(result, ":") {
					t.Error("Value-only field should not contain colon")
				}
			}
		})
	}
}
