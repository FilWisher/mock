package main

import (
	"fmt"
	"strings"
	"text/template"
	"os"
	"io"
	"bufio"

	"go/ast"
	"go/parser"
	"go/token"
	"go/printer"
)

type StructMethod struct {
	Name string
	Params string
	Results string
	CallArgs string
}

type Struct struct {
	Name string
	Methods []StructMethod
}

var tmplString = `
type Mock{{.Name}} struct {
	{{- range .Methods }}
	{{ .Name }}Fn func({{.Params}}) {{.Results}}
	{{- end }}
}
{{ range .Methods }}
func (o Mock{{$.Name}}) {{ .Name }}({{.Params}}) {{.Results}} {
	return o.{{.Name}}Fn({{.CallArgs}})
}
{{ end -}}
`

var tmpl = template.Must(template.New("").Parse(tmplString))

func (s Struct) String() string {
	var b strings.Builder
	err := tmpl.Execute(&b, s)
	if err != nil {
		panic(err)
	}
// 	return strings.TrimRight(b.String(), "\n\t ")
 	return b.String()
}

// A `<name> <type>` pair used as parameters and return types for function
// declarations
type Arg struct {
	Name string
	Type string
}

// An interface method
type Method struct {
	Name string
	TypeParams []*Arg
	Params []*Arg
	Results []*Arg
}

// An interface type
type Interface struct {
	Name string
	Methods []Method
}

// prints the name of a type as it was declared
func _print(fset *token.FileSet, node any) string {
	var b strings.Builder
	printer.Fprint(&b, fset, node)
	return b.String()
}

func Name(idents []*ast.Ident) string {
	var parts []string
	for _, ident := range idents {
		parts = append(parts, ident.Name)
	}
	return strings.Join(parts, ".")
}

func Names(fields []*ast.Field) []string {
	var out []string
	for _, field := range fields {
		out = append(out, Name(field.Names))
	}
	return out
}

func (i *Interface) Defaults() {

	// This will only support up to 26 unnamed parameters. If anyone is
	// using more than that, they deserve an "index out of range"
	// error anyway.
	alphabet := "abcdefghijklmnopqrstuvwxyz"

	for _, method := range i.Methods {

		i := 0

		for _, param := range method.Params {
			if param.Name == "" {
				param.Name = string(alphabet[i])
				i += 1
			}

			if param.Name == " " {
				param.Name = "_"
			}
		}

		// Check whether we're using named params
		named := false
		for _, param := range method.Results {
			if param.Name != "" {
				named = true
				break
			}
		}

		for _, param := range method.Results {
			if named && param.Name == "" {
				param.Name = string(alphabet[i])
				i += 1
			}

			if param.Name == " " {
				param.Name = "_"
			}
		}
	}
}

func (m Method) String() string {
	parts := []string{
		fmt.Sprintf("func %s(", m.Name),
	}

	var args []string

	for _, param := range m.Params {
		args = append(args, fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	parts = append(parts, strings.Join(args, ", "), ")")


	if len(m.Results) > 1 {
		parts = append(parts, "(")
	}


	var results []string
	for _, param := range m.Results {
		results = append(results, fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	parts = append(parts, strings.Join(results, ", "))

	if len(m.Results) > 1 {
		parts = append(parts, ")")
	}

	return strings.Join(parts, "")
}

func (i *Interface) String() string {
	parts := []string{
		fmt.Sprintf("type %s interface {", i.Name),
	}

	for _, method := range i.Methods {
		parts = append(parts, "\t" + method.String())
	}

	parts = append(parts, "}")

	return strings.Join(parts, "\n")
}

func (i *Interface) ToStruct() Struct {
	s := Struct{Name: i.Name}


	for _, method := range i.Methods {

		var args []string
		var callArgs []string

		for _, param := range method.Params {
			args = append(args, fmt.Sprintf("%s %s", param.Name, param.Type))
			callArgs = append(callArgs, param.Name)
		}

		var parts []string

		if len(method.Results) > 1 {
			parts = append(parts, "(")
		}

		var results []string
		for _, param := range method.Results {
			if param.Name == "" {
				results = append(results, fmt.Sprintf("%s", param.Type))
			} else {
				results = append(results, fmt.Sprintf("%s %s", param.Name, param.Type))
			}
		}

		parts = append(parts, strings.Join(results, ", "))

		if len(method.Results) > 1 {
			parts = append(parts, ")")
		}

		s.Methods = append(s.Methods, StructMethod{
			Name: method.Name,
			Params: strings.Join(args, ", "),
			Results: strings.Join(parts, ""),
			CallArgs: strings.Join(callArgs, ", "),
		})
	}

	return s
}

func slurp(r io.Reader) (string, error) {
	var b strings.Builder
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		b.WriteString(sc.Text())
		b.WriteString("\n")
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return b.String(), nil
}

func parse(raw string) (*Interface, error) {

	src := fmt.Sprintf(`package main

%s`, raw)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "demo", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	iface := &Interface{}

	ast.Inspect(file, func(x ast.Node) bool {
		t, ok := x.(*ast.TypeSpec)
		if ok {
			iface.Name = t.Name.Name
			return true
		}

		s, ok := x.(*ast.InterfaceType)
		if !ok {
			return true
		}

		for _, field := range s.Methods.List {

			method := Method{
				Name: Name(field.Names),
			}
			
			typ := field.Type.(*ast.FuncType)

			if typ.TypeParams != nil {
				for _, field := range typ.TypeParams.List {
					method.TypeParams = append(method.TypeParams, &Arg{
						Name: Name(field.Names),
						Type: _print(fset, field.Type),
					})
				}
			}

			if typ.Params != nil {
				for _, field := range typ.Params.List {
					method.Params = append(method.Params, &Arg{
						Name: Name(field.Names),
						Type: _print(fset, field.Type),
					})
				}
			}

			if typ.Results != nil {
				for _, field := range typ.Results.List {
					method.Results = append(method.Results, &Arg{
						Name: Name(field.Names),
						Type: _print(fset, field.Type),
					})
				}
			}

			iface.Methods = append(iface.Methods, method)
		}
		return false
	})

	iface.Defaults()
	return iface, nil
}

func generate(raw string, w io.Writer) error {
	iface, err := parse(raw)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s%s", raw, iface.ToStruct())
	return nil
}

func main() {
	raw, err := slurp(os.Stdin)
	if err != nil {
		panic(err)
	}

	generate(raw, os.Stdout)
}
