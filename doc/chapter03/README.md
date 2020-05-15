# Redis请求报文组装

## 请求报文组包

发送给Redis服务端的所有参数都是二进制安全的。以下是通用形式：

```
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
```

整个请求大体可分为2部分：

**报文头**

报文头以`*`开头，后跟的参数的个数，最后以`\r\n`结尾，参数个数指：`set name zhangsan`这个命令，表示3个参数，组装后的报文头为

```
*3\r\n
```

**报文体**

报文体组装分为5个步骤：

1、`$`开头，

2、后跟第一个参数字节大小，

3、以`\r\n`结尾，

4、参数值

5、最后以`\r\n`结尾

如果有多个参数则重复次动作，如我们使用`set name zhangsan`这个命令，这个命令有3个参数，则组装后的报文体为：

```
*3\r\n$3\r\nset\r\n$4\r\nname\r\n$8\r\nzhangsan\r\n
```

**完整请求报文**

```
$3\r\nset\r\n$4\r\nname\r\n$8\r\nzhangsan\r\n
```

## 响应报文解析

- 用单行回复，回复的第一个字节将是“+”
- 错误消息，回复的第一个字节将是“-”
- 整型数字，回复的第一个字节将是“:”
- 批量回复，回复的第一个字节将是“$”
- 多个批量回复，回复的第一个字节将是“*”

