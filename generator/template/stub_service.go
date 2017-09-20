package template

import (
	"os"
	"path/filepath"

	"fmt"

	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

type stubInterfaceTemplate struct {
	Info *GenerationInfo

	alreadyRenderedMethods []string
	isStructExist          bool
	isConstructorExist     bool
}

func NewStubInterfaceTemplate(info *GenerationInfo) Template {
	infoCopy := info.Duplicate()
	return &stubInterfaceTemplate{
		Info: infoCopy,
	}
}

// Renders exchanges file.
//
//  package visitsvc
//
//  import (
//  	"gitlab.devim.team/microservices/visitsvc/entity"
//  )
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
//  type CreateVisitResponse struct {
//  	Res *entity.Visit `json:"res"`
//  	Err error         `json:"err"`
//  }
//
func (t *stubInterfaceTemplate) Render() write_strategy.Renderer {
	f := &Statement{}

	if !t.isStructExist {
		f.Comment(`Generated by "microgen" tool.`).Line().
			Commentf(`Structure %s implements %s interface.`, util.ToLowerFirst(t.Info.Iface.Name), t.Info.Iface.Name).Line().
			Type().Id(util.ToLowerFirst(t.Info.Iface.Name)).Struct(Line()).Line()
	}

	if !t.isConstructorExist {
		f.Func().Id(constructorName(t.Info.Iface)).Params().Id(t.Info.Iface.Name).Block(
			Panic(Lit("constructor not provided")),
		).Line()
	}

	for _, signature := range t.Info.Iface.Methods {
		if !util.IsInStringSlice(signature.Name, t.alreadyRenderedMethods) {
			f.Line().Add(methodDefinition(util.ToLowerFirst(t.Info.Iface.Name), signature)).Block(
				Panic(Lit("method not provided")),
			).Line()
		}
	}
	return f
}

func (stubInterfaceTemplate) DefaultPath() string {
	return "."
}

func (t *stubInterfaceTemplate) Prepare() error {
	if err := util.TryToOpenFile(t.Info.SourceFilePath, t.DefaultPath()); os.IsNotExist(err) {
		fmt.Println(err)
		return nil
	}
	file, err := util.ParseFile(filepath.Join(t.Info.SourceFilePath, t.DefaultPath()))
	if err != nil {
		return err
	}

	for i := range file.Methods {
		if file.Methods[i].Receiver.Type.Name == util.ToLowerFirst(t.Info.Iface.Name) && file.Methods[i].Receiver.Type.Import == nil {
			t.alreadyRenderedMethods = append(t.alreadyRenderedMethods, file.Methods[i].Name)
		}
	}

	for i := range file.Structures {
		if file.Structures[i].Name == util.ToLowerFirst(t.Info.Iface.Name) {
			t.isStructExist = true
			break
		}
	}

	for i := range file.Functions {
		if file.Functions[i].Name == constructorName(t.Info.Iface) {
			t.isConstructorExist = true
			break
		}
	}

	return nil
}

func (t *stubInterfaceTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewAppendToFileStrategy(t.Info.SourceFilePath, t.DefaultPath()), nil
}

func constructorName(p *types.Interface) string {
	return "New" + p.Name
}
