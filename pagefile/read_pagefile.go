package main

import (
	"flag"
	"fmt"
	"github.com/trpedersen/io"
	"github.com/trpedersen/page"
	"log"
	"os"
)

const ()

func main() {
	scanner := io.NewScanner(os.Stdin)
	path := flag.String("path", "pagefile", "path+filename")
	flag.Parse()
	file, err := os.Open(*path) // todo: path for topic
	if err != nil {
		log.Printf("Error opening file, path: %s, err: %s\n", *path, err)
		return
	}

	var pageNumber int
	for {
		//fmt.Print("\noffset length: ")
		fmt.Print("page number: ")

		//offset, err = scanner.ReadLong()
		pageNumber, err = scanner.ReadInt()
		if err != nil {
			fmt.Println("Enter an int\n")
			continue
		}

		bytes := make([]byte, page.PAGE_SIZE)
		var n int
		offset := int64(pageNumber) * int64(page.PAGE_SIZE)
		fmt.Printf("offset: %d\n", offset)
		if n, err = file.ReadAt(bytes, offset); (err != nil) && (n == 0) {
			fmt.Println(err)
		} else {
			page := page.NewPage()
			if err = page.UnmarshalBinary(bytes); err != nil {
				fmt.Println(err)
			} else {
				//fmt.Println(page)
				fmt.Printf("record count: %d\n", page.GetRecordCount())
				for i := uint16(0); i < page.GetRecordCount(); i++ {
					var record []byte
					record, err = page.GetRecord(i)
					if err != nil {
						fmt.Printf("page.GetRecord, err: %s\n", err)
					} else {
						fmt.Printf("record %d: %t\n", i, record)
					}
				}
			}
		}
	}
}
