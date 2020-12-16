学习笔记

ent gorm 械框架

git clone $GOPATH/src/golang.org/x/sync
git get -v github.com/golang/tools
go install github.com/golang/tools
go install github.com/golang/x/tools/cmd
go install github.com/golang/x/tools/gopls
go get -u github.com/newhook/go-symbols


1. data race
    go build -race
    go test -race

2. 查看汇编 向量时钟

    go tool compile -S .\src\Go-000\Week03\main.go

3. 最晚加锁，最早释放，轻量
   
   避免死锁

4. sync.atomic.value 读多写少时使用比较多
  
  benchmark

  go test -bench=.main_test.go

  copy-on-write redis fork 一个进程备份数据 bgsave

5. sync.errgroup
  增加panic保护  https://github.com/go-kratos/kratos/blob/master/pkg/sync/errgroup/errgroup.go

6. sync.Pool 看源码

7. context 放函数首参
  put 前reset
   *a = a{}
   
   
   kratos 超时处理
   计算密集 不好处理
   网络密集与context结合使用 可以被托管
https://github.com/go-kratos/kratos/blob/master/pkg/cache/redis/conn.go
line520
  wit
8. go interface  nil 坑

9. channel https://github.com/bilibili/overlord

10. https://github.com/facebookarchive/grace/blob/master/gracehttp/http.go