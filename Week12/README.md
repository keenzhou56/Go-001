学习笔记
https://juejin.cn/post/6844903940190912519
https://www.bilibili.com/video/BV12p4y1W7Dz
https://lailin.xyz/post/go-training-week3-sync.html

1.自我介绍
2. 代码效率分析，考察局部性原理
3. 多核CPU场景下，cache如何保持一致、不冲突？ MESI 原理
4. uint类型溢出
5. 介绍rune类型
6. 编程题：3个函数分别打印cat、dog、fish，要求每个函数都要起一个goroutine，按照cat、dog、fish顺序打印在屏幕上100次。
7. 介绍一下channel，无缓冲和有缓冲区别
8. 是否了解channel底层实现，比如实现channel的数据结构是什么？
9. channel是否线程安全？
10. Mutex是悲观锁还是乐观锁？悲观锁、乐观锁是什么？
11. Mutex几种模式？
12. Mutex可以做自旋锁吗？
13. 介绍一下RWMutex
14. 项目中用过的锁？
15. 介绍一下线程安全的共享内存方式
16. 介绍一下goroutine
17. goroutine自旋占用cpu如何解决（go调用、gmp）
18. 介绍linux系统信号
19. goroutine抢占时机（gc 栈扫描）
20. Gc触发时机
21. 是否了解其他gc机制
22. Go内存管理方式
23. Channel分配在栈上还是堆上？哪些对象分配在堆上，哪些对象分配在栈上？
24. 介绍一下大对象小对象，为什么小对象多了会造成gc压力？
25. 项目中遇到的oom情况？
26. 项目中使用go遇到的坑？
27. 工作遇到的难题、有挑战的事情，如何解决？
28. 如何指定指令执行顺序？


转自知乎：记录下我马上直接想到的问题，算给自己出面试题做个笔记
如果要求是CS背景良好、写过20-30万行商用后台代码、1-2年go经验的，我问如下问题：
1.  1.9/1.10中，time.Now()返回的是什么时间？这样做的决定因素是什么?
  2.  golang的sync.atomic和C++11的atomic最显著的在golang doc里提到的差别在哪里，如何解决或者说规避？
  3.  1.11为止，sync.RWMutex最主要的性能问题最容易在什么常见场景下暴露。有哪些解决或者规避方法？
  4.  如何做一个逻辑正确但golang调度器(1.10)无法正确应对，进而导致无法产生预期结果的程序。调度器如何改进可以解决此问题？
  5.  列出下面操作延迟数量级(1ms, 10us或100ns等)，cgo调用c代码，c调用go代码，channel在单线程单case的select中被选中，high contention下对未满的buffered channel的写延迟。
  6.  如何设计实现一个简单的goroutine leak检测库，可用于在测试运行结束后打印出所测试程序泄露的goroutine的stacktrace以及goroutine被创建的文件名和行号。
  7.  选择三个常见golang组件（channel, goroutine, [], map, sync.Map等），列举它们常见的严重伤害性能的anti-pattern。
  8.  一个C/C++程序需要调用一个go库，某一export了的go函数需高频率调用，且调用间隔需要调用根据go函数的返回调用其它C/C++函数处理，无法改变调用次序、相互依赖关系的前提下，如何最小化这样高频调用的性能损耗？
  9.  不考虑调度器修改本身，仅考虑runtime库的API扩展，如果要给调度器添加NUMA awareness且需要支持不同拓扑，runtime库需要添加哪些函数，哪些函数接口必须改动。
  10.  stw的pause绝大多数情况下在100us量级，但有时跳增一个数量级。描述几种可能引起这一现象的触发因素和他们的解决方法。
  11.  已经对GC做了较充分优化的程序，在无法减小内存使用量的情况下，如何继续显著减低GC开销。
  12.  有一个常见说法是“我能用channel简单封装出一个类似sync.Pool功能的实现”。在多线程、high contention、管理不同资源的前提下，两者各方面性能上有哪些显著不同。
  13.  为何只有一个time.Sleep(time.Millisecond)循环的go程序CPU占用显著高于同类C/C++程序？或请详述只有一个goroutine的Go程序，每次time.Sleep(time.Millisecond)调用runtime所发生的事情。
  14.  一个Go程序如果尝试用fork()创建一个子进程，然后尝试在该子进程中对Go程序维护的一个数据结构串行化以后将串型化结果保存到磁盘。上述尝试会有什么结果？
  15.  请列举两种不同的观察GC占用CPU程度的方法，观察方法无需绝对精确，但需要可实际运用于profiling和optimization。
  16.  GOMAXPROCS与性能的关系如何？在给定的硬件、程序上，如何设定GOMAXPROCS以期获得最佳性能？
  17.  一个成熟完备久经优化的Golang后台系统，程序只有一个进程，由golang实现。其核心处理的部分由数十个goroutine负责处理数万goroutine发起的请求。由于无法设定goroutine调度优先级，使得核心处理的那数十个goroutine往往无法及时被调度得到CPU，进而影响了处理的延迟。有什么改进办法？
  18.  列举几个近期版本runtime性能方面的regression。