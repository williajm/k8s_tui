package models

// DescribeFormat represents the format for displaying resource details
type DescribeFormat int

const (
	FormatDescribe DescribeFormat = iota
	FormatYAML
	FormatJSON
)

// DescribeData holds formatted resource description information
type DescribeData struct {
	Kind      string
	Name      string
	Namespace string
	Sections  []DescribeSection
	RawYAML   string
	RawJSON   string
}

// DescribeSection represents a section in the describe output
type DescribeSection struct {
	Title  string
	Fields []DescribeField
}

// DescribeField represents a single field in a describe section
type DescribeField struct {
	Key    string
	Value  string
	Indent int
}

// NewDescribeData creates a new DescribeData structure
func NewDescribeData(kind, name, namespace string) *DescribeData {
	return &DescribeData{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		Sections:  make([]DescribeSection, 0),
	}
}

// AddSection adds a new section to the describe data
func (d *DescribeData) AddSection(title string) *DescribeSection {
	section := DescribeSection{
		Title:  title,
		Fields: make([]DescribeField, 0),
	}
	d.Sections = append(d.Sections, section)
	return &d.Sections[len(d.Sections)-1]
}

// AddField adds a field to a section
func (s *DescribeSection) AddField(key, value string, indent int) {
	s.Fields = append(s.Fields, DescribeField{
		Key:    key,
		Value:  value,
		Indent: indent,
	})
}

// String returns the format as a string
func (f DescribeFormat) String() string {
	switch f {
	case FormatDescribe:
		return "Describe"
	case FormatYAML:
		return "YAML"
	case FormatJSON:
		return "JSON"
	default:
		return "Unknown"
	}
}
