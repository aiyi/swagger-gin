package generator

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	codeGen = NewGenerator()
)

// Generator is the type whose methods generate the output, stored in the associated response structure.
type Generator struct {
	*bytes.Buffer
	//def    *GenDefinition
	//op     *GenOperation
}

// New creates a new generator and allocates the request and response protobufs.
func NewGenerator() *Generator {
	g := new(Generator)
	return g
}

// uppercase first character
func (g *Generator) caps(word string) string {
	return strings.ToUpper(word[:1]) + word[1:]
}

// reports a problem and exits the program.
func (g *Generator) fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("error:", s)
	os.Exit(1)
}

// Fill the response protocol buffer with the generated output for all the files we're
// supposed to generate.
func (g *Generator) generateModel(buf *bytes.Buffer, def *GenDefinition) {
	g.Buffer = buf

	g.P("package ", def.Package)
	g.P()

	if len(def.DefaultImports) > 0 {
		g.generateImported(def)
	}

	g.generateStruct(def)

	if def.GenSchema.sharedValidations.HasValidations {
		g.generateValidator(def)
	}

	for _, prop := range def.Properties {
		if prop.sharedValidations.HasValidations {
			g.generatePropValidator(&prop)
		}
	}

}

func (g *Generator) generateHandler(buf *bytes.Buffer, op *GenOperation) {
}

func (g *Generator) generateParameterModel(buf *bytes.Buffer, op *GenOperation) {
}

// P prints the arguments to the generated output.  It handles strings and int32s, plus
// handling indirections because they may be *string, etc.
func (g *Generator) P(str ...interface{}) {
	for _, v := range str {
		switch s := v.(type) {
		case string:
			g.WriteString(s)
		case *string:
			g.WriteString(*s)
		case bool:
			g.WriteString(fmt.Sprintf("%t", s))
		case *bool:
			g.WriteString(fmt.Sprintf("%t", *s))
		case int:
			g.WriteString(fmt.Sprintf("%d", s))
		case *int32:
			g.WriteString(fmt.Sprintf("%d", *s))
		case *int64:
			g.WriteString(fmt.Sprintf("%d", *s))
		case float64:
			g.WriteString(fmt.Sprintf("%g", s))
		case *float64:
			g.WriteString(fmt.Sprintf("%g", *s))
		default:
			g.fail(fmt.Sprintf("unknown type in printer: %T", v))
		}
	}
	g.WriteByte('\n')
}

func (g *Generator) generateImported(def *GenDefinition) {
	g.P("import (")
	for _, imprt := range def.DefaultImports {
		g.P("\"", imprt, "\"")
	}
	g.P(")")
	g.P()
}

func (g *Generator) generateStruct(def *GenDefinition) {
	g.P("type ", def.GenSchema.Name, " struct {")
	for _, prop := range def.Properties {
		if prop.sharedValidations.Required {
			g.P(g.caps(prop.Name), " ", prop.resolvedType.GoType, " `json:\"", prop.Name, "\"`")
		} else {
			g.P(g.caps(prop.Name), " ", prop.resolvedType.GoType, " `json:\"", prop.Name, ",omitempty\"`")
		}
	}
	g.P("}")
	g.P()
}

func (g *Generator) generateValidator(def *GenDefinition) {
}

func (g *Generator) generatePropValidator(prop *GenSchema) {
}
