// Command generate produces Terraform resource files (model, schema, resource)
// from declarative YAML schemas, eliminating the boilerplate of the four-file
// pattern for the many near-identical OPNsense resources.
//
// Usage: go run ./internal/generate  (or: go generate ./...)
//
// It reads every *.yaml under internal/generate/schemas and writes
// {name}_model.gen.go, {name}_schema.gen.go, {name}_resource.gen.go into
// internal/service/{package}. Generated files carry a DO NOT EDIT header.
package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Field describes one attribute of a resource.
type Field struct {
	Name     string   `yaml:"name"`
	TF       string   `yaml:"tf"`
	JSON     string   `yaml:"json"`
	Type     string   `yaml:"type"` // bool|string|selectmap|selectmaplist|csvset|int
	Required bool     `yaml:"required"`
	Default  string   `yaml:"default"`
	Desc     string   `yaml:"desc"`
	Options  []string `yaml:"options"`
	// TestValue is a raw HCL literal used in the generated acceptance test for
	// this field (e.g. `"permit"`, `["ipv4"]`, `65010`). Required for required
	// selectmap/set fields, which have no sensible auto value.
	TestValue string `yaml:"test_value"`
}

// Resource describes a single generated resource.
type Resource struct {
	Name        string            `yaml:"name"`
	GoType      string            `yaml:"go_type"`
	TypeName    string            `yaml:"type_name"`
	Title       string            `yaml:"title"`
	Kind        string            `yaml:"kind"` // item|singleton
	ID          string            `yaml:"id"`
	Reconfigure string            `yaml:"reconfigure"`
	Monad       string            `yaml:"monad"`
	Endpoints   map[string]string `yaml:"endpoints"`
	Fields      []Field           `yaml:"fields"`
}

// Schema is one YAML schema file.
type Schema struct {
	Package   string     `yaml:"package"`
	Resources []Resource `yaml:"resources"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
}

func run() error {
	schemaDir := filepath.Join("internal", "generate", "schemas")
	entries, err := filepath.Glob(filepath.Join(schemaDir, "*.yaml"))
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no schemas found in %s", schemaDir)
	}
	count := 0
	for _, path := range entries {
		data, err := os.ReadFile(path) //nolint:gosec // schema path from glob
		if err != nil {
			return err
		}
		var s Schema
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for i := range s.Resources {
			if err := generateResource(s.Package, &s.Resources[i]); err != nil {
				return fmt.Errorf("%s/%s: %w", s.Package, s.Resources[i].Name, err)
			}
			count++
		}
	}
	fmt.Printf("generated %d resources\n", count)
	return nil
}

func generateResource(pkg string, r *Resource) error {
	dir := filepath.Join("internal", "service", pkg)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	files := []struct {
		kind    string
		tmpl    *template.Template
		imports string
	}{
		{"model", modelTmpl, modelImports(r)},
		{"schema", schemaTmpl, schemaImports(r)},
		{"resource", resourceTmpl, resourceImports(r)},
		{"resource_test", testTmpl, testImports(r)},
	}
	for _, f := range files {
		var buf strings.Builder
		err := f.tmpl.Execute(&buf, map[string]any{"Pkg": pkg, "R": r, "Imports": f.imports})
		if err != nil {
			return err
		}
		formatted, err := format.Source([]byte(buf.String()))
		if err != nil {
			return fmt.Errorf("%s: format: %w\n---\n%s", f.kind, err, buf.String())
		}
		// Test files must end in _test.go for Go to treat them as tests; the
		// _gen_ infix + DO NOT EDIT header mark them generated.
		name := fmt.Sprintf("%s_%s.gen.go", r.Name, f.kind)
		if f.kind == "resource_test" {
			name = fmt.Sprintf("%s_resource_gen_test.go", r.Name)
		}
		out := filepath.Join(dir, name)
		if err := os.WriteFile(out, formatted, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// --- import computation (grouped: stdlib / framework / local) ---

const (
	pkgTypes    = `"github.com/hashicorp/terraform-plugin-framework/types"`
	pkgOpnsense = `"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"`
	pkgTfconv   = `"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"`
)

func renderImports(stdlib, framework, local []string) string {
	groups := [][]string{stdlib, framework, local}
	var parts []string
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		sort.Strings(g)
		parts = append(parts, "\t"+strings.Join(g, "\n\t"))
	}
	return strings.Join(parts, "\n\n")
}

func modelImports(r *Resource) string {
	local := []string{pkgOpnsense}
	if hasSet(r) || hasInt(r) {
		local = append(local, pkgTfconv)
	}
	return renderImports(
		[]string{`"context"`},
		[]string{pkgTypes},
		local,
	)
}

func hasInt(r *Resource) bool {
	for _, f := range r.Fields {
		if f.Type == "int" {
			return true
		}
	}
	return false
}

func schemaImports(r *Resource) string {
	fw := []string{
		`"github.com/hashicorp/terraform-plugin-framework/resource"`,
		`"github.com/hashicorp/terraform-plugin-framework/resource/schema"`,
		`"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"`,
		`"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"`,
	}
	add := map[string]bool{}
	for _, f := range r.Fields {
		switch f.Type {
		case "bool":
			add[`"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"`] = true
		case "int":
			if !f.Required {
				add[`"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"`] = true
			}
		case "selectmaplist", "csvset":
			add[pkgTypes] = true
		case "string", "selectmap":
			if !f.Required {
				add[`"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"`] = true
			}
			if len(f.Options) > 0 {
				add[`"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"`] = true
				add[`"github.com/hashicorp/terraform-plugin-framework/schema/validator"`] = true
			}
		}
	}
	for k := range add {
		fw = append(fw, k)
	}
	return renderImports([]string{`"context"`}, fw, nil)
}

func resourceImports(r *Resource) string {
	std := []string{`"context"`, `"errors"`, `"fmt"`}
	fw := []string{
		`"github.com/hashicorp/terraform-plugin-framework/path"`,
		`"github.com/hashicorp/terraform-plugin-framework/resource"`,
	}
	if r.Kind == "singleton" {
		fw = append(fw, pkgTypes)
	}
	return renderImports(std, fw, []string{pkgOpnsense})
}

func testImports(r *Resource) string {
	local := []string{`"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"`}
	if r.Kind == "item" {
		local = append(local, pkgOpnsense)
	}
	return renderImports(
		[]string{`"testing"`},
		[]string{`"github.com/hashicorp/terraform-plugin-testing/helper/resource"`},
		local,
	)
}

// testFieldsHCL emits HCL assignments for the fields a generated acceptance test
// must set: every required field plus any field with an explicit test_value.
func testFieldsHCL(r *Resource) string {
	var b strings.Builder
	for _, f := range r.Fields {
		if !f.Required && f.TestValue == "" {
			continue
		}
		v := f.TestValue
		if v == "" {
			switch f.Type {
			case "bool":
				v = "true"
			case "int":
				v = "1"
			default:
				v = `"test"`
			}
		}
		fmt.Fprintf(&b, "  %s = %s\n", f.TF, v)
	}
	return b.String()
}

// --- template helpers ---

var funcs = template.FuncMap{
	"camel":       camelName,
	"goType":      goType,
	"respType":    respType,
	"toAPI":       toAPILine,
	"fromAPI":     fromAPILine,
	"schemaAttr":  schemaAttr,
	"isItem":      func(r *Resource) bool { return r.Kind == "item" },
	"isSingleton": func(r *Resource) bool { return r.Kind == "singleton" },
	"hasSet":      hasSet,
	"testFields":  testFieldsHCL,
}

func hasSet(r *Resource) bool {
	for _, f := range r.Fields {
		if f.Type == "selectmaplist" || f.Type == "csvset" {
			return true
		}
	}
	return false
}

func goType(f Field) string {
	switch f.Type {
	case "bool":
		return "types.Bool"
	case "int":
		return "types.Int64"
	case "selectmaplist", "csvset":
		return "types.Set"
	default:
		return "types.String"
	}
}

func respType(f Field) string {
	switch f.Type {
	case "selectmap":
		return "opnsense.SelectedMap"
	case "selectmaplist":
		return "opnsense.SelectedMapList"
	default:
		return "string"
	}
}

func toAPILine(f Field) string {
	switch f.Type {
	case "bool":
		return fmt.Sprintf("opnsense.BoolToString(m.%s.ValueBool())", f.Name)
	case "int":
		if f.Required {
			return fmt.Sprintf("opnsense.Int64ToString(m.%s.ValueInt64())", f.Name)
		}
		return fmt.Sprintf("tfconv.IntOrEmpty(m.%s.ValueInt64())", f.Name)
	case "selectmaplist", "csvset":
		return fmt.Sprintf("tfconv.SetToCSV(ctx, m.%s)", f.Name)
	default:
		return fmt.Sprintf("m.%s.ValueString()", f.Name)
	}
}

func fromAPILine(f Field) string {
	switch f.Type {
	case "bool":
		return fmt.Sprintf("m.%s = types.BoolValue(opnsense.StringToBool(a.%s))", f.Name, f.Name)
	case "int":
		return fmt.Sprintf("m.%s = types.Int64Value(tfconv.IntOrZero(a.%s))", f.Name, f.Name)
	case "selectmap":
		return fmt.Sprintf("m.%s = types.StringValue(string(a.%s))", f.Name, f.Name)
	case "selectmaplist":
		return fmt.Sprintf("m.%s = tfconv.SliceToSet(a.%s)", f.Name, f.Name)
	case "csvset":
		return fmt.Sprintf("m.%s = tfconv.SliceToSet(opnsense.CSVToSlice(a.%s))", f.Name, f.Name)
	default:
		return fmt.Sprintf("m.%s = types.StringValue(a.%s)", f.Name, f.Name)
	}
}

func schemaAttr(f Field) string {
	var b strings.Builder
	switch f.Type {
	case "bool":
		def := "false"
		if f.Default == "true" {
			def = "true"
		}
		fmt.Fprintf(&b, "%q: schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(%s), MarkdownDescription: %q},", f.TF, def, f.Desc)
	case "int":
		if f.Required {
			fmt.Fprintf(&b, "%q: schema.Int64Attribute{Required: true, MarkdownDescription: %q},", f.TF, f.Desc)
		} else {
			d := f.Default
			if d == "" {
				d = "0"
			}
			fmt.Fprintf(&b, "%q: schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(%s), MarkdownDescription: %q},", f.TF, d, f.Desc)
		}
	case "selectmaplist", "csvset":
		fmt.Fprintf(&b, "%q: schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: %q},", f.TF, f.Desc)
	default:
		if f.Required {
			fmt.Fprintf(&b, "%q: schema.StringAttribute{Required: true, MarkdownDescription: %q", f.TF, f.Desc)
		} else {
			fmt.Fprintf(&b, "%q: schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(%q), MarkdownDescription: %q", f.TF, f.Default, f.Desc)
		}
		if len(f.Options) > 0 {
			quoted := make([]string, len(f.Options))
			for i, o := range f.Options {
				quoted[i] = fmt.Sprintf("%q", o)
			}
			fmt.Fprintf(&b, ", Validators: []validator.String{stringvalidator.OneOf(%s)}", strings.Join(quoted, ", "))
		}
		b.WriteString("},")
	}
	return b.String()
}

func init() {
	modelTmpl = template.Must(template.New("model").Funcs(funcs).Parse(modelText))
	schemaTmpl = template.Must(template.New("schema").Funcs(funcs).Parse(schemaText))
	resourceTmpl = template.Must(template.New("resource").Funcs(funcs).Parse(resourceText))
	testTmpl = template.Must(template.New("test").Funcs(funcs).Parse(testText))
}

var (
	modelTmpl    *template.Template
	schemaTmpl   *template.Template
	resourceTmpl *template.Template
	testTmpl     *template.Template
)
