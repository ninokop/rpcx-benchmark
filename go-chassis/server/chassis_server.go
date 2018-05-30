package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/ServiceComb/go-chassis"
	"github.com/ServiceComb/go-chassis/core/server"
	"github.com/rpcx-ecosystem/rpcx-benchmark/go-chassis/schemas"
)

func main() {
	chassis.RegisterSchema("highway", &Hello{}, server.WithSchemaID("Hello"))
	if err := chassis.Init(); err != nil {
		return
	}
	go func() {
		log.Println(http.ListenAndServe(*debugAddr, nil))
	}()

	chassis.Run()
}

var (
	// host       = flag.String("s", "127.0.0.1:8972", "listened ip and port")
	// cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	delay     = flag.Duration("delay", 0, "delay to mock business processing")
	debugAddr = flag.String("d", "127.0.0.1:6065", "server ip and port")
)

type Hello struct{}

func (h *Hello) Say(ctx context.Context, in *schemas.BenchmarkMessage) (*schemas.BenchmarkMessage, error) {
	s := "OK"
	var i int32 = 100
	in.Field1 = s
	in.Field2 = i
	if *delay > 0 {
		time.Sleep(*delay)
	} else {
		runtime.Gosched()
	}
	return in, nil
}
