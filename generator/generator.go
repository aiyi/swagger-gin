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

	"github.com/aiyi/swagger-gin/spec"
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
	g.p("import (")
	g.p("	\"encoding/json\"")
	g.p("	\"fmt\"")
	g.p("	\"time\"")
	g.p("	\"github.com/asaskevich/govalidator\"")
	g.p("	\"github.com/aiyi/swagger-gin/validate\"")
	g.p(")")
	g.p()

	g.generateStruct(def)

	for _, prop := range def.Properties {
		if g.hasExtendFormat(&prop) {
			def.GenSchema.sharedValidations.HasValidations = true
		}
	}

	g.generateValidator(def)

	for _, prop := range def.Properties {
		if prop.sharedValidations.HasValidations || g.hasExtendFormat(&prop) {
			g.generatePropValidator(def.Name, &prop)
		}
	}
}

func (g *Generator) generateHandlers(buf *bytes.Buffer, specDoc *spec.Document) {
	g.Buffer = buf
	paths := specDoc.AllPaths()
	groups := make(map[string]string)

	g.p("package ", specDoc.Spec().Info.Title)
	g.p()
	g.p()
	g.p("import (")
	g.p("	\"github.com/gin-gonic/gin\"")
	g.p(")")
	g.p()

	for _, path := range paths {
		operations := path.PathItemProps
		if post := operations.Post; post != nil {
			tag := post.OperationProps.Tags[0]
			groups[tag] = g.caps(tag)
		}
		if get := operations.Get; get != nil {
			tag := get.OperationProps.Tags[0]
			groups[tag] = g.caps(tag)
		}
		if put := operations.Put; put != nil {
			tag := put.OperationProps.Tags[0]
			groups[tag] = g.caps(tag)
		}
		if del := operations.Delete; del != nil {
			tag := del.OperationProps.Tags[0]
			groups[tag] = g.caps(tag)
		}
	}

	g.p("var (")
	for _, group := range groups {
		g.p(group, " *gin.RouterGroup")
	}
	g.p(")")
	g.p()
	g.p("func AddRoutes() {")

	for group, _ := range groups {
		for pname, path := range paths {
			operations := path.PathItemProps
			if post := operations.Post; post != nil {
				g.generateRouter("POST", group, pname, post)
			}
			if get := operations.Get; get != nil {
				g.generateRouter("GET", group, pname, get)
			}
			if put := operations.Put; put != nil {
				g.generateRouter("PUT", group, pname, put)
			}
			if del := operations.Delete; del != nil {
				g.generateRouter("DELETE", group, pname, del)
			}
		}
		g.p()
	}

	g.p("}")
	g.p()

	for group, _ := range groups {
		for _, path := range paths {
			operations := path.PathItemProps
			if post := operations.Post; post != nil {
				g.generateHandler(group, post)
			}
			if get := operations.Get; get != nil {
				g.generateHandler(group, get)
			}
			if put := operations.Put; put != nil {
				g.generateHandler(group, put)
			}
			if del := operations.Delete; del != nil {
				g.generateHandler(group, del)
			}
		}
	}
}

func (g *Generator) generateRouter(method, group, path string, op *spec.Operation) {
	routeGroup := op.OperationProps.Tags[0]
	if routeGroup != group {
		return
	}

	routePath := strings.TrimPrefix(path, "/"+routeGroup)
	routePath = strings.Replace(routePath, "{", ":", -1)
	routePath = strings.Replace(routePath, "}", "", -1)
	g.p(g.caps(routeGroup), ".", method, "(\"", routePath, "\", ", g.caps(op.OperationProps.ID), "Handler)")
}

func (g *Generator) generateHandler(group string, op *spec.Operation) {
	if op.OperationProps.Tags[0] != group {
		return
	}

	var hasBodyParam, hasQueryParam bool
	parameters := op.OperationProps.Parameters
	responses := op.Responses.ResponsesProps.StatusCodeResponses
	opParams := ""
	modelResp := ""

	for status, resp := range responses {
		if status == 200 {
			if refUrl := resp.Schema.SchemaProps.Ref.Ref.ReferenceURL; refUrl != nil {
				modelResp = strings.TrimPrefix(refUrl.Fragment, "/definitions/")
				break
			}
		}
	}

	for _, param := range parameters {
		pp := param.ParamProps
		if pp.In == "body" {
			hasBodyParam = true
		} else if pp.In == "query" {
			hasQueryParam = true
		}
	}

	g.p("func ", g.caps(op.OperationProps.ID), "Handler(c *gin.Context) {")

	if hasQueryParam {
		g.p("queryValues := c.Request.URL.Query()")
		g.p()
	}

	for _, param := range parameters {
		pp := param.ParamProps
		if pp.In == "body" {
			ref := pp.Schema.SchemaProps.Ref.Ref.ReferenceURL.Fragment
			g.p("var ", pp.Name, " models.", strings.TrimPrefix(ref, "/definitions/"))
			g.p()
			g.p("if err := c.BindJSON(&body); err != nil {")
			g.p("		return")
			g.p("	}")
			g.p()
			g.p("if err := body.Validate(); err != nil {")
			g.p("	c.JSON(http.StatusBadRequest, err)")
			g.p("	return")
			g.p("}")
			g.p()

		} else if pp.In == "query" {
			if param.SimpleSchema.Type == "string" {
				g.p(pp.Name, " := queryValues.Get(\"", pp.Name, "\")")
				if pp.Required {
					g.p("if ", pp.Name, " == \"\" {")
					g.p("	c.JSON(http.StatusBadRequest, gin.H{\"missing\": \"", pp.Name, "\"})")
					g.p("	return")
					g.p("}")
				}
				g.p()
			} else {
				strName := "str" + g.caps(pp.Name)
				g.p(strName, " := queryValues.Get(\"", pp.Name, "\")")
				if pp.Required {
					g.p("if ", strName, " == \"\" {")
					g.p("	c.JSON(http.StatusBadRequest, gin.H{\"missing\": \"", pp.Name, "\"})")
					g.p("	return")
					g.p("}")
				}
				g.p()
				g.generateParamInt(strName, pp.Name, param.SimpleSchema.Format)
			}
			opParams += pp.Name + ", "

		} else if pp.In == "formData" {
			if param.SimpleSchema.Type == "string" {
				g.p(pp.Name, " := c.Request.PostFormValue(\"", pp.Name, "\")")
				if pp.Required {
					g.p("if ", pp.Name, " == \"\" {")
					g.p("	c.JSON(http.StatusBadRequest, gin.H{\"missing\": \"", pp.Name, "\"})")
					g.p("	return")
					g.p("}")
				}
				g.p()
			} else {
				strName := "str" + g.caps(pp.Name)
				g.p(strName, " := c.Request.PostFormValue(\"", pp.Name, "\")")
				if pp.Required {
					g.p("if ", strName, " == \"\" {")
					g.p("	c.JSON(http.StatusBadRequest, gin.H{\"missing\": \"", pp.Name, "\"})")
					g.p("	return")
					g.p("}")
				}
				g.p()
				g.generateParamInt(strName, pp.Name, param.SimpleSchema.Format)
			}
			opParams += pp.Name + ", "

		} else if pp.In == "path" {
			if param.SimpleSchema.Type == "string" {
				g.p(pp.Name, " := c.Param(\"", pp.Name, "\")")
				g.p()
			} else {
				strName := "str" + g.caps(pp.Name)
				g.p(strName, " := c.Param(\"", pp.Name, "\")")
				g.generateParamInt(strName, pp.Name, param.SimpleSchema.Format)
			}
			opParams += pp.Name + ", "
		}
	}

	if hasBodyParam {
		opParams += "&body"
	}

	if modelResp != "" {
		g.p("if resp, err := operations.", g.caps(op.OperationProps.ID), "(", strings.TrimSuffix(opParams, ", "), "); err == nil {")
		g.p("	c.JSON(http.StatusOK, resp)")
	} else {
		g.p("if err := operations.", g.caps(op.OperationProps.ID), "(", strings.TrimSuffix(opParams, ", "), "); err == nil {")
		g.p("	c.String(http.StatusOK, \"Success\")")
	}
	g.p("} else {")
	g.p("	c.JSON(http.StatusBadRequest, err)")
	g.p("}")
	g.p("}")
	g.p()
}

func (g *Generator) generateParamInt(strName, name, format string) {
	g.p("var ", name, " ", format)
	g.p("if i, err := strconv.ParseInt(", strName, ", 10, ", strings.TrimPrefix(format, "int"), "); err != nil {")
	g.p("	c.JSON(http.StatusBadRequest, gin.H{\"invalid\": \"", name, "\"})")
	g.p("	return")
	g.p("} else {")
	g.p(name, " = ", format, "(i)")
	g.p("}")
	g.p()
}

func (g *Generator) generateOperations(buf *bytes.Buffer, specDoc *spec.Document) {
	g.Buffer = buf
	paths := specDoc.AllPaths()

	g.p("package operations")
	g.p()

	for _, path := range paths {
		operations := path.PathItemProps
		if post := operations.Post; post != nil {
			g.generateOperation(post)
		}
		if get := operations.Get; get != nil {
			g.generateOperation(get)
		}
		if put := operations.Put; put != nil {
			g.generateOperation(put)
		}
		if del := operations.Delete; del != nil {
			g.generateOperation(del)
		}
	}
}

func (g *Generator) generateOperation(op *spec.Operation) {
	var hasBodyParam bool
	parameters := op.OperationProps.Parameters
	responses := op.Responses.ResponsesProps.StatusCodeResponses
	model := ""
	modelResp := ""
	opParams := ""

	for status, resp := range responses {
		if status == 200 {
			if refUrl := resp.Schema.SchemaProps.Ref.Ref.ReferenceURL; refUrl != nil {
				modelResp = strings.TrimPrefix(refUrl.Fragment, "/definitions/")
				break
			}
		}
	}

	for _, param := range parameters {
		pp := param.ParamProps
		if pp.In == "body" {
			hasBodyParam = true
		}
	}

	for _, param := range parameters {
		pp := param.ParamProps
		if pp.In == "body" {
			ref := pp.Schema.SchemaProps.Ref.Ref.ReferenceURL.Fragment
			model = strings.TrimPrefix(ref, "/definitions/")

		} else if pp.In == "query" {
			if param.SimpleSchema.Type == "string" {
				opParams += pp.Name + " string, "
			} else {
				opParams += pp.Name + " " + param.SimpleSchema.Format + ", "
			}

		} else if pp.In == "formData" {
			if param.SimpleSchema.Type == "string" {
				opParams += pp.Name + " string, "
			} else {
				opParams += pp.Name + " " + param.SimpleSchema.Format + ", "
			}

		} else if pp.In == "path" {
			if param.SimpleSchema.Type == "string" {
				opParams += pp.Name + " string, "
			} else {
				opParams += pp.Name + " " + param.SimpleSchema.Format + ", "
			}

		}
	}

	if hasBodyParam {
		opParams += g.lowerFirst(model) + " *models." + model
	}

	if modelResp != "" {
		g.p("func ", g.caps(op.OperationProps.ID), "(", strings.TrimSuffix(opParams, ", "), ") (* models.", modelResp, ", error) {")
		g.p("	return &models.", modelResp, "{}, nil")
	} else {
		g.p("func ", g.caps(op.OperationProps.ID), "(", strings.TrimSuffix(opParams, ", "), ") (error) {")
		g.p("	return nil")
	}
	g.p("}")
	g.p()
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

func (g *Generator) generateStruct(def *GenDefinition) {
	g.p("type ", def.GenSchema.Name, " struct {")
	for _, prop := range def.Properties {
		if g.hasExtendFormat(&prop) {
			prop.resolvedType.GoType = "string"
		} else if prop.resolvedType.SwaggerFormat == "date-time" {
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
		g.p("	if err:= validate.Enum(path, location, value, ", varEnum, "); err != nil {")
		g.p("		return err")
		g.p("	}")
		g.p()
		g.p("	return nil")
		g.p("}")
		g.p()
	}

	g.p("func (m *", model, ") validate", propName, "() error {")

	if prop.sharedValidations.Required == false {
		if prop.resolvedType.GoType == "string" {
			g.p("if m.", propName, " == \"\" {")
			g.p("	return nil")
			g.p("}")
			g.p()
		} else if strings.HasPrefix(prop.resolvedType.GoType, "int") {
			g.p("if m.", propName, " == 0 {")
			g.p("	return nil")
			g.p("}")
			g.p()
		}
	}
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
		g.p("if err := m.validate", propName, "Enum(\"", prop.Name, "\", \"body\", ", "m.", propName, "); err != nil {")
		g.p("	return err")
		g.p("}")
		g.p()
	}
	if g.hasExtendFormat(prop) {
		if prop.sharedValidations.Required == false {
			g.p("if m.", propName, " == \"\" {")
			g.p("	return nil")
			g.p("}")
			g.p()
		}
		validatefunc, _ := govalidator.TagMap[prop.resolvedType.SwaggerFormat]
		funcName := runtime.FuncForPC(reflect.ValueOf(validatefunc).Pointer()).Name()
		g.p("if ", funcName[22:], "(m.", propName, ") != true {")
		g.p("	return errors.InvalidType(\"", prop.Name, "\",\"body\", \"", prop.resolvedType.SwaggerFormat, "\", m.", propName, ")")
		g.p("}")
		g.p()
	}
	g.p("	return nil")
	g.p("}")
	g.p()
}
