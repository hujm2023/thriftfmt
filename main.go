package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"github.com/hujm2023/thriftfmt/parser"
)

var (
	patchRequired         = pflag.Bool("patchRequired", false, "if true will patch the miss required for field in struct or others. (default false)")
	indent                = pflag.Int("indent", 4, "indentation level for indentation of struct fields.")
	overwrite             = pflag.Bool("overwrite", false, "if true will overwrite existing file. (default false)")
	verbose               = pflag.Bool("verbose", false, "if true will print the processing logs. (default false)")
	enumFieldDelimiter    = pflag.StringP("enumFieldDelimiter", "e", ",", "delimiter for enum fields.")
	structFieldDelimiter  = pflag.StringP("structFieldDelimiter", "s", ",", "delimiter for struct fields.")
	serviceFieldDelimiter = pflag.StringP("serviceFieldDelimiter", "f", ",", "delimiter for service fields.")
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("filePath is required")
		os.Exit(1)
	}

	pflag.Parse()
	parser.SetTabSize(*indent)
	parser.SetPatchRequired(*patchRequired)
	parser.SetEnumDelimiter(*enumFieldDelimiter)
	parser.SetStructDelimiter(*structFieldDelimiter)
	parser.SetServiceDelimiter(*serviceFieldDelimiter)

	filePath := os.Args[1]
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	// if the path is dir, overwrite will be set to true
	if stat.IsDir() {
		err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(info.Name(), ".thrift") {
				return nil
			}

			return processFile(path, true)
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err = processFile(filePath, *overwrite); err != nil {
			log.Fatal(err)
			return
		}
	}

	fmt.Printf("%s format finished!\n", absPath)
}

func processFile(filePath string, ow bool) error {
	print("processing file %s\n", filePath)
	formatted, err := doFormat(filePath)
	if err != nil {
		return err
	}
	if ow {
		ff, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer ff.Close()
		_, _ = ff.WriteString(formatted)
		print("!!file %s overwritten!!\n", filePath)
	} else {
		print("!file %s parsed!\n", filePath)
		fmt.Println(formatted)
	}
	return nil
}

func doFormat(filePath string) (string, error) {
	// 解析thrift文件
	tr, err := parser.ParseFileAsFormatters(filePath, nil, false)
	if err != nil {
		return "", err
	}

	return parser.FormatInline(tr), nil
}

func print(fotmat string, args ...interface{}) {
	if *verbose {
		fmt.Printf(fotmat, args...)
	}
}
