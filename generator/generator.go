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

	g.p("package ", def.Package)
	g.p()

	if len(def.DefaultImports) > 0 {
		g.generateImported(def)
	}

	g.generateStruct(def)

	if def.GenSchema.sharedValidations.HasValidations {
		g.generateValidator(def)
	}

	for _, prop := range def.Properties {
		if prop.sharedValidations.HasValidations {
			g.generatePropValidator(def.Name, &prop)
		}
	}

}

func (g *Generator) generateHandler(buf *bytes.Buffer, op *GenOperation) {
}

func (g *Generator) generateParameterModel(buf *bytes.Buffer, op *GenOperation) {
}

// P prints the arguments to the generated output.  It handles strings and int32s, plus
// handling indirections because they may be *string, etc.
func (g *Generator) p(str ...interface{}) {
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
	g.p("import (")
	for _, imprt := range def.DefaultImports {
		g.p("\"", imprt, "\"")
	}
	g.p(")")
	g.p()
}

func (g *Generator) generateStruct(def *GenDefinition) {
	g.p("type ", def.GenSchema.Name, " struct {")
	for _, prop := range def.Properties {
		if prop.sharedValidations.Required {
			g.p(g.caps(prop.Name), " ", prop.resolvedType.GoType, " `json:\"", prop.Name, "\" binding:\"required\"`")
		} else {
			g.p(g.caps(prop.Name), " ", prop.resolvedType.GoType, " `json:\"", prop.Name, ",omitempty\"`")
		}
	}
	g.p("}")
	g.p()
}

func (g *Generator) generateValidator(def *GenDefinition) {
	g.p("func (m *", def.Name, ") Validate() error {")
	for _, prop := range def.Properties {
		if prop.sharedValidations.HasValidations {
			g.p("if err := m.validate", g.caps(prop.Name), "(); err != nil {")
			g.p("	return err")
			g.p("}")
			g.p()
		}
	}
	g.p("	return nil")
	g.p("}")
	g.p()
}

func (g *Generator) generatePropValidator(model string, prop *GenSchema) {
	g.p("func (m *", model, ") validate", g.caps(prop.Name), "() error {")
	if prop.sharedValidations.MaxLength != nil {
		g.p("if err := validate.MaxLength(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", g.caps(prop.Name), "), ", prop.sharedValidations.MaxLength, "); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	if prop.sharedValidations.MinLength != nil {
		g.p("if err := validate.MinLength(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", g.caps(prop.Name), "), ", prop.sharedValidations.MinLength, "); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	if prop.sharedValidations.Pattern != "" {
		g.p("if err := validate.Pattern(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", g.caps(prop.Name), "), `", prop.sharedValidations.Pattern, "`); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	if prop.sharedValidations.MultipleOf != nil {
		g.p("if err := validate.MultipleOf(\"", prop.Name, "\", \"body\", ", "float64(m.", g.caps(prop.Name), "), ", prop.sharedValidations.MultipleOf, "); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	if prop.sharedValidations.Minimum != nil {
		g.p("if err := validate.Minimum(\"", prop.Name, "\", \"body\", ", "float64(m.", g.caps(prop.Name), "), ", prop.sharedValidations.Minimum, ", false); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	if prop.sharedValidations.Maximum != nil {
		g.p("if err := validate.Maximum(\"", prop.Name, "\", \"body\", ", "float64(m.", g.caps(prop.Name), "), ", prop.sharedValidations.Maximum, ", false); err != nil {")
		g.p("	return err")
		g.p("	}")
		g.p()
	}
	g.p("	return nil")
	g.p("}")
	g.p()
}
