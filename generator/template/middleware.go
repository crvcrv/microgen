package template

import (
	"github.com/devimteam/microgen/generator/write_strategy"
	. "github.com/vetcher/jennifer/jen"
)

const (
	MiddlewareTypeName = "Middleware"
)

type middlewareTemplate struct {
	Info *GenerationInfo
}

func NewMiddlewareTemplate(info *GenerationInfo) Template {
	return &middlewareTemplate{
		Info: info.Duplicate(),
	}
}

// Render middleware decorator
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package middleware
//
//		import svc "github.com/devimteam/microgen/test/svc"
//
//		type Middleware func(svc.StringService) svc.StringService
//
func (t *middlewareTemplate) Render() write_strategy.Renderer {
	f := NewFile(t.Info.ServiceImportPackageName)
	f.PackageComment(FileHeader)
	f.PackageComment(`Please, do not edit.`)
	f.Type().Id(MiddlewareTypeName).Func().Call(Qual(t.Info.ServiceImportPath, t.Info.Iface.Name)).Qual(t.Info.ServiceImportPath, t.Info.Iface.Name)
	return f
}

func (middlewareTemplate) DefaultPath() string {
	return "./middleware/middleware.go"
}

func (middlewareTemplate) Prepare() error {
	return nil
}

func (t *middlewareTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewFileMethod(t.Info.AbsOutPath, t.DefaultPath()), nil
}
