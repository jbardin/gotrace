package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	importStmt = `
import __log "github.com/jbardin/gotrace/log"
`

	setup = `
var _ = __log.Setup("stderr", "%s ")
`

	tmpl = `
__traceCount := __log.Next()
__log.L.Printf("[%d] {{.fname}}({{.formatters}}) \n", __traceCount, {{.args}})
{{ if .exit }}defer func() {
	__log.L.Printf("[%d] {{.fname}} exited\n", __traceCount)
}(){{ end }}
`
)

var (
	funcTemplate *template.Template
	showExit     bool
	exportedOnly bool
	prefix       string
	showPackage  bool
)

// return n '%v's for formatting
func formatters(n int) string {
	f := []string{}
	for i := 0; i < n; i++ {
		f = append(f, "%v")
	}
	return strings.Join(f, ", ")
}

// convert function parameters to a list of names
func paramNames(params *ast.FieldList) []string {
	var p []string
	for _, f := range params.List {
		for _, n := range f.Names {
			p = append(p, n.Name)
		}
	}
	return p
}

func debugCall(fName string, args ...string) []byte {
	vals := make(map[string]string)
	if len(args) > 0 {
		vals["formatters"] = formatters(len(args))
		vals["args"] = strings.Join(args, ", ")
	} else {
		vals["formatters"] = ""
		vals["args"] = ""
	}

	vals["fname"] = fName

	if showExit {
		vals["exit"] = "true"
	}

	var b bytes.Buffer
	err := funcTemplate.Execute(&b, vals)
	if err != nil {
		log.Fatal(err)
	}
	return b.Bytes()
}

type edit struct {
	pos int
	val []byte
}

type editList struct {
	edits       []edit
	packageName string
}

func (e *editList) Add(pos int, val []byte) {
	e.edits = append(e.edits, edit{pos: pos, val: val})
}

func (e *editList) inspect(node ast.Node) bool {
	if node == nil {
		return false
	}

	var funcType *ast.FuncType
	var body *ast.BlockStmt
	var funcName string

	switch n := node.(type) {
	case *ast.FuncDecl:
		body = n.Body
		if body == nil {
			return true
		}
		funcType = n.Type
		funcName = n.Name.Name

	case *ast.FuncLit:
		body = n.Body
		funcType = n.Type
		funcName = "func"

	default:
		return true
	}

	if exportedOnly && !ast.IsExported(funcName) {
		return true
	}

	if showPackage {
		funcName = e.packageName + "." + funcName
	}

	e.Add(int(body.Lbrace), debugCall(funcName, paramNames(funcType.Params)...))

	return true
}

func annotate(file string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	edits := editList{packageName: f.Name.Name}

	// insert our import directly after the package line
	edits.Add(int(f.Name.End()), []byte(importStmt))

	ast.Inspect(f, edits.inspect)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		panic(err)
	}

	data := buf.Bytes()

	var pos int
	var out []byte
	for _, e := range edits.edits {
		out = append(out, data[pos:e.pos]...)
		out = append(out, []byte(e.val)...)
		pos = e.pos
	}
	out = append(out, data[pos:]...)

	// it's easier to append the setup code at the end
	out = append(out, []byte(setup)...)

	src, err := format.Source(out)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(src))
}

func init() {
	funcTemplate = template.Must(template.New("debug").Parse(tmpl))
}

func main() {
	flag.BoolVar(&showExit, "exits", false, "show function exits")
	flag.BoolVar(&exportedOnly, "exported", false, "only annotate exported functions")
	flag.StringVar(&prefix, "prefix", "\t", "log prefix")
	flag.BoolVar(&showPackage, "package", false, "show package name prefix on function calls")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("missing file name")
		os.Exit(1)
	}

	setup = fmt.Sprintf(setup, prefix)

	annotate(flag.Arg(0))
}
