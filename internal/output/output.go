package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Format represents the output format
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Writer handles output formatting
type Writer struct {
	format Format
	out    io.Writer
}

// NewWriter creates a new output writer
func NewWriter(format Format) *Writer {
	return &Writer{
		format: format,
		out:    os.Stdout,
	}
}

// WriteTable writes data as a table or JSON depending on format
func (w *Writer) WriteTable(headers []string, rows [][]string) error {
	if w.format == FormatJSON {
		return w.writeTableAsJSON(headers, rows)
	}
	return w.writeASCIITable(headers, rows)
}

// WriteJSON writes data as JSON
func (w *Writer) WriteJSON(data interface{}) error {
	enc := json.NewEncoder(w.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// WriteSuccess writes a success message
func (w *Writer) WriteSuccess(msg string) {
	if w.format == FormatJSON {
		w.WriteJSON(map[string]string{"status": "success", "message": msg})
	} else {
		fmt.Fprintln(w.out, msg)
	}
}

// WriteError writes an error message to stderr
func (w *Writer) WriteError(err error) {
	if w.format == FormatJSON {
		json.NewEncoder(os.Stderr).Encode(map[string]string{"status": "error", "message": err.Error()})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func (w *Writer) writeASCIITable(headers []string, rows [][]string) error {
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

	// Print headers
	var headerParts []string
	for i, h := range headers {
		headerParts = append(headerParts, fmt.Sprintf("%-*s", widths[i], h))
	}
	fmt.Fprintln(w.out, strings.Join(headerParts, "  "))

	// Print rows
	for _, row := range rows {
		var rowParts []string
		for i, cell := range row {
			if i < len(widths) {
				rowParts = append(rowParts, fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		fmt.Fprintln(w.out, strings.Join(rowParts, "  "))
	}

	return nil
}

func (w *Writer) writeTableAsJSON(headers []string, rows [][]string) error {
	var result []map[string]string
	for _, row := range rows {
		item := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				item[header] = row[i]
			}
		}
		result = append(result, item)
	}
	return w.WriteJSON(result)
}

// FormatBool formats a boolean for display
func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// FormatInt formats an int for display
func FormatInt(n int) string {
	return strconv.Itoa(n)
}

// FormatTTL formats a TTL value for display
func FormatTTL(ttl int) string {
	if ttl == 1 {
		return "Auto"
	}
	return strconv.Itoa(ttl)
}
