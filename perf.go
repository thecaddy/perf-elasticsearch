package main

import (
  "fmt"
  "log"
  "time"
  "flag"
  "net/http"
  "encoding/json"
  "math/rand"
  "sync"
  "github.com/fatih/color"
  elastigo "github.com/shutej/elastigo/lib"
)

var b = [...]string{"Penn", "Teller", "Left", "right", "orange", "black", "street",
  "giant", "treehouse", "bland", "yard", "grove", "dance", "lights", "chicken", "farm land",
  "land house", "tyler", "high", "view", "lake", "ocean"}

func Log(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    cyan := color.New(color.FgCyan).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()
    t1 := time.Now()
    fmt.Printf("%s --> %s %s\n", time.Now().UTC().Format(time.RFC3339), cyan(r.Method), r.URL)

    handler.ServeHTTP(w, r)

    t2 := time.Now()

    s := t2.Sub(t1).String()
    dur,_ := time.ParseDuration("300ms")
    if t2.Sub(t1) > dur {
      s = red(s)
    }else{
      s = green(s)
    }
    fmt.Printf("%s <-- %s %s %s\n", time.Now().UTC().Format(time.RFC3339),
     cyan(r.Method), r.URL, s)
  })
}

var (
  host *string = flag.String("host", "addrhere", "Elasticsearch Host")
  host2 *string = flag.String("host2", "addrhere", "Elasticsearch Host")
)

type Performance struct {
  MaxLatency  int64
  MaxEsearch  int
  MinLatency  int64
  MinEsearch  int
  AvgLatency  int64
  AvgEsearch  int
}
var lock sync.Mutex
var perf = Performance{}
var Threads int = 252
var Requests int = 100
var c = make(chan int, Threads)

func logPerf(logPerf Performance){
  lock.Lock()
  if logPerf.MaxLatency > perf.MaxLatency {
    perf.MaxLatency = logPerf.MaxLatency
  }
  if logPerf.MaxEsearch > perf.MaxEsearch {
    perf.MaxEsearch = logPerf.MaxEsearch
  }
  if logPerf.MinLatency < perf.MinLatency {
    perf.MinLatency = logPerf.MinLatency
  }
  if logPerf.MinEsearch < perf.MinEsearch {
    perf.MinEsearch = logPerf.MinEsearch
  }
  if perf.MinEsearch == 0 {
    perf.MinEsearch = logPerf.MinEsearch
    perf.AvgEsearch = logPerf.AvgEsearch
  }else{
    perf.AvgEsearch = (perf.AvgEsearch+logPerf.AvgEsearch)/2
  }
  if perf.MinLatency == 0 {
    perf.MinLatency = logPerf.MinLatency
    perf.AvgLatency = logPerf.AvgLatency
  }else{
    perf.AvgLatency = (perf.AvgLatency+logPerf.AvgLatency)/2
  }
  lock.Unlock()
}

func hitEsearch(){
  perfMon := Performance{}
  for i := 0; i < Requests; i++ {
    e := elastigo.NewConn()
    log.SetFlags(log.LstdFlags)
    flag.Parse()

    //get random hosts
    if randInt(0,1) > 0 {
      e.Domain = *host
    }else{
      e.Domain = *host2
    }

    //get random search word
    search := b[randInt(0,len(b))]

    t1 := time.Now()
    out,err := elastigo.Search("stuff").Type("thing").
      Size("100").Search(search).Result(e)
    _,err = json.Marshal(out)
    t2 := time.Now()

    
    duration := t2.Sub(t1)
    //s := duration.String()
    var castToInt64 int64 = duration.Nanoseconds() / 1e6

    if castToInt64 > perfMon.MaxLatency {
      perfMon.MaxLatency = castToInt64
    }
    if out.Took > perfMon.MaxEsearch {
      perfMon.MaxEsearch = out.Took
    }
    if castToInt64 < perfMon.MinLatency {
      perfMon.MinLatency = castToInt64
    }
    if out.Took < perfMon.MinEsearch {
      perfMon.MinEsearch = out.Took
    }
    if perfMon.MinEsearch == 0 {
      perfMon.MinEsearch = out.Took
      perfMon.AvgEsearch = out.Took
    }else{
      perfMon.AvgEsearch = (perfMon.AvgEsearch+out.Took)/2
    }
    if perfMon.MinLatency == 0 {
      perfMon.MinLatency = castToInt64
      perfMon.AvgLatency = castToInt64
    }else{
      perfMon.AvgLatency = (perfMon.AvgLatency+castToInt64)/2
    }



    checkErr(err, "search err:")
  }
  fmt.Printf("Thread Log: %v\n", perfMon)
  logPerf(perfMon)
  c<-1
}
func handler(w http.ResponseWriter, r *http.Request) {

  fmt.Fprintf(w, "Dang Man")
  //fmt.Fprintf(w, "Hi there, this is sweet %s%s!", r.Host, r.URL.Path[1:])
}
func apiResponse(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hi this is so cool, I am in like with %s%s!", r.Host, r.URL)
}
func put(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "PUT!")
}
func delete(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "DELETE!")
}

func main() {
  red := color.New(color.FgRed).SprintFunc()
  green := color.New(color.FgGreen).SprintFunc()
  t1 := time.Now()
  fmt.Printf("START: %s\n", time.Now().UTC().Format(time.RFC3339))

  for i := 0; i < Threads; i++ {
    go hitEsearch() 
  }
  for i := 0; i < Threads; i++ {
    <-c    // wait for one task to complete
  }
  fmt.Printf("Complete: %v\n", perf)

  t2 := time.Now()

  duration := t2.Sub(t1)
  s := duration.String()
  dur,_ := time.ParseDuration("300ms")
  if t2.Sub(t1) > dur {
    s = red(s)
  }else{
    s = green(s)
  }
  total := Threads * Requests
  num := int(total) / int(duration.Seconds())
  fmt.Printf("END: %s -- %s %d Requests %d Requests/Second\n",
  time.Now().UTC().Format(time.RFC3339),s, total, num)
}

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

func randInt(min int, max int) int {
    return min + rand.Intn(max-min)
}