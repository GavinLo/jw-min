# jw-min
jw-min是一个优化和发布html和相关js、css文件的工具

## 安装

	先安装go，然后运行make && make install

## 用法

	USAGE: jw-min [options] your-html-file -o output-dir 

	OPTIONS:
		-t 	自动添加日期目录到目标目录下(Auto Add Date String in output-dir).
		-d 	调试输出(Debug output).
		-s 	静态路由文件，默认为’static.json’(Static Route File, default:static.json).
		format: {
				pattern: path,
				...
			}

		
# 开发日志

## 2017.07.17

* 添加静态路由表功能，js和css的编译时，文件路径会先通过路由表匹配，找到实际的文件之后再编译

## 2017.06.22

* `bug` 修改Makefile在install时没有将相应的第三方工具（jar包）拷贝到系统目录的bug
* `bug` 修改程序运行时，由于工作目录设置而不能调用第三方工具问题