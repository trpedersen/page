package page

import (
	"bytes"
	"testing"
)

func TestHeaderFields(t *testing.T) {

	page := NewPage()

	page.setRecordCount(99)
	if recordCount := page.GetRecordCount(); recordCount != 99 {
		t.Errorf("page.GetRecordCount expected 99, got %d", recordCount)
	}

	if freePointer := page.getFreePointer(); freePointer != PAGE_SIZE {
		t.Errorf("page.GetFreePointer expected %d, got %d", PAGE_SIZE, freePointer)
	}

	page.setFreePointer(312)
	if freePointer := page.getFreePointer(); freePointer != 312 {
		t.Errorf("page.GetFreePointer expected %d, got %d", 312, freePointer)
	}

	//fmt.Printf("%t\n", page)
}

func TestFillPage(t *testing.T) {
	page := NewPage()
	recLen := int16(10)
	for i := 0; page.GetFreeSpace() > recLen; i++ {
		record := make([]byte, recLen)
		for j := int16(0); j < recLen; j++ {
			record[j] = byte(i)
		}
		if recordNumber, err := page.AddRecord(record); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		} else if read, err := page.GetRecord(recordNumber); err != nil {
			t.Fatalf("page.GetRecord, err: %s", err)
		} else if bytes.Compare(record, read) != 0 {
			t.Errorf("bytes.Compare: expected %t, got %t", record, read)
			break
		}
	}
}

func TestMarshalBinary(t *testing.T) {
	page1 := NewPage()
	recLen := int16(10)
	for i := 0; page1.GetFreeSpace() > recLen; i++ {
		record := make([]byte, recLen)
		for j := int16(0); j < recLen; j++ {
			record[j] = byte(i)
		}
		if _, err := page1.AddRecord(record); err != nil {
			t.Fatalf("page.AddRecord, err: %s", err)
		}
	}

	pageBytes, err := page1.MarshalBinary()
	if err != nil {
		t.Fatalf("page1.MarshalBinary, err: %s", err)
	}

	page2 := NewPage()

	err = page2.UnmarshalBinary(pageBytes)
	if err != nil {
		t.Fatalf("page2.UnmarshalBinary, err: %s", err)
	}

	if page1.GetRecordCount() != page2.GetRecordCount() {
		t.Errorf(".GetRecordCount, expecting: %d, got: %d", page1.GetRecordCount(), page2.GetRecordCount())
	}

	for i := uint16(0); i < page1.GetRecordCount(); i++ {
		record1, err := page1.GetRecord(i)
		if err != nil {
			t.Fatalf("page1.GetRecord, err: %s", err)
		}
		record2, err := page2.GetRecord(i)
		if err != nil {
			t.Fatalf("page2.GetRecord, err: %s", err)
		}
		if bytes.Compare(record1, record2) != 0 {
			t.Errorf("bytes.Compare, expecting: %s, got: %s", record1, record2)
		}
	}

}

func BenchmarkFillPage(b *testing.B) {
	page := NewPage()
	recLen := int16(10)
	for k := 0; k < b.N; k++ {
		for i := 0; page.GetFreeSpace() > recLen; i++ {
			record := make([]byte, recLen)
			for j := int16(0); j < recLen; j++ {
				record[j] = byte(i)
			}
			if recordNumber, err := page.AddRecord(record); err != nil {
				b.Fatalf("page.AddRecord, err: %s", err)
			} else if read, err := page.GetRecord(recordNumber); err != nil {
				b.Fatalf("page.GetRecord, err: %s", err)
			} else if bytes.Compare(record, read) != 0 {
				b.Errorf("bytes.Compare: expected %t, got %t", record, read)
				break
			}

		}
	}
}
