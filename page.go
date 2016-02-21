// Package page provides low-level types and functions for managing data pages and records.
// Writing pages to disk is left to other packages.
package page

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	PAGE_SIZE = uint16(4096) // PAGE_SIZE is typically the same as the filesystem blocksize
)

type pageid uint64

type Page struct {
	id          pageid // 0:8
	prevId      pageid // 8:16
	nextId      pageid // 16:24
	recordCount uint16 // 24:26
	freePointer uint16 // 26:28

	header      []byte
	recordTable []byte
	bytes       []byte // bytes contains all the data in the page, including header fields, free space and the records

}

// NewPage returns a new page of size page.PAGE_SIZE
func NewPage() *Page {
	page := &Page{
		bytes: make([]byte, PAGE_SIZE),
	}
	page.header = page.bytes[0:28]
	page.recordTable = page.bytes[28:28]
	page.freePointer = PAGE_SIZE
	return page
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// The page is encoded as a []byte PAGE_SIZE long, ready for serialisation.
func (page *Page) MarshalBinary() ([]byte, error) {

	binary.LittleEndian.PutUint64(page.header[0:8], uint64(page.id))
	binary.LittleEndian.PutUint64(page.header[8:16], uint64(page.prevId))
	binary.LittleEndian.PutUint64(page.header[16:24], uint64(page.nextId))
	binary.LittleEndian.PutUint16(page.header[24:26], page.recordCount)
	binary.LittleEndian.PutUint16(page.header[26:28], page.freePointer)

	return page.bytes, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
// PAGE_SIZE bytes are used to rehydrate the page.
func (page *Page) UnmarshalBinary(data []byte) error {

	page.bytes = make([]byte, PAGE_SIZE, PAGE_SIZE)
	copy(page.bytes, data)

	page.header = page.bytes[0:28]

	page.id = pageid(binary.LittleEndian.Uint64(page.header[0:8]))
	page.prevId = pageid(binary.LittleEndian.Uint64(page.header[8:16]))
	page.nextId = pageid(binary.LittleEndian.Uint64(page.header[16:24]))
	page.recordCount = binary.LittleEndian.Uint16(page.header[24:26])
	page.freePointer = binary.LittleEndian.Uint16(page.header[26:28])

	page.recordTable = page.bytes[28 : page.recordCount*4]

	return nil
}

// GetRecordCount returns the number of records held in page.
func (page *Page) GetRecordCount() uint16 {
	return page.recordCount
}

func (page *Page) setRecordTable(recordNumber uint16, offset uint16, recLen uint16) error {
	tableOffset := recordNumber * 4
	// resize recordTable
	len := len(page.recordTable)
	page.recordTable = page.recordTable[0 : len+4] // add two more uint16 == 4 bytes
	binary.LittleEndian.PutUint16(page.recordTable[tableOffset:tableOffset+2], offset)
	binary.LittleEndian.PutUint16(page.recordTable[tableOffset+2:tableOffset+4], recLen)
	return nil
}

// GetFreeSpace return the amount of free space available to store a record (inclusive of any header fields.)
func (page *Page) GetFreeSpace() uint16 {
	return uint16(page.freePointer) - 2 - uint16(page.recordCount*4) - 4 // free pointer - 4 bytes header fields - #records * 4 bytes per table entry - another table entry
}

// AddRecord adds record to page, using copy semantics.
// Returns an error if insufficient page free space.
func (page *Page) AddRecord(record []byte) (uint16, error) {
	recLen := uint16(len(record))
	if uint16(recLen) > page.GetFreeSpace() {
		return 0, errors.New("Record length exceeds free space")
	}

	offset := page.freePointer - recLen
	copy(page.bytes[offset:page.freePointer], record)
	page.freePointer = offset
	recordNumber := page.recordCount // NB 0-based
	page.recordCount += 1
	page.setRecordTable(recordNumber, offset, recLen)
	return recordNumber, nil
}

// GetRecord returns record specified by recordNumber.
// Note: record numbers are 0 based.
func (page *Page) GetRecord(recordNumber uint16) ([]byte, error) {
	// recordNumber is 0 based
	if recordNumber+1 > page.recordCount {
		return nil, errors.New(fmt.Sprintf("Invalid record number: %d, record count: %d", recordNumber, page.recordCount))
	}
	tableOffset := recordNumber * 4
	offset := binary.LittleEndian.Uint16(page.recordTable[tableOffset : tableOffset+2])
	len := binary.LittleEndian.Uint16(page.recordTable[tableOffset+2 : tableOffset+4])
	record := make([]byte, len, len)
	copy(record, page.bytes[offset:offset+len])
	return record, nil
}
