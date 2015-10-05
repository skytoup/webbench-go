# webbench-go
本项目是模仿c语言的[webbench](http://home.tiscali.cz/~cz210552/webbench.html)	
纯属练手，写得非常不好，而且只有原项目的一部分功能，而且非常不稳定		
如有侵权，请联系我删除

# 使用
clone下项目后，到项目目录，输入以下命令：
```shell
go build webbench.go
./webbench -h
```

详情请参考help使用

---

# 已发现存在的问题
1. client数量只能到500左右，多了会出现错误：dial tcp4 [host:port]: socket: too many open files
2. 速度不太稳定
3. 如果在读取数据，测试时间到了还会继续运行，直到读取成功

-----
##关于我
* 一枚普通的即将大三的珠海大学生
* 希望大三实习、毕业的工作地方都在珠海

-----
##联系方式
* QQ：875766917，请备注
* Mail：875766917@qq.com

-----
##开源协议（License）
GPL2.0