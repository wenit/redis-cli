## 初始化命令行参数

使用`flag`包进行命令参数的解析，flag包支持String、Int、Bool等类型参数的读取

``` go
var host string
var port int

func init() {
	flag.StringVar(&host, "h", "localhost", "host")
	flag.IntVar(&port, "p", 6379, "port")
}


func main() {
	// 解析命令行参数
	flag.Parse()
 
    // 获取命令行输入
    loop()
}
```



## 命令行输入
使用`os.Stdin`获取标准输入，并使用`bufio.NewReader`包简化读取过程

``` go
func loop(conn net.Conn) {
	for {
		fmt.Printf("%s:%d>", host, port)
		bio := bufio.NewReader(os.Stdin)
		input, _, err := bio.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", input)
	}
}
```



## 预览效果

使用 `go build` 编译后，执行`redis-cli`命令，可以获取命令行参数，参数错误或者输入不合法，会出现提示。

![image](../image/01-first_cmd_line.gif)