package console

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// OutputFormat defines the output format type
type OutputFormat string

const (
	// TableFormat outputs as a table
	TableFormat OutputFormat = "table"
	// JSONFormat outputs as JSON
	JSONFormat OutputFormat = "json"
	// YAMLFormat outputs as YAML
	YAMLFormat OutputFormat = "yaml"
)

// TableOptions configures the appearance of table output
type TableOptions struct {
	// MinWidth is the minimum cell width including any padding
	MinWidth int
	// TabWidth is the width of tab characters
	TabWidth int
	// Padding is the cell padding
	Padding int
	// PadChar is the padding character
	PadChar byte
	// Flags controls formatting behavior
	Flags uint
	// Header determines if headers should be displayed
	Header bool
	// MaxColumnWidth limits the width of columns (0 means no limit)
	MaxColumnWidth int
}

// DefaultTableOptions returns the default table formatting options
func DefaultTableOptions() TableOptions {
	return TableOptions{
		MinWidth:       0,
		TabWidth:       2,
		Padding:        1,
		PadChar:        ' ',
		Flags:          0,
		Header:         true,
		MaxColumnWidth: 50,
	}
}

// Formatter formats data for CLI output
type Formatter struct {
	OutFormat    OutputFormat
	Writer       io.Writer
	TableOptions TableOptions
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format string) *Formatter {
	outputFormat := TableFormat
	switch strings.ToLower(format) {
	case "json":
		outputFormat = JSONFormat
	case "yaml":
		outputFormat = YAMLFormat
	}

	return &Formatter{
		OutFormat:    outputFormat,
		Writer:       os.Stdout,
		TableOptions: DefaultTableOptions(),
	}
}

// WithTableOptions sets custom table options and returns the formatter
func (f *Formatter) WithTableOptions(options TableOptions) *Formatter {
	f.TableOptions = options
	return f
}

// WithWriter sets a custom writer and returns the formatter
func (f *Formatter) WithWriter(writer io.Writer) *Formatter {
	if writer != nil {
		f.Writer = writer
	}
	return f
}

// Format formats the data according to the formatter's format
func (f *Formatter) Format(data interface{}) error {
	// Handle nil data gracefully
	if data == nil {
		fmt.Fprintln(f.Writer, "No data available")
		return nil
	}

	// Ensure we have a valid writer
	if f.Writer == nil {
		f.Writer = os.Stdout
	}

	switch f.OutFormat {
	case JSONFormat:
		return f.formatJSON(data)
	case YAMLFormat:
		return f.formatYAML(data)
	default:
		return f.formatTable(data)
	}
}

// formatJSON formats data as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		fmt.Fprintf(f.Writer, "Error encoding to JSON: %v\n", err)
		return err
	}
	return nil
}

// formatYAML formats data as YAML
func (f *Formatter) formatYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.Writer)
	err := encoder.Encode(data)
	if err != nil {
		fmt.Fprintf(f.Writer, "Error encoding to YAML: %v\n", err)
		return err
	}
	return nil
}

// formatTable formats data as a table
func (f *Formatter) formatTable(data interface{}) error {
	v := reflect.ValueOf(data)
	
	// Handle nil pointers and interfaces
	if !v.IsValid() {
		fmt.Fprintln(f.Writer, "No data available")
		return nil
	}
	
	// Dereference pointers safely
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			fmt.Fprintln(f.Writer, "No data available (nil pointer)")
			return nil
		}
		v = v.Elem()
	}

	// Handle different types
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return f.formatSliceAsTable(v)
	case reflect.Struct:
		return f.formatStructAsTable(v)
	case reflect.Map:
		return f.formatMapAsTable(v)
	default:
		// For simple types, just print the value
		fmt.Fprintf(f.Writer, "%v\n", data)
		return nil
	}
}

// createTabWriter creates a tabwriter with the formatter's options
func (f *Formatter) createTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(
		f.Writer,
		f.TableOptions.MinWidth,
		f.TableOptions.TabWidth,
		f.TableOptions.Padding,
		f.TableOptions.PadChar,
		f.TableOptions.Flags,
	)
}

// truncateValue truncates a string value if it exceeds the max column width
func (f *Formatter) truncateValue(value string) string {
	if f.TableOptions.MaxColumnWidth > 0 && len(value) > f.TableOptions.MaxColumnWidth {
		return value[:f.TableOptions.MaxColumnWidth-3] + "..."
	}
	return value
}

// safeString safely converts a value to a string, handling potential panics
func (f *Formatter) safeString(v reflect.Value) string {
	defer func() {
		if r := recover(); r != nil {
			// If we panic during string conversion, return a safe fallback
			fmt.Fprintf(f.Writer, "Warning: Error converting value to string: %v\n", r)
		}
	}()
	
	// Handle nil values
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		return "<nil>"
	}
	
	// For specific types that might need special handling
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() == 0 {
			return "<empty>"
		}
	}
	
	// Default string conversion
	return fmt.Sprintf("%v", v.Interface())
}

// formatSliceAsTable formats a slice as a table
func (f *Formatter) formatSliceAsTable(v reflect.Value) error {
	// Handle empty slices
	if v.Len() == 0 {
		fmt.Fprintln(f.Writer, "No data available (empty collection)")
		return nil
	}

	// Try to get the first element safely
	if v.Len() > 0 {
		firstElem := v.Index(0)
		
		// Handle slice of structs specially
		if firstElem.Kind() == reflect.Struct {
			return f.formatStructSliceAsTable(v)
		} else if firstElem.Kind() == reflect.Pointer && 
			!firstElem.IsNil() && 
			firstElem.Elem().Kind() == reflect.Struct {
			// Handle slice of struct pointers
			return f.formatStructPtrSliceAsTable(v)
		}
	}
	
	// For non-struct slices, use the simple format
	w := f.createTabWriter()
	defer w.Flush()
	
	// Print a simple header for non-struct slices
	if f.TableOptions.Header {
		fmt.Fprintln(w, "VALUE")
	}
	
	// Print each value safely
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		value := f.safeString(elem)
		fmt.Fprintf(w, "%s\n", f.truncateValue(value))
	}
	
	return nil
}

// formatStructSliceAsTable formats a slice of structs as a table with headers
func (f *Formatter) formatStructSliceAsTable(v reflect.Value) error {
	// Handle empty slices
	if v.Len() == 0 {
		fmt.Fprintln(f.Writer, "No data available (empty collection)")
		return nil
	}
	
	w := f.createTabWriter()
	defer w.Flush()
	
	// Get the struct type from the first element
	firstElem := v.Index(0)
	if !firstElem.IsValid() {
		fmt.Fprintln(f.Writer, "Error: Invalid first element in slice")
		return nil
	}
	
	structType := firstElem.Type()
	
	// Extract field names for headers, considering JSON tags
	headers := make([]string, 0)
	fieldIndices := make([]int, 0)
	
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Check for json tag to use as header
		header := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "-" { // Skip fields with json:"-"
				if parts[0] != "" {
					header = parts[0]
				}
				headers = append(headers, header)
				fieldIndices = append(fieldIndices, i)
			}
		} else {
			headers = append(headers, header)
			fieldIndices = append(fieldIndices, i)
		}
	}
	
	// Handle case where no fields were found
	if len(headers) == 0 {
		fmt.Fprintln(f.Writer, "No displayable fields found in struct")
		return nil
	}
	
	// Print headers
	if f.TableOptions.Header {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	
	// Print each row
	for i := 0; i < v.Len(); i++ {
		row := v.Index(i)
		if !row.IsValid() {
			continue // Skip invalid rows
		}
		
		values := make([]string, len(fieldIndices))
		
		for j, fieldIdx := range fieldIndices {
			if fieldIdx >= row.NumField() {
				values[j] = "<error>"
				continue
			}
			
			field := row.Field(fieldIdx)
			value := f.safeString(field)
			values[j] = f.truncateValue(value)
		}
		
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	
	return nil
}

// formatStructPtrSliceAsTable formats a slice of struct pointers as a table with headers
func (f *Formatter) formatStructPtrSliceAsTable(v reflect.Value) error {
	// Handle empty slices
	if v.Len() == 0 {
		fmt.Fprintln(f.Writer, "No data available (empty collection)")
		return nil
	}
	
	w := f.createTabWriter()
	defer w.Flush()
	
	// Get the struct type from the first element
	firstElem := v.Index(0)
	if !firstElem.IsValid() || firstElem.IsNil() {
		fmt.Fprintln(f.Writer, "Error: First element is nil or invalid")
		return nil
	}
	
	elemValue := firstElem.Elem()
	if !elemValue.IsValid() {
		fmt.Fprintln(f.Writer, "Error: Invalid element value")
		return nil
	}
	
	structType := elemValue.Type()
	
	// Extract field names for headers, considering JSON tags
	headers := make([]string, 0)
	fieldIndices := make([]int, 0)
	
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Check for json tag to use as header
		header := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "-" { // Skip fields with json:"-"
				if parts[0] != "" {
					header = parts[0]
				}
				headers = append(headers, header)
				fieldIndices = append(fieldIndices, i)
			}
		} else {
			headers = append(headers, header)
			fieldIndices = append(fieldIndices, i)
		}
	}
	
	// Handle case where no fields were found
	if len(headers) == 0 {
		fmt.Fprintln(f.Writer, "No displayable fields found in struct")
		return nil
	}
	
	// Print headers
	if f.TableOptions.Header {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	
	// Print each row
	for i := 0; i < v.Len(); i++ {
		ptrValue := v.Index(i)
		if !ptrValue.IsValid() || ptrValue.IsNil() {
			// Handle nil pointers in the slice
			values := make([]string, len(fieldIndices))
			for j := range values {
				values[j] = "<nil>"
			}
			fmt.Fprintln(w, strings.Join(values, "\t"))
			continue
		}
		
		row := ptrValue.Elem() // Dereference the pointer
		if !row.IsValid() {
			continue // Skip invalid rows
		}
		
		values := make([]string, len(fieldIndices))
		
		for j, fieldIdx := range fieldIndices {
			if fieldIdx >= row.NumField() {
				values[j] = "<error>"
				continue
			}
			
			field := row.Field(fieldIdx)
			value := f.safeString(field)
			values[j] = f.truncateValue(value)
		}
		
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	
	return nil
}

// formatStructAsTable formats a struct as a table
func (f *Formatter) formatStructAsTable(v reflect.Value) error {
	// Check if the struct is valid
	if !v.IsValid() {
		fmt.Fprintln(f.Writer, "Error: Invalid struct value")
		return nil
	}
	
	w := f.createTabWriter()
	defer w.Flush()

	if f.TableOptions.Header {
		fmt.Fprintln(w, "FIELD\tVALUE")
	}
	
	t := v.Type()
	fieldCount := 0
	
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		
		// Check for json tag to use as field name
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] == "-" {
				continue // Skip fields with json:"-"
			}
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}
		
		fieldValue := v.Field(i)
		value := f.safeString(fieldValue)
		fmt.Fprintf(w, "%s\t%s\n", fieldName, f.truncateValue(value))
		fieldCount++
	}
	
	// Handle case where no fields were found
	if fieldCount == 0 {
		fmt.Fprintln(f.Writer, "No displayable fields found in struct")
	}
	
	return nil
}

// formatMapAsTable formats a map as a table
func (f *Formatter) formatMapAsTable(v reflect.Value) error {
	// Check if the map is valid
	if !v.IsValid() {
		fmt.Fprintln(f.Writer, "Error: Invalid map value")
		return nil
	}
	
	// Handle nil maps
	if v.IsNil() {
		fmt.Fprintln(f.Writer, "No data available (nil map)")
		return nil
	}
	
	// Handle empty maps
	if v.Len() == 0 {
		fmt.Fprintln(f.Writer, "No data available (empty map)")
		return nil
	}
	
	w := f.createTabWriter()
	defer w.Flush()

	if f.TableOptions.Header {
		fmt.Fprintln(w, "KEY\tVALUE")
	}
	
	// Use MapRange to safely iterate over the map
	iter := v.MapRange()
	for iter.Next() {
		key := f.safeString(iter.Key())
		value := f.safeString(iter.Value())
		fmt.Fprintf(w, "%s\t%s\n", f.truncateValue(key), f.truncateValue(value))
	}
	
	return nil
}
