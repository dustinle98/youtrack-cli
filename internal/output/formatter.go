package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Format specifies the output format.
type Format int

const (
	FormatTable Format = iota
	FormatJSON
	FormatQuiet
)

// Formatter handles output rendering.
type Formatter struct {
	format Format
	fields []string
}

// New creates a new Formatter based on flags.
func New(jsonFlag, quietFlag bool, fieldsFlag string) *Formatter {
	f := &Formatter{format: FormatTable}
	if jsonFlag {
		f.format = FormatJSON
	}
	if quietFlag {
		f.format = FormatQuiet
	}
	if fieldsFlag != "" {
		for _, field := range strings.Split(fieldsFlag, ",") {
			f.fields = append(f.fields, strings.TrimSpace(field))
		}
	}
	return f
}

// PrintJSON outputs an object as JSON.
func (f *Formatter) PrintJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// If fields filter is set, filter the JSON
	if len(f.fields) > 0 {
		filtered, fErr := filterJSON(data, f.fields)
		if fErr == nil {
			data = filtered
		}
	}

	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

// PrintTable outputs data as a simple aligned table.
func (f *Formatter) PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%-*s", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i, w := range widths {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("─", w))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i := 0; i < len(headers); i++ {
			if i > 0 {
				fmt.Print("  ")
			}
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Printf("%-*s", widths[i], cell)
		}
		fmt.Println()
	}
}

// PrintQuiet outputs minimal data (one item per line).
func (f *Formatter) PrintQuiet(items []string) {
	for _, item := range items {
		fmt.Println(item)
	}
}

// IsJSON returns true if JSON output is requested.
func (f *Formatter) IsJSON() bool {
	return f.format == FormatJSON
}

// IsQuiet returns true if quiet output is requested.
func (f *Formatter) IsQuiet() bool {
	return f.format == FormatQuiet
}

// filterJSON filters a JSON object/array to only include specified fields.
func filterJSON(data []byte, fields []string) ([]byte, error) {
	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[strings.ToLower(f)] = true
	}

	// Try as object
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		filtered := filterMap(obj, fieldSet)
		return json.MarshalIndent(filtered, "", "  ")
	}

	// Try as array
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			result[i] = filterMap(item, fieldSet)
		}
		return json.MarshalIndent(result, "", "  ")
	}

	return data, nil
}

func filterMap(m map[string]interface{}, fields map[string]bool) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if fields[strings.ToLower(k)] {
			result[k] = v
		}
	}
	return result
}
