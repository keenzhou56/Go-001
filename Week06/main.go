// 测试 rollingnumber

package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"Go-000/Week06/pkg/stat/metric/rollingnumber"
)

var (
	rn     *rollingnumber.RollingNumber
	logger = log.New()
)

func init() {
	rn = rollingnumber.NewRollingNumber()
	runtime.GOMAXPROCS(1)
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger.Out = logFile
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		rn.Increment()
		logger.Infof("avg: %d", rn.Avg(now))
		response := fmt.Sprintf("timestamp: %d\n", now.Unix())
		fmt.Fprintf(w, response)
	})

	http.ListenAndServe(":8000", nil)
}
