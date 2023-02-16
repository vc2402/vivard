package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vc2402/vivard/gen"
	"github.com/vc2402/vivard/gen/js"
	"github.com/vc2402/vivard/gen/vue"
)

func main() {
	pflag.String("package", "test", "default package name")
	pflag.String("in", ".", "Input directory")
	pflag.String("out", ".", "Output directory")
	pflag.String("clientOut", "/home/victor/work/vivasoft/vue/gen-test/src/generated", "Output directory for client files")
	pflag.Bool("print", false, "Print result to stdout")
	pflag.String("cfgPath", ".", "Path to config file")
	pflag.String("cfg", "gen.json", "Config file name")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetConfigName(viper.GetString("cfg"))
	viper.AddConfigPath(viper.GetString("cfgPath"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Println("no files to parse given")
		return
	}
	in := viper.GetString("in")
	if in != "" && in != "." {
		for i, fn := range args {
			if !filepath.IsAbs(fn) {
				args[i] = filepath.Join(in, fn)
			}
		}
	}
	res, err := gen.Parse(args)
	if err != nil {
		fmt.Printf("errors found: %v\n", err)
	} else {
		// fmt.Printf("parsed: %#+v: \n", res)
		desc := gen.New(
			res,
			gen.Options(viper.GetString("out")).
				With(gen.PackagePrefixOption("parnasas.lt")).
				With(gen.UnknownAnnotationWarning).
				With(gen.NullablePointers).
				WithCustom("mongo", map[string]interface{}{"idGenerator": false}).
				WithCustom("gql-ts", map[string]interface{}{"path": viper.GetString("clientOut")}).
				WithCustom("vue", &vue.VueClientOptions{
					Components: map[string]vue.VCOptionComponentSpec{
						vue.VCOptionDateComponent:  {Name: "InputDateComponent", Import: "@/components/DateComponent.vue"},
						vue.VCOptionMapComponent:   {Name: "KeyValueComponent", Import: "@/components/KVComponent.vue"},
						vue.VCOptionColorComponent: {Name: "ColorPickerComponent", Import: "@/components/ColorPickerComponent.vue"},
					},
					ApolloClientVar: "this.$apolloProvider.clients['statistics']",
				}).
				WithCustom(gen.CodeGeneratorOptionsName, map[string]interface{}{"AllowEmbeddedArraysForDictionary": true}),
		)
		desc.With(&gen.GQLGenerator{}).
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

		// desc.OutputDir = viper.GetString("out")
		err = desc.Generate()
		if err != nil {
			fmt.Println("error found: ", err)
			return
		}
		if len(desc.Warnings) > 0 {
			fmt.Println("\nWarnings found: ")
			for _, w := range desc.Warnings {
				fmt.Println("\t", w)
			}
			fmt.Println("")
		}
		if viper.GetBool("print") {
			desc.Print()
		} else {
			err = desc.WriteToFiles()
			if err != nil {
				fmt.Println("error writing result: ", err)
			}
		}
	}
}
