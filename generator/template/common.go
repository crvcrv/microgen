package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const Context = "ctx"

func structFieldName(field *parser.FuncField) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Renders struct field.
//
//  	Visit *entity.Visit `json:"visit"`
//
func structField(field *parser.FuncField) *Statement {
	s := structFieldName(field)
	s.Add(fieldType(field))
	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})
	return s
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []*parser.FuncField) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Id(util.ToLowerFirst(field.Name)).Add(fieldType(field))
		}
	})
	return c
}

// Renders field type for given func field.
//
//  	*repository.Visit
//
func fieldType(field *parser.FuncField) *Statement {
	c := &Statement{}

	if field.IsArray {
		c.Index()
	}

	if field.IsPointer {
		c.Op("*")
	}

	if field.Package != nil {
		c.Qual(field.Package.Path, field.Type)
	} else {
		c.Id(field.Type)
	}

	return c
}

// Renders key/value pairs wrapped in Dict for provided fields
//
//		Err:    err,
//		Result: result,
//
func mapInitByFuncFields(fields []*parser.FuncField) Dict {
	return DictFunc(func(d Dict) {
		for i, field := range fields {
			if field.Package != nil && field.Package.Path == PackageAliasContext && i == 0 {
				continue
			}
			d[structFieldName(field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Renders func params for call.
//
//  	req.Visit, req.Err
//
func funcCallParams(obj string, fields []*parser.FuncField) *Statement {
	var list []Code
	for i, field := range fields {
		if field.Package != nil && field.Package.Path == PackageAliasContext && i == 0 {
			continue
		}
		list = append(list, Id(obj).Dot(util.ToUpperFirst(field.Name)))
	}
	return List(list...)
}

func funcResivers(fields []*parser.FuncField) *Statement {
	var list []Code
	for _, field := range fields {
		list = append(list, Id(util.ToLowerFirst(field.Name)))
	}
	return List(list...)
}

func withCtx(param Code) *Statement {
	return List(Id(Context), param)
}

func fullServiceMethodCall(service, request string, signature *parser.FuncSignature) *Statement {
	return funcResivers(signature.Results).Op(":=").Id(service).Dot(signature.Name).Call(withCtx(funcCallParams(request, signature.Params)))
}

func typeCasting(resiever, iface, to string) *Statement {
	if resiever == "" {
		return Id(iface).Assert(Op("*").Id(to))
	}
	return Id(resiever).Op(":=").Id(iface).Assert(Op("*").Id(to))
}

func methodDefinition(obj string, signature *parser.FuncSignature) *Statement {
	return Func().
		Params(Id(util.FirstLowerChar(obj)).Op("*").Id(obj)).
		Id(signature.Name).
		Params(funcDefinitionParams(signature.Params)).
		Params(funcDefinitionParams(signature.Results))
}
