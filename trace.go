package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var (
	importName = "__log"
	importPath = `"github.com/jbardin/gotrace/log"`

	importStmt = `
import __log "github.com/jbardin/gotrace/log"
`

	setup = `
var _ = __log.Setup("stderr", "%s", %d)
`

	tmpl = `
__traceID := __log.Next(){{if .timing}}
__start := __log.Now(){{end}}
__log.L.Printf("[%d] {{.fname}}(%s){{if .position}} [{{.position}}]{{ end }}\n", __traceID, __log.Format({{.args}}))
{{if .return}}defer func() {
	{{if .timing}}since := "in " + __log.Since(__start).String(){{else}}since := ""{{end}}
	__log.L.Printf("[%d] {{.fname}}{{if .position}} [{{.position}}]{{ end }} returned %s\n", __traceID, since)
}(){{ end }}
`
)

var (
	fset         *token.FileSet
	funcTemplate *template.Template
	showReturn   bool
	exportedOnly bool
	prefix       string
	showPackage  bool
	writeFiles   bool
	filterFlag   string
	excludeFlag  string
	formatLength int
	timing       bool

	filter  *regexp.Regexp
	exclude *regexp.Regexp
)

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

func debugCall(fName string, pos token.Pos, args ...string) []byte {
	vals := make(map[string]string)
	if len(args) > 0 {
		vals["args"] = strings.Join(args, ", ")
	} else {
		vals["args"] = ""
	}

	if timing {
		vals["timing"] = "true"
	}

	vals["fname"] = fName

	if pos.IsValid() {
		vals["position"] = fset.Position(pos).String()
	}

	if showReturn {
		vals["return"] = "true"
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

	var pos token.Pos
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

		// prepend our receiver type
		if n.Recv != nil && len(n.Recv.List) > 0 {
			switch t := n.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				funcName = t.X.(*ast.Ident).Name + "." + funcName
			case *ast.Ident:
				funcName = t.Name + "." + funcName
			}
		}

	case *ast.FuncLit:
		body = n.Body
		funcType = n.Type
		funcName = "func"
		pos = n.Pos()

	default:
		return true
	}

	if exportedOnly && !ast.IsExported(funcName) {
		return true
	}

	if showPackage {
		funcName = e.packageName + "." + funcName
	}

	if !filter.MatchString(funcName) {
		return true
	}

	if exclude != nil && exclude.MatchString(funcName) {
		return true
	}

	e.Add(int(body.Lbrace), debugCall(funcName, pos, paramNames(funcType.Params)...))

	return true
}

func annotate(file string) {
	fset = token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, imp := range f.Imports {
		if imp.Name != nil && imp.Name.Name == importName {
			log.Printf(`"%s" already imported. skipping %s`, importName, file)
			return
		}
		if imp.Path.Value == importPath {
			log.Printf(`"%s", already imported. skipping %s`, importPath, file)
			return
		}
	}

	edits := editList{packageName: f.Name.Name}

	// insert our import directly after the package line
	edits.Add(int(f.Name.End()), []byte(importStmt))

	ast.Inspect(f, edits.inspect)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		log.Fatal("format.Node", err)
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
		log.Fatal("format.Source ", err)
	}

	if !writeFiles {
		fmt.Println(string(src))
		return
	}

	err = ioutil.WriteFile(file, src, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	funcTemplate = template.Must(template.New("debug").Parse(tmpl))
}

func main() {
	flag.BoolVar(&showReturn, "returns", false, "show function return")
	flag.BoolVar(&exportedOnly, "exported", false, "only annotate exported functions")
	flag.StringVar(&prefix, "prefix", "", "log prefix")
	flag.BoolVar(&showPackage, "package", false, "show package name prefix on function calls")
	flag.BoolVar(&writeFiles, "w", false, "re-write files in place")
	flag.StringVar(&filterFlag, "filter", ".", "only annotate functions matching the regular expression")
	flag.StringVar(&excludeFlag, "exclude", "", "exclude any matching functions, takes precedence over filter")
	flag.IntVar(&formatLength, "formatLength", 1024, "limit the formatted length of each argumnet to 'size'")
	flag.BoolVar(&timing, "timing", false, "print function durations. Implies -returns")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	setup = fmt.Sprintf(setup, prefix, formatLength)

	var err error
	filter, err = regexp.Compile(filterFlag)
	if err != nil {
		log.Fatal(err)
	}

	if excludeFlag != "" {
		exclude, err = regexp.Compile(excludeFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	if timing {
		showReturn = true
	}

	for _, file := range flag.Args() {
		annotate(file)
	}
}
