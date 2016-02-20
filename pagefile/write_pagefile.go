package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/trpedersen/page"
)

func makePage() *page.Page {
	page := page.NewPage()
	recLen := int16(10)
	for i := 0; page.GetFreeSpace() > recLen; i++ {
		record := make([]byte, recLen)
		for j := int16(0); j < recLen; j++ {
			record[j] = byte(i)
		}
		if _, err := page.AddRecord(record); err != nil {
			panic(err)
		}
	}
	return page
}

func main() {
	path := flag.String("path", "pagefile", "path+filename")
	count := flag.Int("count", 10, "number of pages, default 10")

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	file, err := os.OpenFile(*path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666) // todo: path for topic
	defer file.Close()

	if err != nil {
		log.Printf("Error opening file, path: %s, err: %s\n", *path, err)
		return
	}

	for i := 0; i < *count; i++ {
		page := makePage()
		bytes, err := page.MarshalBinary()
		if err != nil {
			panic(err)
		}
		file.Write(bytes)
	}

}
