package main

import (
  "fmt"
  "github.com/docker/docker/pkg/archive"
  "io"
  "os"
  "sync"
)

func main() {
  var wg sync.WaitGroup

  for i := 0; i < 10; i++ {
    wg.Add(1)
    f, err := os.Open("data.gz")
    if err != nil {
      fmt.Println(err.Error())
      os.Exit(1)
    }
    go decompress(f, &wg)
  }
  fmt.Println("Waiting...")
  wg.Wait()
  fmt.Println("Complete")
}

func decompress(f io.Reader, wg *sync.WaitGroup) {
  r, _ := archive.DecompressStream(f)
  defer r.Close()
  defer wg.Done()
  fmt.Println("Done")
}
