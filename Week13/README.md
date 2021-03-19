学习笔记

国内：积累一波，TFS Taobao FileSystem -> OSS  （点 -> 面） 腾讯SharpP（h265）
海外：Facebook （Haystack、Warm Blob Strogage）、Google （GFS、GFS v2）、Uber、Netflix、Amazon、Linkerin）



eng.uber.com
research.fb.com/publications
netflix enginering blog

性能优化


小对象结构体合并 **
bytes.Buffer
slice map 预创建
长调用栈 (频繁使用defer)
避免频繁创建临时对象
字符串拼接 strings.Builder
不必要的 memory copy
分析内存逃逸
Readv Writev 非连续内存

https://github.com/dgryski/go-perfbook
https://cch123.github.io/perf_opt/

开源项目
https://github.com/alibaba/canal
https://github.com/bilibili/discovery
https://github.com/go-kratos/kratos
github.com/facebook/ent
https://github.com/uber-go/automaxprocs

https://github.com/chai2010/advanced-go-programming-book
https://github.com/cch123

文章
https://gobyexample.com/
What are microservices?
关于Golang GC的一些误解--真的比Java算法更领先吗？
什么是写屏障、混合写屏障，如何实现？
Microservice 微服务的理论模型和现实路径
微服务架构 BFF和网关是如何演化出来的 
公众号【迈莫coding】https://mp.weixin.qq.com/s/jywYEckHzVj20K0uQ0PQdQ
https://go101.org/article/101.html
https://www.ardanlabs.com/blog/   特别是 William Kennedy 的文章
https://martinfowler.com/bliki/CQRS.html  Martin Folwer 关于 CQRS 的解释
图示 Golang 垃圾回收机制 https://zhuanlan.zhihu.com/p/297177002?utm_source=wechat_session&utm_medium=social&utm_oi=26711194337280&utm_campaign=shareopn
康威定律： https://www.cnblogs.com/still-smile/p/11646609.html
https://blog.golang.org/wire
google API 设计指南 https://cloud.google.com/apis/design(毛大推荐)
Consistent Hashing with Bounded Loads
限流算法：github.com/go-kratos/kratos/pkg/ratelimit/bbr
https://roaringbitmap.org/
https://github.com/intercom/lease
https://research.fb.com/publications/scaling-memcache-at-facebook/
万亿级日访问量下，Redis 在微博的 9 年优化历程
[go 语言]go goroutine调度机制 && goroutine池
Goroutine并发调度模型深度解析之手撸一个协程池
fasthttp 快在哪里 
Go 运行程序中的线程数 
Go：g0，特殊的 Goroutine
Go：Goroutine 的切换过程实际上涉及了什么 
signal_based_preemption
书籍
《UNIX环境高级编程》
《Go 语言问题集》
《Google SRE》 https://landing.google.com/sre/books/
《Effective Go》
《Google软件测试之道》
《领域驱动设计》
《架构整洁之道》
《The Site Reliability Workbook》
《代码整洁之道》
《领域驱动设计精粹》
https://research.fb.com/publications/
https://eng.uber.com
课程时间点