Go的50坑：新Golang开发者要注意的陷阱、技巧和常见错误

https://studygolang.com/articles/9995

channel之坑
    如何优雅的关闭channel: https://www.jianshu.com/p/d24dfbb33781
    1. 连续关闭两次，会导致panic
    2.  从已经关闭的channel中写入数据
    3.  从已经关闭的channel中读取数据
    1）无缓冲channel或者缓冲channel已经读取完毕（2）缓冲channel未读取完毕，可以继续读取channel中的剩余的数据

defer之坑

    1. defer函数的参数，会在defer声明时求值， 而不是函数执行时
    2. 被defer的调用会在包含的函数的末尾执行，而不是包含代码块的末尾
    3.一个很常犯的错误就是无法区分被defer的代码执行规则和变量作用规则。如果你有一个长时运行的函数，而函数内有一个 for循环试图在每次迭代时都 defer资源清理调用，那就会出现问题。
    4. 解决方法就是把代码块写成一个函数

defer，return和返回值的顺序之坑
    1.  defer 无名返回值
    2.  defer有名返回值

    1.多个defer的执行顺序为“先进后出”；

    2.defer、return、返回值三者的执行逻辑应该是：return最先执行，return负责将结果写入返回值中；接着defer开始执行一些收尾工作；最后函数携带当前返回值退出。 

    解释两种结果的不同：

    1. 第一个例子 函数的返回值没有被提前声明，其值来自于其他变量的赋值，而defer中修改的也是其他变量，而非返回值本身，因此函数退出时返回值并没有被改变。 
    2. 第二个例子 函数的返回值被提前声明，也就意味着defer中是可以调用到真实返回值的，因此defer在return赋值返回值 i 之后，再一次地修改了 i 的值，最终函数退出后的返回值才会是defer修改过的值。

HTTP之坑
    1. 关闭http的响应

             当你使用标准http库发起请求时，你得到一个http的响应变量。如果你不读取响应主体，你依旧需要关闭它。大多数情况下，当你的http响应失败时， resp变量将为 nil，而 err变量将是 non-nil。然而，当你得到一个重定向的错误时，两个变量都将是 non-nil。这意味着你最后依然会内存泄露。

通过在http响应错误处理中添加一个关闭 non-nil响应主体的的调用来修复这个问题。另一个方法是使用一个 defer调用来关闭所有失败和成功的请求的响应主体。

    2. 关闭http链接

    
                一些HTTP服务器保持会保持一段时间的网络连接（根据HTTP 1.1的说明和服务器端的“keep-alive”配置）。默认情况下，标准http库只在目标HTTP服务器要求关闭时才会关闭网络连接。这意味着你的应用在某些条件下消耗完sockets/file的描述符。

    
                你可以通过设置请求变量中的 Close域的值为 true，来让http库在请求完成时关闭连接。

    
                另一个选项是添加一个 Connection的请求头，并设置为 close。目标HTTP服务器应该也会响应一个 Connection: close的头。当http库看到这个响应头时，它也将会关闭连接。

 func main() {
    resp, err := http.Get("http://baidu.com")
    if resp != nil {
        defer resp.Body.Close()
    }


    if err != nil {
        fmt.Println(err)
        return
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(body))
}

func main() {
    tr := &http.Transport{DisableKeepAlives: true}
    client := &http.Client{Transport: tr}
    resp, err := client.Get("http://baidu.com")
    if resp != nil {
        defer resp.Body.Close()


    }

    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(resp.StatusCode)
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(len(string(body)))
}


interface之坑
    1.  interface 和 nil

        func main() {
            var p interface{} = nil
            if p == nil {
                fmt.Println("p is nil")
            } else {
                fmt.Println("p is not nil")
            }
        }
        接口类型的变量底层是作为两个成员来实现，一个是type，一个是data。type用于存储变量的动态类型，data用于存储变量的具体数据

        由于nil是untyped(无类型)，而又将nil赋值给了变量val，所以val实际上存储的是(nil, nil)。因此很容易就知道val和nil的相等比较是为true的。

        interface类型的变量和nil的相等比较出现最多的地方应该是error接口类型的值与nil的比较
        无论该指针的值是什么：(*interface{}, nil)，这样的接口值总是非nil的，即使在该指针的内部为nil。

        error是一个接口类型，test方法中返回的指针p虽然数据是nil，但是由于它被返回成包装的error类型，也即它是有类型的。所以它的底层结构应该是(*data, nil)，很明显它是非nil的。

        type data struct {
}


func (this *data) Error() string {
    return ""
}

func bad() bool {
    return true
}

func test() error {
    var p *data = nil
    if bad() {
        return p
    }
    return nil
}

func main() {
    var e error = test()
    if e == nil || (reflect.ValueOf(e).Kind() == reflect.Ptr && reflect.ValueOf(e).IsNil()) {
        fmt.Println("e is nil")
    } else {
        fmt.Println("e is not nil")
    }
}

总的来说：只有当接口的类型和值都为nil的时候，接口变量才为nil。 


 map之坑
     1. map引用不存在的key，不会报错
     2. map利用range循环时，并不是录入顺序，而是随机顺序
     3. 更新map的值
      4. 并发访问map
       这里，go官博是有说明的，即并发访问map是不安全的，会出现未定义行为，导致程序退出


    1. 切片作为参数

              数组在作为参数时，其实作为值来传递的。
              切片在作为参数时，其实作为引用来传递的。
func main() {
    arr := [5]int{1, 2, 3, 4, 5}
    mod(arr)
    fmt.Println("main array:", arr)
 
    arr1 := []int{1, 2, 3, 4, 5}
    mod1(arr1)
    fmt.Println("main slice:", arr1)
}
 
func mod(a [5]int) {
    a[0] = 9
    fmt.Println("mod array", a)
}
 
func mod1(a []int) {
    a[0] = 9
    fmt.Println("mod slice: ", a)
}

输出结果：

    2. 切片遍历

             在使用 range 遍历 slice 的时候，range 会创建每个元素的副本
func main() {
    s := []int{1, 2, 3, 4, 5}
    for i, v := range s {
        fmt.Printf("value: %d value address: %X  ElemAddr: %X\n", v, &v, &s[i])
    }
}

输出结果：


    3. append

             append会改变切片的地址。

 
func main() {
    s := []int{1, 2, 3}
    test(s)
    for _, v := range s {
        fmt.Println(v)
    }
}
 
func test(arr []int) {
    arr = append(arr, 4)
}

输出结果：

可以看到并没打印出来4，因为切片的地址已经改变了

如果要打印，请这样操作：

 
func main() {
    s := []int{1, 2, 3}
    res := test(s)
    for _, v := range res {
        fmt.Println(v)
    }
}
 
func test(arr []int) []int {
    arr = append(arr, 4)
    return arr
}

