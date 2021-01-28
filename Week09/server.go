package main

import (
	"bufio"
	context "context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {

	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 使用TCP协议监听端口
	remote := "127.0.0.1:8081"
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", remote)
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			log.Println("main defer recover error:", err)
		}
		tcpListener.Close()
		log.Println("===============server closed", "==============")

	}()

	//通知连接关闭
	ctx, cancelFunc := context.WithCancel(context.Background())

	// 监听signal信号
	signalChan := make(chan os.Signal, 1)
	go func() {
		signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	}()

	//监听signal信号,收到signal信号通知其他goroutine关闭连接
	go func() {
		for {
			select {
			case <-signalChan:
				log.Println("received os signal, ready cancel other conn")
				cancelFunc()
			}
		}
	}()

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Errorf("accept failed,err is %+v", err)
			continue
		}

		accept := runtime.NumCPU()
		// 连接成功，开始监听消息
		for i := 0; i < accept; i++ {
			go tcpPipe(ctx, tcpConn)
		}

	}

}

func tcpPipe(ctx context.Context, conn *net.TCPConn) {
	cancelCtx, _ := context.WithCancel(ctx)

	var msgChan = make(chan []byte, 1)
	go read(cancelCtx, conn, msgChan)
	go write(cancelCtx, conn, msgChan)
}

func read(ctx context.Context, conn *net.TCPConn, dataChan chan []byte) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("close read")
			close(dataChan)
			conn.Close()
			return
		default:
			reader := bufio.NewReader(conn)
			var buf [1024]byte
			n, err := reader.Read(buf[:])
			if err != nil {
				fmt.Errorf("read failed")
				break
			}
			dataChan <- buf[:n]

		}
	}
}

func write(ctx context.Context, conn *net.TCPConn, dataChan chan []byte) {

	for {
		select {
		case <-ctx.Done():
			fmt.Println("close write")
			return
		default:

			dataByte := <-dataChan
			// TODO
			_, err := conn.Write(dataByte[:])
			fmt.Println("send message：" + string(dataByte[:]))
			if err != nil {
				break
				fmt.Errorf("write message failed,error is %+v", err)
			}

		}
	}

}
