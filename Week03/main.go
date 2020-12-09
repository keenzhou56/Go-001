package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// 启动http服务
func startServer(ctx context.Context, port int, pattern string) error {

	mux := http.ServeMux{}
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(pattern))
	})

	addr := fmt.Sprintf(":%d", port)
	s := http.Server{
		Addr:    addr,
		Handler: &mux,
	}

	go func() {
		<-ctx.Done()
		s.Shutdown(ctx)
		log.Printf("Shutdown server: %d", port)
	}()

	log.Println("Start server on port:", port)

	return s.ListenAndServe()

}

func main() {

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return startServer(ctx, 8080, "/in")
	})

	g.Go(func() error {
		return startServer(ctx, 8081, "/reg")
	})

	g.Go(func() error {
		return startServer(ctx, 8082, "/out")
	})

	// 终止信号
	g.Go(func() error {

		defer func() {
			if err := recover(); err != nil {
				log.Print(err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-quit:
			return fmt.Errorf("quit by signal")
		case <-ctx.Done():
			log.Println(ctx.Err())
		}
		return nil

	})

	if err := g.Wait(); err != nil {
		log.Printf("Shutdown all server: %v", err)
	}

}
