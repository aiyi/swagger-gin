package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/asaskevich/govalidator"
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

// lowercase first character
func (g *Generator) lowerFirst(word string) string {
	return strings.ToLower(word[:1]) + word[1:]
}

// reports a problem and exits the program.
func (g *Generator) fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("error:", s)
	os.Exit(1)
}

func (g *Generator) hasExtendFormat(prop *GenSchema) bool {
	if prop.resolvedType.SwaggerType == "string" && prop.resolvedType.SwaggerFormat != "" {
		if _, ok := govalidator.TagMap[prop.resolvedType.SwaggerFormat]; ok {
			return true
		}
	}

	return false
}

// Fill the response protocol buffer with the generated output for all the files we're
// supposed to generate.
func (g *Generator) generateModel(buf *bytes.Buffer, def *GenDefinition) {
	g.Buffer = buf

	g.p("package ", def.Package)
	g.p()
	g.generateImported(def)
	g.generateStruct(def)

	for _, prop := range def.Properties {
		if g.hasExtendFormat(&prop) {
			def.GenSchema.sharedValidations.HasValidations = true
		}
	}

	if def.GenSchema.sharedValidations.HasValidations {
		g.generateValidator(def)
	}

	for _, prop := range def.Properties {
		if prop.sharedValidations.HasValidations || g.hasExtendFormat(&prop) {
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
	g.p("	\"encoding/json\"")
	g.p("	\"fmt\"")
	g.p("	\"time\"")
	g.p("	\"github.com/asaskevich/govalidator\"")
	g.p("	\"github.com/aiyi/httpkit/validate\"")
	g.p(")")
	g.p()
}

func (g *Generator) generateStruct(def *GenDefinition) {
	g.p("type ", def.GenSchema.Name, " struct {")
	for _, prop := range def.Properties {
		if g.hasExtendFormat(&prop) {
			prop.resolvedType.GoType = "string"
		} else if prop.resolvedType.SwaggerFormat == "date" {
				prop.resolvedType.GoType = "time.Time"
		} 
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
		if prop.sharedValidations.HasValidations || g.hasExtendFormat(&prop) {
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
	propName := g.caps(prop.Name)

	if prop.sharedValidations.Enum != nil {
		varEnum := g.lowerFirst(model) + propName + "Enum"
		jsonEnum, _ := json.Marshal(prop.sharedValidations.Enum)
		g.p("var ", varEnum, " []interface{}")
		g.p()
		g.p("func (m *", model, ") validate", propName, "Enum(path, location string, value ", prop.resolvedType.GoType, ") error {")
		g.p("	if ", varEnum, " == nil {")
		g.p("		var res []", prop.resolvedType.GoType)
		g.p("		if err := json.Unmarshal([]byte(`", string(jsonEnum), "`), &res); err != nil {")
		g.p("			return err")
		g.p("		}")
		g.p("		for _, v := range res {")
		g.p("			", varEnum, " = append(", varEnum, ", v)")
		g.p("		}")
		g.p("	}")
		g.p("	return validate.Enum(path, location, value, ", varEnum, ")")
		g.p("}")
		g.p()
	}

	g.p("func (m *", model, ") validate", propName, "() error {")
	if prop.sharedValidations.MaxLength != nil {
		g.p("if err := validate.MaxLength(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", propName, "), ", prop.sharedValidations.MaxLength, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.MinLength != nil {
		g.p("if err := validate.MinLength(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", propName, "), ", prop.sharedValidations.MinLength, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.Pattern != "" {
		g.p("if err := validate.Pattern(\"", prop.Name, "\", \"body\", ", prop.resolvedType.GoType, "(m.", propName, "), `", prop.sharedValidations.Pattern, "`); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.MultipleOf != nil {
		g.p("if err := validate.MultipleOf(\"", prop.Name, "\", \"body\", ", "float64(m.", propName, "), ", prop.sharedValidations.MultipleOf, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.Minimum != nil {
		exclusive := "false"
		if prop.sharedValidations.ExclusiveMinimum {
			exclusive = "true"
		}
		g.p("if err := validate.Minimum(\"", prop.Name, "\", \"body\", ", "float64(m.", propName, "), ", prop.sharedValidations.Minimum, ", ", exclusive, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.Maximum != nil {
		exclusive := "false"
		if prop.sharedValidations.ExclusiveMaximum {
			exclusive = "true"
		}
		g.p("if err := validate.Maximum(\"", prop.Name, "\", \"body\", ", "float64(m.", propName, "), ", prop.sharedValidations.Maximum, ", ", exclusive, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if prop.sharedValidations.Enum != nil {
		g.p("if err := validate", propName, "Enum(\"", prop.Name, "\", \"body\", ", "m.", propName, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if g.hasExtendFormat(prop) {
		validatefunc, _ := govalidator.TagMap[prop.resolvedType.SwaggerFormat]
		funcName := runtime.FuncForPC(reflect.ValueOf(validatefunc).Pointer()).Name()
		g.p("if ", funcName[22:], "(m.", propName, ") != true {")
		g.p("	return fmt.Errorf(\"invalid format of ", propName, "\")")
		g.p("}")
		g.p()
	}
	g.p("	return nil")
	g.p("}")
	g.p()
}
