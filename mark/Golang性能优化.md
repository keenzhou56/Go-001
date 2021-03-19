
Golang性能优化

    1.内存优化
        1.1 小对象合并成结构体一次分配，减少内存分配次数
        1.2 缓存区内容一次分配足够大小空间，并适当复用
        1.3 slice和map采make创建时，预估大小指定容量
        1.4 长调用栈避免申请较多的临时对象
        1.5 避免频繁创建临时对象
        1.6 大的struct对象尽量用指针传递，避免深拷贝
    2.并发优化
        2.1 goroutine池
        2.2 减少系统调用
        2.3 减少锁，减少大锁
        2.4 请求合并singleflight
        2.5 压缩协议PB
        2.6 批量协议
        2.7 并行请求errgroup
        2.8 merge请求pipeline
    3.其它优化
        3.1 避免使用CGO或者减少CGO调用次数

1.内存优化
1.1 小对象合并成结构体一次分配，减少内存分配次数

做过C/C++的同学可能知道，小对象在堆上频繁地申请释放，会造成内存碎片（有的叫空洞），导致分配大的对象时无法申请到连续的内存空间，一般建议是采用内存池。Go runtime底层也采用内存池，但每个span大小为4k，同时维护一个cache。cache有一个0到n的list数组，list数组的每个单元挂载的是一个链表，链表的每个节点就是一块可用的内存，同一链表中的所有节点内存块都是大小相等的；但是不同链表的内存大小是不等的，也就是说list数组的一个单元存储的是一类固定大小的内存块，不同单元里存储的内存块大小是不等的。这就说明cache缓存的是不同类大小的内存对象，当然想申请的内存大小最接近于哪类缓存内存块时，就分配哪类内存块。当cache不够再向spanalloc中分配。

建议：小对象合并成结构体一次分配，示意如下：

goroutine的实现，是通过同步来模拟异步操作。在如下操作操作不会阻塞go runtime的线程调度：
for k, v := range m {
    x := struct {k , v string} {k, v} // copy for capturing by the goroutine
    go func() {
        // using x.k & x.v
    }()
}
1.2 缓存区内容一次分配足够大小空间，并适当复用

在协议编解码时，需要频繁地操作[]byte，可以使用bytes.Buffer或其它byte缓存区对象。

建议：bytes.Buffer等通过预先分配足够大的内存，避免当Grow时动态申请内存，这样可以减少内存分配次数。同时对于byte缓存区对象考虑适当地复用。
1.3 slice和map采make创建时，预估大小指定容量

slice和map与数组不一样，不存在固定空间大小，可以根据增加元素来动态扩容。

slice初始会指定一个数组，当对slice进行append等操作时，当容量不够时，会自动扩容：

如果新的大小是当前大小2倍以上，则容量增涨为新的大小，否则循环以下操作：如果当前容量小于1024，按2倍增加；否则每次按当前容量1/4增涨，直到增涨的容量超过或等新大小。

map的扩容比较复杂，每次扩容会增加到上次容量的2倍。它的结构体中有一个buckets和oldbuckets，用于实现增量扩容：

正常情况下，直接使用buckets，oldbuckets为空，如果正在扩容，则oldbuckets不为空，buckets是oldbuckets的2倍。

建议：初始化时预估大小指定容量
m := make(map[string]string, 100)
s := make([]string, 0, 100) // 注意：对于slice make时，第二个参数是初始大小，第三个参数才是容量
1.4 长调用栈避免申请较多的临时对象

goroutine的调用栈默认大小是4K（1.7修改为2K），它采用连续栈机制，当栈空间不够时，Go runtime会不断扩容：

当栈空间不够时，按2倍增加，原有栈的变量崆直接copy到新的栈空间，变量指针指向新的空间地址。退栈会释放栈空间的占用，GC时发现栈空间占用不到1/4时，则栈空间减少一半。比如栈的最终大小2M，则极端情况下，就会有10次的扩栈操作，这会带来性能下降。

建议：控制调用栈和函数的复杂度，不要在一个goroutine做完所有逻辑，如查的确需要长调用栈，而考虑goroutine池化，避免频繁创建goroutine带来栈空间的变化。
1.5 避免频繁创建临时对象

Go在GC时会引发stop the world，即整个情况暂停。虽1.7版本已大幅优化GC性能，1.8甚至量坏情况下GC为100us。但暂停时间还是取决于临时对象的个数，临时对象数量越多，暂停时间可能越长，并消耗CPU。

建议：GC优化方式是尽可能地减少临时对象的个数，尽量使用局部变量，对多个局部变量合并一个大的结构体或数组，减少扫描对象的次数，一次回尽可能多的内存。
1.6 大的struct对象尽量用指针传递，避免深拷贝
2.并发优化
2.1 goroutine池

goroutine虽轻量，但对于高并发的轻量任务处理，频繁来创建goroutine来执行，执行效率并不会太高效：

过多的goroutine创建，会影响go runtime对goroutine调度，以及GC消耗；

高并时若出现调用异常阻塞积压，大量的goroutine短时间积压可能导致程序崩溃。

例如library/cache/cache.go启动了多个wocker共同消费一个chan：
......
// New new a cache struct.
func New(worker, size int) *Cache {
    if worker <= 0 {
        worker = 1
    }
    c := &Cache{
        ch:     make(chan func(), size),
        worker: worker,
    }
    c.waiter.Add(worker)
    for i := 0; i < worker; i++ {
        go c.proc()
    }
    return c
}
  
func (c *Cache) proc() {
    defer c.waiter.Done()
    for {
        f := <-c.ch
        if f == nil {
            return
        }
        wrapFunc(f)()
        stats.State("cache_channel", int64(len(c.ch)))
    }
}
......
2.2 减少系统调用

goroutine的实现，是通过同步来模拟异步操作。在如下操作操作不会阻塞go runtime的线程调度：

    网络IO锁
    channel
    time.sleep
    基于底层系统异步调用的Syscall

下面阻塞会创建新的调度线程：

    本地IO调用
    基于底层系统同步调用的Syscall
    CGo方式调用C语言动态库中的调用IO或其它阻塞
    网络IO可以基于epoll的异步机制（或kqueue等异步机制），但对于一些系统函数并没有提供异步机制。例如常见的posix api中，对文件的操作就是同步操作。虽有开源的fileepoll来模拟异步文件操作。但Go的Syscall还是依赖底层的操作系统的API。系统API没有异步，Go也做不了异步化处理。

建议：把涉及到同步调用的goroutine，隔离到可控的goroutine中，而不是直接高并的goroutine调用。
2.3 减少锁，减少大锁

Go推荐使用channel的方式去通讯，而不是共享内存。若goroutine间存在大锁，可以把锁的粒度拆细。
2.4 请求合并singleflight

singleflight处理相同key的多个请求访问磁盘，只有一个请求访问磁盘，其他等待结果。常常用于回源。
import "golang.org/x/sync/singleflight"
 
var singleGroup singleflight.Group
 
v, err, _ := singleGroup.Do(aid, func() (res interface{}, err error) {
    return s.dao.ResTagMap(c, aid)
})
2.5 压缩协议PB

Protocol Buffers是google 的一种数据交换的格式。它是一种二进制的数据格式，具有更高的传输，打包和解包效率。推荐在缓存打包存储、数据传输时替代JSON。

Protocol buffers在序列化结构化数据方面有许多优点：

    更简单
    数据描述文件只需原来的1/10至1/3
    解析速度是原来的20倍至100倍
    减少了二义性
    生成了更容易在编程中使用的数据访问类
    支持多种编程语言

2.6 批量协议

对请求访问数据的接口需提供批量协议，可以减少非常多的封包解包、IO和QPS来提升性能。例如稿件RPC批量请求：
as, err := s.arcRPC.Archives3(c, &archive.ArgAids2{Aids: aids})

redis也可以通过pipeline批量加入多个命令，例如点赞的redis缓存。
// AddCacheUserLikeList .
func (d *Dao) AddCacheUserLikeList(c context.Context, mid int64, miss []*model.ItemLikeRecord, businessID int64, state int8) (err error) {
    if len(miss) == 0 {
        return
    }
    limit := d.BusinessIDMap[businessID].UserLikesLimit
    var count int
    conn := d.redis.Get(c)
    defer conn.Close()
    key := userLikesKey(businessID, mid, state)
    if err = conn.Send("DEL", key); err != nil {
        err = pkgerr.Wrap(err, "")
        PromError("redis:用户点赞列表")
        log.Errorv(c, log.KV("AddCacheUserLikeList", fmt.Sprintf("conn.Send(DEL, %s) error(%+v)", key, err)))
        return
    }
    count++
    for _, item := range miss {
        id := item.MessageID
        score := int64(item.Time)
        if err = conn.Send("ZADD", key, "CH", score, id); err != nil {
            err = pkgerr.Wrap(err, "")
            PromError("redis:用户点赞列表")
            log.Errorv(c, log.KV("log", fmt.Sprintf("conn.Send(ZADD, %s, %d, %v) error(%v)", key, score, id, err)))
            return
        }
        count++
    }
    if err = conn.Send("ZREMRANGEBYRANK", key, 0, -(limit + 1)); err != nil {
        err = pkgerr.Wrap(err, "")
        PromError("redis:用户点赞列表rm")
        log.Errorv(c, log.KV("log", fmt.Sprintf("conn.Send(ZREMRANGEBYRANK, %s, 0, %d) error(%v)", key, -(limit+1), err)))
        return
    }
    count++
    if err = conn.Send("EXPIRE", key, d.redisUserLikesExpire); err != nil {
        err = pkgerr.Wrap(err, "")
        PromError("redis:用户点赞列表过期")
        log.Errorv(c, log.KV("log", fmt.Sprintf("conn.Send(EXPIRE, %s, %d) error(%v)", key, d.redisUserLikesExpire, err)))
        return
    }
    count++
    if err = conn.Flush(); err != nil {
        err = pkgerr.Wrap(err, "")
        PromError("redis:用户点赞列表flush")
        log.Errorv(c, log.KV("log", fmt.Sprintf("conn.Flush error(%v)", err)))
        return
    }
    for i := 0; i < count; i++ {
        if _, err = conn.Receive(); err != nil {
            err = pkgerr.Wrap(err, "")
            PromError("redis:用户点赞列表receive")
            log.Errorv(c, log.KV("log", fmt.Sprintf("conn.Receive error(%v)", err)))
            return
        }
    }
    return
}
2.7 并行请求errgroup

对于网关接口，常常需要聚合多个业务模块的数据。当这些业务模块之间没有依赖关系的时候，往往可以并行请求达到优化耗时的目的。于是我们可以用errgroup库了。
import "go-common/library/sync/errgroup"
  
type ABC struct {
    CBA int
}
  
func Normal() (map[int]*ABC, error) {
    var (
        abcs = make(map[int]*ABC)
        g    errgroup.Group
        err  error
    )
    for i := 0; i < 10; i++ {
        abcs[i] = &ABC{CBA: i}
    }
    g.Go(func() (err error) {
        abcs[1].CBA++
        return
    })
    g.Go(func() (err error) {
        abcs[2].CBA++
        return
    })
    if err = g.Wait(); err != nil {
        log.Error("%v", err)
        return nil, err
    }
    return abcs, nil
}
2.8 merge请求pipeline

同一个key的数据可以merge成一个请求一起发送。

建议：使用GO大仓库的library/sync/pipeline/pipeline.go，比如播放历史，根据mid聚合：
import "go-common/library/sync/pipeline"
......
func (s *Service) initMerge() {
    s.merge = pipeline.NewPipeline(s.c.Merge)
    s.merge.Split = func(a string) int {
        mid, _ := strconv.ParseInt(a, 10, 64)
        return int(mid) % s.c.Merge.Worker
    }
    s.merge.Do = func(c context.Context, ch int, values map[string][]interface{}) {
        var merges []*model.Merge
        for _, vs := range values {
            for _, v := range vs {
                merges = append(merges, v.(*model.Merge))
            }
        }
        s.dao.AddHistoryMessage(c, ch, merges)
    }
    s.merge.Start()
}
3.其它优化
3.1 避免使用CGO或者减少CGO调用次数

GO可以调用C库函数，但Go带有垃圾收集器且Go的栈动态增涨，但这些无法与C无缝地对接。Go的环境转入C代码执行前，必须为C创建一个新的调用栈，把栈变量赋值给C调用栈，调用结束现拷贝回来。而这个调用开销也非常大，需要维护Go与C的调用上下文，两者调用栈的映射。相比直接的GO调用栈，单纯的调用栈可能有2个甚至3个数量级以上。

建议：尽量避免使用CGO，无法避免时，要减少跨CGO的调用次数。
