package microgen

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"plugin"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/devimteam/microgen/generator"
	lg "github.com/devimteam/microgen/logger"
)

var (
	flagDstDir  = flag.String("dst", ".", "Destiny path.")
	flagVerbose = flag.Int("v", common, "Sets microgen verbose level.")
	flagDebug   = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagConfig  = flag.String("cfg", "microgen.toml", "")
	flagDry     = flag.Bool("dry", false, "Do everything except writing files.")
)

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

func Exec() {
	var err error
	defer func() {
		if err != nil {
			lg.Logger.Logln(critical, "fatal:", err)
			os.Exit(1)
		}
		if err := recover(); err != nil {
			lg.Logger.Logln(critical, "panic:", err)
			os.Exit(1)
		}
	}()
	begin := time.Now()
	defer func() {
		lg.Logger.Logln(info, "Done")
		lg.Logger.Logln(info, "Duration:", time.Since(begin))
	}()
	if *flagVerbose < critical {
		*flagVerbose = critical
	}
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = debug
	}
	lg.Logger.Logln(common, "microgen", "1.0.0")

	lg.Logger.Logln(detail, "Config:", *flagConfig)
	cfg, err := processConfig(*flagConfig)
	if err != nil {
		return
	}

	pkg, err := astra.GetPackage(".", astra.AllowAnyImportAliases,
		astra.IgnoreStructs, astra.IgnoreFunctions, astra.IgnoreConstants,
		astra.IgnoreMethods, astra.IgnoreTypes, astra.IgnoreVariables,
	)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	ii := findInterfaces(pkg)
	iface, err := selectInterface(ii, cfg.Interface)
	if err != nil {
		lg.Logger.Logln(detail, "All founded interfaces:")
		lg.Logger.Logln(detail, listInterfaces(pkg.Interfaces))
		return
	}

	err = initPlugins(cfg.Plugins)
	if err != nil {
		return
	}

	source, err := os.Getwd()
	if err != nil {
		return
	}
	sourcePackage, err := getPkgPath(".", true)
	if err != nil {
		return
	}

	lg.Logger.Logln(debug, "Start generation")
	ctx := Context{
		Interface:           iface,
		Source:              source,
		SourcePackageImport: sourcePackage,
		Files:               nil,
	}
	lg.Logger.Logln(debug, "Exec plugins")
	for _, pcfg := range cfg.Generate {
		err = func() error {
			defer func() {
				if err := recover(); err != nil {
					err = errors.Errorf("recover panic from %s plugin. Message: %v", pcfg.Name, err)
				}
			}()
			p, ok := pluginsRepository[pcfg.Name]
			if !ok {
				return errors.Errorf("plugin %s not registered")
			}
			lg.Logger.Logln(debug, "run", pcfg.Name, "plugin with args:", pcfg.Args)
			ctx, err = p.Generate(ctx)
			if err != nil {
				return errors.Wrapf(err, "%s plugin returns an error", pcfg.Name)
			}
			return nil
		}()
		if err != nil {
			return
		}
	}
	if *flagDry {
		lg.Logger.Logln(debug, "dry execution: do not create files")
		return
	}
	lg.Logger.Logln(debug, "Write files")
	for _, f := range ctx.Files {
		lg.Logger.Logln(debug, "create", f.Path)
		tgtFile, err := os.Create(f.Path)
		if err != nil {
			err = errors.Wrapf(err, "plugin %s: during creating %s file", f.Name, f.Path)
			return
		}
		lg.Logger.Logln(debug, "write", f.Path)
		_, err = tgtFile.Write(f.Content)
		if err != nil {
			err = errors.Wrapf(err, "plugin %s: during creating %s file", f.Name, f.Path)
			return
		}
	}
}

func findInterfaces(file *types.File) []*types.Interface {
	var ifaces []*types.Interface
	for i := range file.Interfaces {
		if docsContainMicrogenTag(file.Interfaces[i].Docs) {
			ifaces = append(ifaces, &file.Interfaces[i])
		}
	}
	return ifaces
}

func listInterfaces(ii []types.Interface) string {
	var s string
	for _, i := range ii {
		s = s + fmt.Sprintf("\t%s(%d methods, %d embedded interfaces)\n", i.Name, len(i.Methods), len(i.Interfaces))
	}
	return s
}

func selectInterface(ii []*types.Interface, name string) (*types.Interface, error) {
	if len(ii) == 0 {
		return ii[0], nil
	}
	if name == "" {
		return nil, fmt.Errorf("%d interfaces founded, but 'interface' config parameter is empty. Add \"interface = InterfaceName\" to config file", len(ii))
	}
	for i := range ii {
		if ii[i].Name == name {
			return ii[i], nil
		}
	}
	return nil, fmt.Errorf("%s interface not found, but %d others are available", name, len(ii))
}

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, generator.TagMark+generator.MicrogenMainTag) {
			return true
		}
	}
	return false
}

func processConfig(pathToConfig string) (*config, error) {
	file, err := os.Open(pathToConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "open file")
	}
	var rawToml bytes.Buffer
	_, err = rawToml.ReadFrom(file)
	if err != nil {
		return nil, errors.WithMessage(err, "read from config")
	}
	var cfg config
	err = toml.NewDecoder(&rawToml).Decode(&cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}
	return &cfg, nil
}

func initPlugins(plugins []string) error {
	for i := range plugins {
		_, err := plugin.Open(plugins[i])
		if err != nil {
			return errors.Wrapf(err, "open plugin %s", plugins[i])
		}
	}
	return nil
}

const (
	critical = iota
	common
	info
	detail
	debug = 100
)
