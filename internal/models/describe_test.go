package models

import (
	"testing"
)

func TestNewDescribeData(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		resName   string
		namespace string
	}{
		{
			name:      "create pod describe data",
			kind:      "Pod",
			resName:   "nginx",
			namespace: "default",
		},
		{
			name:      "create service describe data",
			kind:      "Service",
			resName:   "api-service",
			namespace: "kube-system",
		},
		{
			name:      "create deployment describe data",
			kind:      "Deployment",
			resName:   "web-app",
			namespace: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewDescribeData(tt.kind, tt.resName, tt.namespace)

			if data == nil {
				t.Fatal("NewDescribeData() returned nil")
			}

			if data.Kind != tt.kind {
				t.Errorf("Kind = %v, want %v", data.Kind, tt.kind)
			}

			if data.Name != tt.resName {
				t.Errorf("Name = %v, want %v", data.Name, tt.resName)
			}

			if data.Namespace != tt.namespace {
				t.Errorf("Namespace = %v, want %v", data.Namespace, tt.namespace)
			}

			if data.Sections == nil {
				t.Error("Sections should be initialized, not nil")
			}

			if len(data.Sections) != 0 {
				t.Errorf("Sections length = %d, want 0", len(data.Sections))
			}

			if data.RawYAML != "" {
				t.Error("RawYAML should be empty initially")
			}

			if data.RawJSON != "" {
				t.Error("RawJSON should be empty initially")
			}
		})
	}
}

func TestDescribeData_AddSection(t *testing.T) {
	data := NewDescribeData("Pod", "test-pod", "default")

	t.Run("add single section", func(t *testing.T) {
		section := data.AddSection("Metadata")

		if section == nil {
			t.Fatal("AddSection() returned nil")
		}

		if len(data.Sections) != 1 {
			t.Errorf("Sections length = %d, want 1", len(data.Sections))
		}

		if data.Sections[0].Title != "Metadata" {
			t.Errorf("Section title = %v, want Metadata", data.Sections[0].Title)
		}

		if section.Title != "Metadata" {
			t.Errorf("Returned section title = %v, want Metadata", section.Title)
		}

		if section.Fields == nil {
			t.Error("Section fields should be initialized")
		}
	})

	t.Run("add multiple sections", func(t *testing.T) {
		data := NewDescribeData("Pod", "test-pod", "default")

		sec1 := data.AddSection("Metadata")
		sec2 := data.AddSection("Spec")
		sec3 := data.AddSection("Status")

		if len(data.Sections) != 3 {
			t.Errorf("Sections length = %d, want 3", len(data.Sections))
		}

		if sec1.Title != "Metadata" {
			t.Errorf("Section 1 title = %v, want Metadata", sec1.Title)
		}

		if sec2.Title != "Spec" {
			t.Errorf("Section 2 title = %v, want Spec", sec2.Title)
		}

		if sec3.Title != "Status" {
			t.Errorf("Section 3 title = %v, want Status", sec3.Title)
		}
	})

	t.Run("returned section modifies data.Sections", func(t *testing.T) {
		data := NewDescribeData("Pod", "test-pod", "default")
		section := data.AddSection("Test")

		// Modify through returned pointer
		section.AddField("Key1", "Value1", 0)

		// Verify it's reflected in data.Sections
		if len(data.Sections[0].Fields) != 1 {
			t.Errorf("Fields length = %d, want 1", len(data.Sections[0].Fields))
		}

		if data.Sections[0].Fields[0].Key != "Key1" {
			t.Error("Field not properly added through returned section pointer")
		}
	})
}

func TestDescribeSection_AddField(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  string
		indent int
	}{
		{
			name:   "basic field",
			key:    "Name",
			value:  "nginx",
			indent: 0,
		},
		{
			name:   "indented field",
			key:    "Container",
			value:  "app",
			indent: 1,
		},
		{
			name:   "deeply indented field",
			key:    "Port",
			value:  "8080",
			indent: 2,
		},
		{
			name:   "empty key",
			key:    "",
			value:  "list item",
			indent: 1,
		},
		{
			name:   "empty value",
			key:    "EmptyValue",
			value:  "",
			indent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := &DescribeSection{
				Title:  "Test Section",
				Fields: make([]DescribeField, 0),
			}

			section.AddField(tt.key, tt.value, tt.indent)

			if len(section.Fields) != 1 {
				t.Fatalf("Fields length = %d, want 1", len(section.Fields))
			}

			field := section.Fields[0]

			if field.Key != tt.key {
				t.Errorf("Field.Key = %v, want %v", field.Key, tt.key)
			}

			if field.Value != tt.value {
				t.Errorf("Field.Value = %v, want %v", field.Value, tt.value)
			}

			if field.Indent != tt.indent {
				t.Errorf("Field.Indent = %v, want %v", field.Indent, tt.indent)
			}
		})
	}

	t.Run("add multiple fields", func(t *testing.T) {
		section := &DescribeSection{
			Title:  "Metadata",
			Fields: make([]DescribeField, 0),
		}

		section.AddField("Name", "nginx", 0)
		section.AddField("Namespace", "default", 0)
		section.AddField("Labels", "app=nginx", 0)

		if len(section.Fields) != 3 {
			t.Errorf("Fields length = %d, want 3", len(section.Fields))
		}

		expectedKeys := []string{"Name", "Namespace", "Labels"}
		for i, expected := range expectedKeys {
			if section.Fields[i].Key != expected {
				t.Errorf("Field[%d].Key = %v, want %v", i, section.Fields[i].Key, expected)
			}
		}
	})
}

func TestDescribeFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format DescribeFormat
		want   string
	}{
		{
			name:   "describe format",
			format: FormatDescribe,
			want:   "Describe",
		},
		{
			name:   "yaml format",
			format: FormatYAML,
			want:   "YAML",
		},
		{
			name:   "json format",
			format: FormatJSON,
			want:   "JSON",
		},
		{
			name:   "unknown format",
			format: DescribeFormat(999),
			want:   "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.want {
				t.Errorf("DescribeFormat(%d).String() = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestDescribeData_CompleteWorkflow(t *testing.T) {
	// Test a complete workflow of building describe data
	data := NewDescribeData("Pod", "nginx-pod", "production")

	// Add metadata section
	metadata := data.AddSection("Metadata")
	metadata.AddField("Name", "nginx-pod", 0)
	metadata.AddField("Namespace", "production", 0)
	metadata.AddField("Labels", "app=nginx,tier=frontend", 0)

	// Add spec section
	spec := data.AddSection("Spec")
	spec.AddField("Containers", "", 0)
	spec.AddField("nginx", "", 1)
	spec.AddField("Image", "nginx:1.21", 2)
	spec.AddField("Port", "80/TCP", 2)

	// Add status section
	status := data.AddSection("Status")
	status.AddField("Phase", "Running", 0)
	status.AddField("IP", "10.0.0.5", 0)

	// Set raw formats
	data.RawYAML = "apiVersion: v1\nkind: Pod"
	data.RawJSON = "{\"apiVersion\": \"v1\", \"kind\": \"Pod\"}"

	// Verify structure
	if len(data.Sections) != 3 {
		t.Errorf("Sections count = %d, want 3", len(data.Sections))
	}

	if data.Sections[0].Title != "Metadata" {
		t.Errorf("Section 0 title = %v, want Metadata", data.Sections[0].Title)
	}

	if len(data.Sections[0].Fields) != 3 {
		t.Errorf("Metadata fields count = %d, want 3", len(data.Sections[0].Fields))
	}

	if data.Sections[1].Title != "Spec" {
		t.Errorf("Section 1 title = %v, want Spec", data.Sections[1].Title)
	}

	if len(data.Sections[1].Fields) != 4 {
		t.Errorf("Spec fields count = %d, want 4", len(data.Sections[1].Fields))
	}

	if data.Sections[2].Title != "Status" {
		t.Errorf("Section 2 title = %v, want Status", data.Sections[2].Title)
	}

	if data.RawYAML == "" {
		t.Error("RawYAML should be set")
	}

	if data.RawJSON == "" {
		t.Error("RawJSON should be set")
	}
}

func TestDescribeField_StructureValidation(t *testing.T) {
	// Ensure DescribeField structure is as expected
	field := DescribeField{
		Key:    "TestKey",
		Value:  "TestValue",
		Indent: 3,
	}

	if field.Key != "TestKey" {
		t.Errorf("Key = %v, want TestKey", field.Key)
	}

	if field.Value != "TestValue" {
		t.Errorf("Value = %v, want TestValue", field.Value)
	}

	if field.Indent != 3 {
		t.Errorf("Indent = %v, want 3", field.Indent)
	}
}

func TestDescribeSection_StructureValidation(t *testing.T) {
	// Ensure DescribeSection structure is as expected
	section := DescribeSection{
		Title: "Test Section",
		Fields: []DescribeField{
			{Key: "Key1", Value: "Value1", Indent: 0},
			{Key: "Key2", Value: "Value2", Indent: 1},
		},
	}

	if section.Title != "Test Section" {
		t.Errorf("Title = %v, want Test Section", section.Title)
	}

	if len(section.Fields) != 2 {
		t.Errorf("Fields length = %d, want 2", len(section.Fields))
	}
}

func TestDescribeFormat_AllFormats(t *testing.T) {
	// Test all format constants exist and are unique
	formats := []struct {
		format DescribeFormat
		name   string
	}{
		{FormatDescribe, "Describe"},
		{FormatYAML, "YAML"},
		{FormatJSON, "JSON"},
	}

	seen := make(map[DescribeFormat]bool)
	for _, tc := range formats {
		if seen[tc.format] {
			t.Errorf("Duplicate format value: %d", tc.format)
		}
		seen[tc.format] = true

		if tc.format.String() != tc.name {
			t.Errorf("Format %d: String() = %v, want %v", tc.format, tc.format.String(), tc.name)
		}
	}
}
