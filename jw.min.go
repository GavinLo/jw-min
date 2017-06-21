/**
 * jw-min是一个优化和发布html和相关js、css文件的工具
 * 它使用了一些第三方工具来帮助其完成工作
 * https://developers.google.com/closure/compiler/
 * http://yui.github.io/yuicompressor/
 * http://code.google.com/p/htmlcompressor/
 **/
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var Debug = false

type Options struct {
	ShowVersion bool
	InputFile   string
	OutputDir   string
}

func (this *Options) String() string {
	return "\nInputFile: " + this.InputFile + "\n" +
		"OutputDir: " + this.OutputDir
}

const (
	version = "jw-win version 1.0.0"

	java = "java"
	jar  = "-jar"
	// https://developers.google.com/closure/compiler/
	jscompiler = "compiler.jar"
	// http://yui.github.io/yuicompressor/
	yuicompressor = "yuicompressor-2.4.8.jar"
	// http://code.google.com/p/htmlcompressor/
	htmlcompressor = "htmlcompressor-1.5.3.jar"
)

func chpwd() string {
	_, pwf, _, _ := runtime.Caller(0)
	pwd := path.Dir(pwf)
	os.Chdir(pwd)
	return pwd
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	// 获取处理参数
	options, err := parseOptions()
	if err != nil {
		fmt.Println(err)
		usage()
		return
	}
	if options.ShowVersion {
		fmt.Println(version)
		return
	}
	if Debug {
		fmt.Println("Options:", options)
	}
	err = os.MkdirAll(options.OutputDir, os.FileMode(0766))
	if err != nil {
		fmt.Println(err)
		usage()
		return
	}

	// 预定义一些值
	inputDir := path.Dir(options.InputFile)
	inputName := strings.Replace(path.Base(options.InputFile), path.Ext(options.InputFile), "", -1)
	if Debug {
		fmt.Println("InputName:", inputName)
	}
	inputFile, err := os.Open(options.InputFile)
	defer inputFile.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	inputData, err := ioutil.ReadAll(inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	inputString := string(inputData)

	// 寻找输入文件（html）中的js声明，并编译js文件
	var jsCmd *exec.Cmd = nil
	outputJS := path.Join("js", inputName) + ".js"
	var scripts []string
	reg := regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	strs := reg.FindAllString(inputString, -1)
	if len(strs) > 0 {
		firstMatch := true
		inputString = reg.ReplaceAllStringFunc(inputString, func(string) string {
			if firstMatch {
				firstMatch = false
				return "<script src=\"" + outputJS + "\"></script>"
			}
			return ""
		})
		reg = regexp.MustCompile("src[\\s]*=[\\s]*\"([\\S]*)\"")
		for _, str := range strs {
			sub := reg.FindStringSubmatch(str)
			if len(sub) > 1 {
				script := sub[1]
				if len(script) > 0 {
					scripts = append(scripts, path.Join(inputDir, script))
				}
			}
		}
	}
	if len(scripts) > 0 {
		if Debug {
			fmt.Println("Scripts:", scripts)
		}
		var jsArgs []string
		outputJS, err := filepath.Abs(path.Join(options.OutputDir, outputJS))
		if err != nil {
			fmt.Printf("[ERROR] 获取绝对路径失败(Get Absolute Path Failed): %s\n", err.Error())
			return
		}
		outputJSDir := path.Dir(outputJS)
		os.MkdirAll(outputJSDir, os.FileMode(0766))
		jsArgs = append(jsArgs, jar, jscompiler, "--js_output_file="+outputJS)
		jsArgs = append(jsArgs, scripts...)
		jsCmd = exec.Command(java, jsArgs...)
		if Debug {
			fmt.Println("[INFO ] jsCmd: ", jsCmd)
		}
	}

	// 寻找输入文件（html）中的css声明，并优化css文件
	var cssCmds []*exec.Cmd
	var csss []string
	reg = regexp.MustCompile("\\<link[\\S\\s]+?/\\>")
	strs = reg.FindAllString(inputString, -1)
	if len(strs) > 0 {
		reg_rel := regexp.MustCompile("rel[\\s]*=[\\s]*\"([\\S]*)\"")
		reg_href := regexp.MustCompile("href[\\s]*=[\\s]*\"([\\S]*)\"")
		for _, str := range strs {
			rels := reg_rel.FindStringSubmatch(str)
			if len(rels) > 1 {
				rel := rels[1]
				if rel != "stylesheet" {
					continue
				}
			}
			hrefs := reg_href.FindStringSubmatch(str)
			if len(hrefs) > 1 {
				href := hrefs[1]
				if len(href) > 0 {
					// csss = append(csss, path.Join(inputDir, href))
					csss = append(csss, href)
				}
			}
		}
	}
	if len(csss) > 0 {
		if Debug {
			fmt.Println("CSS:", csss)
		}
		for _, css := range csss {
			cssIn, err := filepath.Abs(path.Join(inputDir, css))
			if err != nil {
				fmt.Printf("[ERROR] 获取绝对路径失败(Get Absolute Path Failed): %s\n", err.Error())
				return
			}
			cssOut, err := filepath.Abs(path.Join(options.OutputDir, css))
			if err != nil {
				fmt.Printf("[ERROR] 获取绝对路径失败(Get Absolute Path Failed): %s\n", err.Error())
				return
			}
			cssOutDir := path.Dir(cssOut)
			os.MkdirAll(cssOutDir, os.FileMode(0766))
			cssCmd := exec.Command(java, jar, yuicompressor, cssIn, "-o", cssOut)
			if Debug {
				fmt.Println("[INFO ] cssCmd: ", cssCmd)
			}
			cssCmds = append(cssCmds, cssCmd)
		}
	}

	// 优化html
	// 生成中间文件
	outputObjHtml, err := filepath.Abs(path.Join(options.OutputDir, inputName) + ".obj.html")
	if err != nil {
		fmt.Printf("[ERROR] 获取绝对路径失败(Get Absolute Path Failed): %s\n", err.Error())
		return
	}
	err = ioutil.WriteFile(outputObjHtml, []byte(inputString), os.FileMode(0644))
	if err != nil {
		fmt.Println(err)
		return
	}

	outputHtml, err := filepath.Abs(path.Join(options.OutputDir, inputName) + ".html")
	if err != nil {
		fmt.Printf("[ERROR] 获取绝对路径失败(Get Absolute Path Failed): %s\n", err.Error())
		return
	}
	htmlCmd := exec.Command(java, jar, htmlcompressor, "-o", outputHtml, outputObjHtml)
	if Debug {
		fmt.Println("[INFO ] htmlCmd: ", htmlCmd)
	}

	// 执行命令
	chpwd()
	fmt.Println("[INFO ] 编译js脚本(Compile Scripts)...")
	jsOut, err := jsCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[ERROR] 编译js脚本失败(Compile Scripts Failed): %s\n", string(jsOut))
		return
	}
	fmt.Println("[INFO ] 编译css文件(Compile CSS files)...")
	for _, cssCmd := range cssCmds {
		cssOut, err := cssCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[ERROR] 编译css文件失败(Compile CSS files Failed): %s\n", string(cssOut))
			return
		}
	}
	fmt.Println("[INFO ] 编译Html(Compile Html)...")
	htmlOut, err := htmlCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[ERROR] 编译Html失败(Compile Html Failed): %s\n", htmlOut)
		return
	}
	// 删除中间文件
	os.Remove(outputObjHtml)

	// 完成
	fmt.Println("[INFO ] 完成(Done).")
}

func parseOptions() (*Options, error) {
	options := &Options{ShowVersion: false}
	var err error
	outputDirIndex := -1
	autoAddTime := false
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		ni := i + 1
		if arg == "-v" {
			options.ShowVersion = true
			return options, nil
		} else if arg == "-o" {
			if len(os.Args) > ni {
				outputDirIndex = ni
				options.OutputDir = os.Args[outputDirIndex]
				if autoAddTime {
					options.OutputDir = path.Join(options.OutputDir, time.Now().Format("2006-01-02 15-04-05"))
				}
				continue
			} else {
				return nil, errors.New("No output directory found.")
			}
		} else if arg == "-t" {
			autoAddTime = true
			options.OutputDir = path.Join(options.OutputDir, time.Now().Format("2006-01-02 15-04-05"))
		} else if arg == "-d" {
			Debug = true
		} else {
			if i != outputDirIndex {
				options.InputFile, err = filepath.Abs(arg)
			}
		}
	}
	return options, err
}

func usage() {
	fmt.Println("USAGE: jw-min [options] your-html-file -o output-dir ")
	fmt.Println("\nOPTIONS:")
	fmt.Println("\t-t \tAuto Add Date String in output-dir.")
	fmt.Println("\t-d \tDebug output.")
}
