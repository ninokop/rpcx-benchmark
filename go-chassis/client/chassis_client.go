package main

import (
	"context"
	"flag"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ServiceComb/go-chassis"
	"github.com/ServiceComb/go-chassis/core"
	"github.com/gogo/protobuf/proto"
	"github.com/montanaflynn/stats"
	"github.com/rpcx-ecosystem/rpcx-benchmark/go-chassis/schemas"
)

var concurrency = flag.Int("c", 1, "concurrency")
var total = flag.Int("n", 1, "total requests for all clients")
var host = flag.String("s", "RPCServer", "listened ip and port")

func main() {
	flag.Parse()
	n := *concurrency
	m := *total / n

	log.Printf("concurrency: %d\nrequests per client: %d\n\n", n, m)

	args := prepareArgs()
	b, _ := proto.Marshal(args)
	log.Printf("message size: %d bytes\n\n", len(b))

	//it contains warmup time but we can ignore it
	if err := chassis.Init(); err != nil {
		log.Fatalf("chassis init failed, %v", err)
		return
	}

	// for {
	// 	log.Printf("===========================")
	var wg sync.WaitGroup
	wg.Add(n * m)

	var trans uint64
	var transOK uint64

	d := make([][]int64, n, n)
	totalT := time.Now().UnixNano()
	for i := 0; i < n; i++ {
		dt := make([]int64, 0, m)
		d = append(d, dt)

		go func(i int) {
			c := core.NewRPCInvoker()
			//warmup
			for j := 0; j < 5; j++ {
				reply := &schemas.BenchmarkMessage{}
				c.Invoke(context.Background(), *host, "Hello", "Say", args, reply)
			}

			for j := 0; j < m; j++ {
				t := time.Now().UnixNano()
				reply := &schemas.BenchmarkMessage{}
				err := c.Invoke(context.Background(), *host, "Hello", "Say", args, reply)
				t = time.Now().UnixNano() - t

				d[i] = append(d[i], t)
				if err == nil && reply.Field1 == "OK" {
					atomic.AddUint64(&transOK, 1)
				}

				atomic.AddUint64(&trans, 1)
				wg.Done()
			}
		}(i)
	}

	wg.Wait()
	totalT = time.Now().UnixNano() - totalT
	totalT = totalT / 1000000
	log.Printf("took %d ms for %d requests\n", totalT, n*m)

	totalD := make([]int64, 0, n*m)
	for _, k := range d {
		totalD = append(totalD, k...)
	}
	totalD2 := make([]float64, 0, n*m)
	for _, k := range totalD {
		totalD2 = append(totalD2, float64(k))
	}

	mean, _ := stats.Mean(totalD2)
	median, _ := stats.Median(totalD2)
	max, _ := stats.Max(totalD2)
	min, _ := stats.Min(totalD2)
	p99, _ := stats.Percentile(totalD2, 99.9)

	log.Printf("sent     requests    : %d\n", n*m)
	log.Printf("received requests    : %d\n", atomic.LoadUint64(&trans))
	log.Printf("received requests_OK : %d\n", atomic.LoadUint64(&transOK))
	log.Printf("throughput  (TPS)    : %d\n", int64(n*m)*1000/totalT)
	log.Printf("mean: %.f ns, median: %.f ns, max: %.f ns, min: %.f ns, p99: %.f ns\n", mean, median, max, min, p99)
	log.Printf("mean: %d ms, median: %d ms, max: %d ms, min: %d ms, p99: %d ms\n", int64(mean/1000000), int64(median/1000000), int64(max/1000000), int64(min/1000000), int64(p99/1000000))
	// time.Sleep(2 * time.Second)
	// }
	// select {}
}

func prepareArgs() *schemas.BenchmarkMessage {
	b := true
	var i int32 = 100000
	var i64 int64 = 100000
	var s = "许多往事在眼前一幕一幕，变的那麼模糊"

	var args schemas.BenchmarkMessage

	v := reflect.ValueOf(&args).Elem()
	num := v.NumField()
	for k := 0; k < num; k++ {
		field := v.Field(k)
		if field.Type().Kind() == reflect.Ptr {
			switch v.Field(k).Type().Elem().Kind() {
			case reflect.Int, reflect.Int32:
				field.Set(reflect.ValueOf(&i))
			case reflect.Int64:
				field.Set(reflect.ValueOf(&i64))
			case reflect.Bool:
				field.Set(reflect.ValueOf(&b))
			case reflect.String:
				field.Set(reflect.ValueOf(&s))
			}
		} else {
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				field.SetInt(100000)
			case reflect.Bool:
				field.SetBool(true)
			case reflect.String:
				field.SetString(s)
			}
		}

	}
	return &args
}
