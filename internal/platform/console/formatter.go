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

// Formatter formats data for CLI output
type Formatter struct {
	OutFormat OutputFormat
	Writer io.Writer
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
		OutFormat: outputFormat,
		Writer: os.Stdout,
	}
}

// Format formats the data according to the formatter's format
func (f *Formatter) Format(data interface{}) error {
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
	return encoder.Encode(data)
}

// formatYAML formats data as YAML
func (f *Formatter) formatYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.Writer)
	return encoder.Encode(data)
}

// formatTable formats data as a table
func (f *Formatter) formatTable(data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		return f.formatSliceAsTable(v)
	case reflect.Struct:
		return f.formatStructAsTable(v)
	case reflect.Map:
		return f.formatMapAsTable(v)
	default:
		fmt.Fprintln(f.Writer, data)
		return nil
	}
}

// formatSliceAsTable formats a slice as a table
func (f *Formatter) formatSliceAsTable(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(f.Writer, "No data available")
		return nil
	}

	// Basic implementation - this would need to be enhanced for real use
	// with proper field extraction for headers
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()
	
	// Here we'd need to inspect the slice elements and build table headers
	// based on their type - for simplicity, let's just print values
	
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		fmt.Fprintf(w, "%v\n", elem.Interface())
	}
	
	return nil
}

// formatStructAsTable formats a struct as a table
func (f *Formatter) formatStructAsTable(v reflect.Value) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fmt.Fprintf(w, "%s\t%v\n", field.Name, v.Field(i).Interface())
		}
	}
	
	return nil
}

// formatMapAsTable formats a map as a table
func (f *Formatter) formatMapAsTable(v reflect.Value) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "KEY\tVALUE")
	iter := v.MapRange()
	for iter.Next() {
		fmt.Fprintf(w, "%v\t%v\n", iter.Key().Interface(), iter.Value().Interface())
	}
	
	return nil
}
