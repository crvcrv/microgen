package template

import (
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

const (
	GolangProtobufPtypesTimestamp = "github.com/golang/protobuf/ptypes/timestamp"
	JsonbPackage                  = "github.com/sas1024/gorm-jsonb/jsonb"
)

func specialTypeConverter(p types.Type) *Statement {
	// error -> string
	if p.Name == "error" && p.Import == nil {
		return (&Statement{}).Id("string")
	}
	// time.Time -> timestamp.Timestamp
	if p.Name == "Time" && p.Import != nil && p.Import.Package == "time" {
		return (&Statement{}).Qual(GolangProtobufPtypesTimestamp, "Timestamp")
	}
	// jsonb.JSONB -> string
	if p.Name == "JSONB" && p.Import != nil && p.Import.Package == JsonbPackage {
		return (&Statement{}).Id("string")
	}
	return nil
}

type StubGRPCTypeConverterTemplate struct {
	PackagePath               string
	ServicePackageName        string
	alreadyRenderedConverters []string
	packageName               string
}

// Render whole file with protobuf converters.
//
//		// This file was automatically generated by "microgen" utility.
//		package protobuf
//
//		func IntListToProto(positions []int) (protoPositions []int64, convPositionsErr error) {
//			panic("method not provided")
//		}
//
//		func ProtoToIntList(protoPositions []int64) (positions []int, convPositionsErr error) {
//			panic("method not provided")
//		}
//
func (t *StubGRPCTypeConverterTemplate) Render(i *types.Interface) *Statement {
	t.packageName = "protobuf"
	f := Statement{}

	for _, signature := range i.Methods {
		args := append(removeContextIfFirst(signature.Args), removeContextIfFirst(signature.Results)...)
		for _, field := range args {
			if _, ok := golangTypeToProto("", &field); !ok && !util.IsInStringSlice(typeToProto(&field.Type), t.alreadyRenderedConverters) {
				f.Add(t.stubConverterToProto(&field)).Line().Line()
				t.alreadyRenderedConverters = append(t.alreadyRenderedConverters, typeToProto(&field.Type))
				f.Add(t.stubConverterProtoTo(&field)).Line().Line()
				t.alreadyRenderedConverters = append(t.alreadyRenderedConverters, protoToType(&field.Type))
			}
		}
	}

	return &f
}

func (StubGRPCTypeConverterTemplate) Path() string {
	return "./transport/converter/protobuf/type_converters.go"
}

func (t *StubGRPCTypeConverterTemplate) PackageName() string {
	return t.packageName
}

// Render stub method for golang to protobuf converter.
//
//		func IntListToProto(positions []int) (protoPositions []int64, convPositionsErr error) {
//			return
//		}
//
func (t *StubGRPCTypeConverterTemplate) stubConverterToProto(field *types.Variable) *Statement {
	return Func().Id(typeToProto(&field.Type)).
		Params(Id(util.ToLowerFirst(field.Name)).Add(fieldType(&field.Type))).
		Params(Id("proto"+util.ToUpperFirst(field.Name)).Add(t.protoFieldType(field)), Id("conv"+util.ToUpperFirst(field.Name)+"Err").Error()).
		Block(
			Panic(Lit("method not provided")),
		)
}

// Render stub method for protobuf to golang converter.
//
//		func ProtoToIntList(protoPositions []int64) (positions []int, convPositionsErr error) {
//			return
//		}
//
func (t *StubGRPCTypeConverterTemplate) stubConverterProtoTo(field *types.Variable) *Statement {
	return Func().Id(protoToType(&field.Type)).
		Params(Id("proto"+util.ToUpperFirst(field.Name)).Add(t.protoFieldType(field))).
		Params(Id(util.ToLowerFirst(field.Name)).Add(fieldType(&field.Type)), Id("conv"+util.ToUpperFirst(field.Name)+"Err").Error()).
		Block(
			Panic(Lit("method not provided")),
		)
}

// Render protobuf field type for given func field.
//
//  	*repository.Visit
//
func (t *StubGRPCTypeConverterTemplate) protoFieldType(field *types.Variable) *Statement {
	c := &Statement{}

	if field.Type.IsArray {
		c.Index()
	}

	if field.Type.IsPointer {
		c.Op("*")
	}

	protoType := field.Type.Name
	if tmp, ok := goToProtoTypesMap[field.Type.Name]; ok {
		protoType = tmp
	}
	if code := specialTypeConverter(field.Type); code != nil {
		return c.Add(code)
	}
	if field.Type.Import != nil {
		c.Qual(protobufPath(t.ServicePackageName), protoType)
	} else {
		c.Id(protoType)
	}

	return c
}
