package main

import (
	"fmt"
	"github.com/vc2402/vivard/gen/js"
	"github.com/vc2402/vivard/gen/vue"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vc2402/vivard/gen"
	_ "github.com/vc2402/vivard/gen/js"
	_ "github.com/vc2402/vivard/gen/vue"
)

func main() {
	pflag.String("package", "test", "default package name")
	pflag.String("in", ".", "Input directory")
	pflag.String("out", ".", "Output directory")
	pflag.String("clientOut", "/home/victor/work/vivasoft/vue/gen-test/src/generated", "Output directory for client files")
	pflag.Bool("print", false, "Print result to stdout")
	pflag.String("cfgPath", ".", "Path to config file")
	pflag.String("cfg", ".vivgen", "Config file name")
	pflag.String("pkgPrefix", "", "Package prefix")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetConfigName(viper.GetString("cfg"))
	viper.AddConfigPath(viper.GetString("."))
	viper.AddConfigPath(viper.GetString("in"))
	viper.AddConfigPath(viper.GetString("cfgPath"))

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("warning: config file not found: %v\n", err)
		} else {
			fmt.Printf("error reading config file: %v\n", err)
			return
		}
	}
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Println("no files to parse given")
		return
	}
	in := viper.GetString("in")
	files := make([]string, 0, len(args))
	if in == "" {
		in = "."
	}
	for _, fn := range args {
		var name string
		if !filepath.IsAbs(fn) {
			name = filepath.Join(in, fn)
		} else {
			name = fn
		}
		matches, err := filepath.Glob(name)
		if err != nil {
			fmt.Printf("invalid input file name/mask: %s\n", fn)
			return
		}
		files = append(files, matches...)
	}
	res, err := gen.Parse(files)
	if err != nil {
		fmt.Printf("errors found: %v\n", err)
	} else {
		// fmt.Printf("parsed: %#+v: \n", res)
		opts := gen.Options(viper.GetString("out")).
			With(gen.UnknownAnnotationWarning).
			With(gen.NullablePointers)

		if options := viper.Get("options"); options != nil {
			err = opts.FromAny(options)
			if err != nil {
				fmt.Printf("config file error: invalid option: %v\n", err)
				return
			}
		}
		if viper.GetString("pkgPrefix") != "" {
			opts.With(gen.PackagePrefixOption(viper.GetString("pkgPrefix")))
		}
		if viper.GetString("clientOut") != "" {
			opts.SetClientOutputDir(viper.GetString("clientOut"))
		}
		var proj *gen.Project
		if pls := viper.Get("plugins"); pls != nil {
			plugins, ok := pls.([]interface{})
			if !ok {
				fmt.Println("config file error: invalid 'plugins' value (should be array)")
				return
			}
			proj = gen.New(res, opts)
			for _, pl := range plugins {
				if name, ok := pl.(string); ok {
					err = proj.WithPlugin(name, nil)
					if err != nil {
						fmt.Printf("error while creating plugin: %v\n", err)
						return
					}
				} else if plugin, ok := pl.(map[string]any); ok {
					opts := map[string]any{}
					addOptions := func(options map[string]any) {
						for o, v := range options {
							opts[o] = v
						}
					}
					var name string
					for k, v := range plugin {
						switch k {
						case "name":
							name, ok = v.(string)
							if !ok {
								fmt.Printf("config file error: plugin name should be a string: %v\n", v)
								return
							}
						case "options":
							if op, ok := v.(map[string]any); ok {
								addOptions(op)
							} else if ops, ok := v.([]map[string]any); ok {
								for _, op := range ops {
									addOptions(op)
								}
							} else {
								fmt.Printf("config file error: plugin options should be amap or an array of maps: %v\n", v)
								return
							}

						default:
							opts[k] = v
						}
					}
					if name == "" {
						fmt.Printf("config file error: there is no name for plugin: %+v", pl)
						return
					}
					if len(opts) == 0 {
						opts = nil
					}
					err = proj.WithPlugin(name, opts)
					if err != nil {
						fmt.Printf("error while creating plugin: %v", err)
						return
					}
				} else {
					fmt.Printf("config file error: invalid plugin descriptor: %+v", pl)
					return
				}
			}
		} else {
			proj = gen.New(
				res,
				opts.
					WithCustom("mongo", map[string]interface{}{"idGenerator": false}).
					WithCustom("gql-ts", map[string]interface{}{"path": viper.GetString("clientOut")}).
					WithCustom("vue", &vue.VueClientOptions{
						Components: map[string]vue.VCOptionComponentSpec{
							vue.VCOptionDateComponent:  {Name: "InputDateComponent", Import: "@/components/DateComponent.vue"},
							vue.VCOptionMapComponent:   {Name: "KeyValueComponent", Import: "@/components/KVComponent.vue"},
							vue.VCOptionColorComponent: {Name: "ColorPickerComponent", Import: "@/components/ColorPickerComponent.vue"},
						},
						//ApolloClientVar: "this.$apolloProvider.clients['statistics']",
					}).
					WithCustom(gen.CodeGeneratorOptionsName, map[string]interface{}{"AllowEmbeddedArraysForDictionary": true}),
			)
			proj.With(&gen.GQLGenerator{}).
				With(&gen.LoggerGenerator{}).
				With(&gen.HistoryGenerator{}).
				With(&gen.BitSetChangeDetectorGenerator{}).
				With(&gen.NoCacheGenerator{}).
				With(&gen.DictionariesGenerator{}).
				With(&gen.MongoGenerator{}).
				With(&gen.SequnceIDGenerator{}).
				With(&js.GQLCLientGenerator{}).
				With(&js.TSValidatorGenerator{}).
				With(&vue.VueCLientGenerator{}).
				With(&gen.CroneGenerator{}).
				With(&gen.ResourceGenerator{}).
				With(&gen.ServiceGenerator{}).
				With(&gen.Validator{})
		}

		// desc.OutputDir = viper.GetString("out")
		err = proj.Generate()
		if err != nil {
			fmt.Println("error found: ", err)
			return
		}
		if len(proj.Warnings) > 0 {
			fmt.Println("\nWarnings found: ")
			for _, w := range proj.Warnings {
				fmt.Println("\t", w)
			}
			fmt.Println("")
		}
		if viper.GetBool("print") {
			proj.Print()
		} else {
			err = proj.WriteToFiles()
			if err != nil {
				fmt.Println("error writing result: ", err)
			}
		}
	}
}
