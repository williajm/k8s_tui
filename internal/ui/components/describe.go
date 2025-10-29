package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/k8s-tui/internal/models"
	"github.com/williajm/k8s-tui/internal/ui/styles"
)

// DescribeViewer displays resource descriptions in multiple formats
type DescribeViewer struct {
	data      *models.DescribeData
	viewport  viewport.Model
	format    models.DescribeFormat
	width     int
	height    int
	yamlCache string
	jsonCache string
}

// NewDescribeViewer creates a new describe viewer component
func NewDescribeViewer() *DescribeViewer {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	return &DescribeViewer{
		viewport: vp,
		format:   models.FormatDescribe,
		width:    80,
		height:   20,
	}
}

// SetSize sets the dimensions of the describe viewer
func (d *DescribeViewer) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.viewport.Width = width - 4
	// Height calculation: total - footer (2 lines outside box) - border (2) - header with border (2) = 6
	d.viewport.Height = height - 6
}

// SetData sets the describe data and updates the viewport
func (d *DescribeViewer) SetData(data *models.DescribeData) {
	d.data = data
	d.updateViewportContent()
}

// SetYAML sets the YAML content for the resource
func (d *DescribeViewer) SetYAML(yaml string) {
	d.yamlCache = yaml
	if d.format == models.FormatYAML {
		d.updateViewportContent()
	}
}

// SetJSON sets the JSON content for the resource
func (d *DescribeViewer) SetJSON(json string) {
	d.jsonCache = json
	if d.format == models.FormatJSON {
		d.updateViewportContent()
	}
}

// SetFormat changes the display format
func (d *DescribeViewer) SetFormat(format models.DescribeFormat) {
	d.format = format
	d.updateViewportContent()
}

// CycleFormat cycles through available formats
func (d *DescribeViewer) CycleFormat() {
	switch d.format {
	case models.FormatDescribe:
		d.format = models.FormatYAML
	case models.FormatYAML:
		d.format = models.FormatJSON
	case models.FormatJSON:
		d.format = models.FormatDescribe
	}
	d.updateViewportContent()
}

// Update handles viewport updates
func (d *DescribeViewer) Update(msg tea.Msg) (*DescribeViewer, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// View renders the describe viewer
func (d *DescribeViewer) View() string {
	// Build header
	header := d.renderHeader()

	// Render viewport
	viewportContent := d.viewport.View()

	// Combine header and viewport in bordered container
	describeContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportContent,
	)

	borderedContent := styles.BorderStyle.
		Width(d.width).
		Height(d.height - 2). // Reserve 2 lines for footer outside border
		Render(describeContent)

	// Build footer (rendered outside the bordered container)
	footer := d.renderFooter()

	// Combine bordered content with footer below it
	return lipgloss.JoinVertical(
		lipgloss.Left,
		borderedContent,
		footer,
	)
}

// renderHeader renders the describe viewer header
func (d *DescribeViewer) renderHeader() string {
	var title string
	if d.data != nil {
		title = fmt.Sprintf("%s: %s/%s [%s]",
			d.data.Kind,
			d.data.Namespace,
			d.data.Name,
			d.format.String(),
		)
	} else {
		title = "Resource Description"
	}

	headerStyle := styles.TableHeaderStyle.Width(d.width - 4)
	return headerStyle.Render(title)
}

// renderFooter renders the describe viewer footer
func (d *DescribeViewer) renderFooter() string {
	// Format indicator
	formatLine := fmt.Sprintf("Format: %s", d.format.String())

	// Keyboard shortcuts with proper styling
	shortcuts := []string{
		styles.RenderKeyHelp("[↑↓]", "Scroll"),
		styles.RenderKeyHelp("[d]", "Describe"),
		styles.RenderKeyHelp("[y]", "YAML"),
		styles.RenderKeyHelp("[j]", "JSON"),
		styles.RenderKeyHelp("[q/Esc]", "Back"),
		styles.RenderKeyHelp("[Ctrl+C]", "Quit"),
	}
	shortcutsLine := strings.Join(shortcuts, "  ")

	// Combine format and shortcuts
	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.FooterStyle.Width(d.width).Render(formatLine),
		styles.FooterStyle.Width(d.width).Render(shortcutsLine),
	)

	return footer
}

// updateViewportContent updates the viewport based on current format
func (d *DescribeViewer) updateViewportContent() {
	var content string

	switch d.format {
	case models.FormatDescribe:
		content = d.renderDescribeFormat()
	case models.FormatYAML:
		content = d.renderYAMLFormat()
	case models.FormatJSON:
		content = d.renderJSONFormat()
	}

	d.viewport.SetContent(content)
	d.viewport.GotoTop()
}

// renderDescribeFormat renders the kubectl-style describe format
func (d *DescribeViewer) renderDescribeFormat() string {
	if d.data == nil {
		return "No data available"
	}

	var lines []string

	// Add resource header
	lines = append(lines,
		fmt.Sprintf("Name:      %s", d.data.Name),
		fmt.Sprintf("Namespace: %s", d.data.Namespace),
		"",
	)

	// Render each section
	for _, section := range d.data.Sections {
		// Section title
		sectionStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))
		lines = append(lines, sectionStyle.Render(section.Title+":"))

		// Section fields
		for _, field := range section.Fields {
			line := d.renderDescribeField(field)
			lines = append(lines, line)
		}

		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderDescribeField renders a single field with proper indentation
func (d *DescribeViewer) renderDescribeField(field models.DescribeField) string {
	indent := strings.Repeat("  ", field.Indent)

	if field.Key == "" {
		// Value-only field (used for list items)
		return indent + field.Value
	}

	// Key-value field
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	key := keyStyle.Render(field.Key + ":")
	value := valueStyle.Render(field.Value)

	// Pad key to align values
	keyWidth := 20 - (field.Indent * 2)
	if keyWidth < 10 {
		keyWidth = 10
	}

	return fmt.Sprintf("%s%-*s %s", indent, keyWidth, key, value)
}

// renderYAMLFormat renders the YAML format
func (d *DescribeViewer) renderYAMLFormat() string {
	if d.yamlCache == "" {
		return "Loading YAML..."
	}
	return d.yamlCache
}

// renderJSONFormat renders the JSON format
func (d *DescribeViewer) renderJSONFormat() string {
	if d.jsonCache == "" {
		return "Loading JSON..."
	}
	return d.jsonCache
}

// GetViewport returns the viewport for external updates
func (d *DescribeViewer) GetViewport() *viewport.Model {
	return &d.viewport
}
