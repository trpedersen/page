package page

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	PAGE_SIZE = uint16(4096)
)

type Page struct {
	bytes            []byte // bytes contains all the data in the page, including header fields, free space and the records
	recordCountBytes []byte
	freePointerBytes []byte
	recordTableBytes []byte
}

func NewPage() *Page {
	page := &Page{
		bytes: make([]byte, PAGE_SIZE),
	}
	page.recordCountBytes = page.bytes[0:2]
	page.setRecordCount(0)
	page.freePointerBytes = page.bytes[2:4]
	//fmt.Printf("capacity fp %d\n", cap(page.freePointerBytes))
	page.setFreePointer(PAGE_SIZE)
	page.recordTableBytes = page.bytes[4:4]
	//fmt.Printf("rt cap %d len %d\n", cap(page.recordTableBytes), len(page.recordTableBytes))
	return page
}

func (page *Page) MarshalBinary() ([]byte, error) {
	result := make([]byte, len(page.bytes))
	copy(result, page.bytes)
	return result, nil
}

func (page *Page) UnmarshalBinary(data []byte) error {
	page.bytes = make([]byte, len(data))
	copy(page.bytes, data)
	page.recordCountBytes = page.bytes[0:2]
	page.freePointerBytes = page.bytes[2:4]
	page.recordTableBytes = page.bytes[4 : page.GetRecordCount()*4]
	return nil
}

func (page *Page) setRecordCount(recordCount uint16) {
	binary.LittleEndian.PutUint16(page.recordCountBytes, uint16(recordCount))
	return
}

func (page *Page) GetRecordCount() uint16 {
	return binary.LittleEndian.Uint16(page.recordCountBytes)
}

func (page *Page) setFreePointer(freePointer uint16) {
	binary.LittleEndian.PutUint16(page.freePointerBytes, freePointer)
}

func (page *Page) getFreePointer() uint16 {
	return binary.LittleEndian.Uint16(page.freePointerBytes)
}

func (page *Page) setRecordTable(recordNumber uint16, offset uint16, recLen uint16) error {
	tableOffset := recordNumber * 4
	// resize recordTable
	len := len(page.recordTableBytes)
	page.recordTableBytes = page.recordTableBytes[0 : len+4] // add two more uint16 == 4 bytes
	binary.LittleEndian.PutUint16(page.recordTableBytes[tableOffset:tableOffset+2], offset)
	binary.LittleEndian.PutUint16(page.recordTableBytes[tableOffset+2:tableOffset+4], recLen)
	return nil
}

// Return the amount of free space available to store a record (inclusive of any header field)
func (page *Page) GetFreeSpace() int16 {
	return int16(page.getFreePointer()) - 2 - int16(page.GetRecordCount()*4) - 4 // free pointer - 4 bytes header fields - #records * 4 bytes per table entry - another table entry
}

func (page *Page) AddRecord(record []byte) (uint16, error) {
	recLen := uint16(len(record))
	//fmt.Println(page.GetFreeSpace())
	if int16(recLen) > page.GetFreeSpace() {
		return 0, errors.New("Record length exceeds free space")
	}
	freePointer := page.getFreePointer()

	offset := freePointer - recLen
	copy(page.bytes[offset:freePointer], record)
	page.setFreePointer(offset)
	recordNumber := page.GetRecordCount() // NB 0-based
	page.setRecordCount(recordNumber + 1)
	page.setRecordTable(recordNumber, offset, recLen)
	return recordNumber, nil
}

func (page *Page) GetRecord(recordNumber uint16) ([]byte, error) {
	// recordNumber is 0 based
	recordCount := page.GetRecordCount()
	if recordNumber+1 > recordCount {
		return nil, errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, recordCount))
	}
	tableOffset := recordNumber * 4
	offset := binary.LittleEndian.Uint16(page.recordTableBytes[tableOffset : tableOffset+2])
	len := binary.LittleEndian.Uint16(page.recordTableBytes[tableOffset+2 : tableOffset+4])
	record := make([]byte, len)
	copy(record, page.bytes[offset:offset+len])
	return record, nil
}
