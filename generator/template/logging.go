package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	loggerVarName            = "logger"
	nextVarName              = "next"
	serviceLoggingStructName = "serviceLogging"

	logIgnoreTag = "logs-ignore"
	lenTag       = "logs-len"
)

type loggingTemplate struct {
	Info         *GenerationInfo
	ignoreParams map[string][]string
	lenParams    map[string][]string
}

func NewLoggingTemplate(info *GenerationInfo) Template {
	return &loggingTemplate{
		Info: info,
	}
}

// Render all logging.go file.
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package middleware
//
//		import (
//			context "context"
//			svc "github.com/devimteam/microgen/example/svc"
//			log "github.com/go-kit/kit/log"
//			time "time"
//		)
//
//		func ServiceLogging(logger log.Logger) Middleware {
//			return func(next svc.StringService) svc.StringService {
//				return &serviceLogging{
//					logger: logger,
//					next:   next,
//				}
//			}
//		}
//
//		type serviceLogging struct {
//			logger log.Logger
//			next   svc.StringService
//		}
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text,
// 					"symbol", symbol,
//					"count", count,
// 					"positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t *loggingTemplate) Render() write_strategy.Renderer {
	f := NewFile("middleware")
	f.PackageComment(t.Info.FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Comment("ServiceLogging writes params, results and working time of method call to provided logger after its execution.").
		Line().Func().Id(util.ToUpperFirst(serviceLoggingStructName)).Params(Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newLoggingBody(t.Info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceLoggingStructName).Struct(
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.Info.ServiceImportPath, t.Info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.Info.Iface.Methods {
		f.Line()
		f.Add(t.loggingFunc(signature)).Line()
	}

	for _, signature := range t.Info.Iface.Methods {
		if params := removeContextIfFirst(signature.Args); t.calcParamAmount(signature.Name, params) > 0 {
			f.Add(t.loggingEntity("log"+requestStructName(signature), signature, params)).Line()
		}
		if params := removeErrorIfLast(signature.Results); t.calcParamAmount(signature.Name, params) > 0 {
			f.Add(t.loggingEntity("log"+responseStructName(signature), signature, params)).Line()
		}
	}

	return f
}

func (loggingTemplate) DefaultPath() string {
	return "./middleware/logging.go"
}

func (t *loggingTemplate) Prepare() error {
	t.ignoreParams = make(map[string][]string)
	t.lenParams = make(map[string][]string)
	for _, fn := range t.Info.Iface.Methods {
		t.ignoreParams[fn.Name] = util.FetchTags(fn.Docs, TagMark+logIgnoreTag)
		t.lenParams[fn.Name] = util.FetchTags(fn.Docs, TagMark+lenTag)
	}
	return nil
}

func (t *loggingTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

// Render body for new logging middleware.
//
//		return func(next svc.StringService) svc.StringService {
//			return &serviceLogging{
//				logger: logger,
//				next:   next,
//			}
//		}
//
func (t *loggingTemplate) newLoggingBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.Info.ServiceImportPath, i.Name),
	).Params(
		Qual(t.Info.ServiceImportPath, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceLoggingStructName).Values(
			Dict{
				Id(loggerVarName): Id(loggerVarName),
				Id(nextVarName):   Id(nextVarName),
			},
		))
	}))
}

func (t *loggingTemplate) loggingEntity(name string, fn *types.Function, params []types.Variable) Code {
	if len(params) == 0 {
		return nil
	}
	return Type().Id(name).StructFunc(func(g *Group) {
		ignore := t.ignoreParams[fn.Name]
		lenParams := t.lenParams[fn.Name]
		for _, field := range params {
			if !util.IsInStringSlice(field.Name, ignore) {
				g.Id(util.ToUpperFirst(field.Name)).Add(fieldType(field.Type, false))
			}
			if util.IsInStringSlice(field.Name, lenParams) {
				g.Id("Len" + util.ToUpperFirst(field.Name)).Int().Tag(map[string]string{"json": "len(" + util.ToUpperFirst(field.Name) + ")"})
			}
		}
	}).Line()
}

// Render logging middleware for interface method.
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text, "symbol", symbol,
//					"count", count, "positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t *loggingTemplate) loggingFunc(signature *types.Function) *Statement {
	return methodDefinition(serviceLoggingStructName, signature).
		BlockFunc(t.loggingFuncBody(signature))
}

// Render logging function body with request/response and time tracking.
//
//		defer func(begin time.Time) {
//			s.logger.Log(
//				"method", "Count",
//				"text", text, "symbol", symbol,
//				"count", count, "positions", positions,
//				"took", time.Since(begin))
//		}(time.Now())
//		return s.next.Count(ctx, text, symbol)
//
func (t *loggingTemplate) loggingFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params(Id("begin").Qual(PackagePathTime, "Time")).Block(
			Id(util.LastUpperOrFirst(serviceLoggingStructName)).Dot(loggerVarName).Dot("Log").CallFunc(func(g *Group) {
				g.Line().Lit("method")
				g.Lit(signature.Name)

				if t.calcParamAmount(signature.Name, removeContextIfFirst(signature.Args)) > 0 {
					g.Line().List(Lit("request"), t.logRequest(signature))
				}
				if t.calcParamAmount(signature.Name, removeErrorIfLast(signature.Results)) > 0 {
					g.Line().List(Lit("response"), t.logResponse(signature))
				}
				if !util.IsInStringSlice(nameOfLastResultError(signature), t.ignoreParams[signature.Name]) {
					g.Line().List(Lit(nameOfLastResultError(signature)), Id(nameOfLastResultError(signature)))
				}

				g.Line().Lit("took")
				g.Qual(PackagePathTime, "Since").Call(Id("begin"))
			}),
		).Call(Qual(PackagePathTime, "Now").Call())

		g.Return().Id(util.LastUpperOrFirst(serviceLoggingStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err,
// 		"result", result,
//		"count", count,
//
func (t *loggingTemplate) paramsNameAndValue(fields []types.Variable, functionName string) *Statement {
	return ListFunc(func(g *Group) {
		ignore := t.ignoreParams[functionName]
		lenParams := t.lenParams[functionName]
		for _, field := range fields {
			if !util.IsInStringSlice(field.Name, ignore) {
				g.Line().List(Lit(field.Name), Id(field.Name))
			}
			if util.IsInStringSlice(field.Name, lenParams) {
				g.Line().List(Lit("len("+field.Name+")"), Len(Id(field.Name)))
			}
		}
	})
}

func (t *loggingTemplate) fillMap(fn *types.Function, params []types.Variable) *Statement {
	return Values(DictFunc(func(d Dict) {
		ignore := t.ignoreParams[fn.Name]
		lenParams := t.lenParams[fn.Name]
		for _, field := range params {
			if !util.IsInStringSlice(field.Name, ignore) {
				d[Id(util.ToUpperFirst(field.Name))] = Id(field.Name)
			}
			if util.IsInStringSlice(field.Name, lenParams) {
				d[Id("Len"+util.ToUpperFirst(field.Name))] = Len(Id(field.Name))
			}
		}
	}))
}

func (t *loggingTemplate) logRequest(fn *types.Function) *Statement {
	paramAmount := t.calcParamAmount(fn.Name, removeContextIfFirst(fn.Args))
	if paramAmount <= 0 {
		return Lit("")
	}
	return Id("log" + requestStructName(fn)).Add(t.fillMap(fn, removeContextIfFirst(fn.Args)))
}

func (t *loggingTemplate) logResponse(fn *types.Function) *Statement {
	paramAmount := t.calcParamAmount(fn.Name, removeErrorIfLast(fn.Results))
	if paramAmount <= 0 {
		return Lit("")
	}
	return Id("log" + responseStructName(fn)).Add(t.fillMap(fn, removeErrorIfLast(fn.Results)))
}

func (t *loggingTemplate) calcParamAmount(name string, params []types.Variable) int {
	ignore := t.ignoreParams[name]
	lenParams := t.lenParams[name]
	paramAmount := len(params)
	for _, field := range params {
		if util.IsInStringSlice(field.Name, ignore) {
			paramAmount -= 1
		}
		if util.IsInStringSlice(field.Name, lenParams) {
			paramAmount += 1
		}
	}
	return paramAmount
}
