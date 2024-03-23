package ir

import (
	"fmt"
	"encoding/json"
	"testing"
	"bufio"
	"os"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/y-scope/clp-ffi-go/ffi"
)

func TestKVWriter(t *testing.T) {
	fmt.Println("Start.")
	writer, _ := NewKVWriter()

	file, err := os.Open("/Users/lzh/go-ffi/clp-ffi-go/test.json")
	if err != nil {
		t.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	outPath := "/Users/lzh/go-ffi/clp-ffi-go/test.json.clp"
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("Error opening or creating file: %v", err)
	}
	defer outFile.Close()
	writer.WriteTo(outFile)

	// Step 2: Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonLine := scanner.Text()

		// Step 3: Unmarshal the JSON line into an interface{}
		var data interface{}
		err := json.Unmarshal([]byte(jsonLine), &data)
		if err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}

		// Step 4: Marshal the data into MessagePack
		msgPackBytes, err := msgpack.Marshal(data)
		if err != nil {
			t.Fatalf("Error marshalling to MessagePack: %v", err)
		}
		if 0 == len(msgPackBytes) {
			t.Fatalf("Encoded msgpack bytes empty: %s", jsonLine)
		}
		event := ffi.LogEvent{
			BinaryRecord: msgPackBytes,
		}
		numBytes, err := writer.Write(event)
		if err != nil {
			t.Fatalf("Failed to serialize msg: %v", err)
		}
		if 0 == numBytes {
			t.Fatalf("0 bytes written: %s", jsonLine)
		}
		writer.WriteTo(outFile)
	}

	// Check for errors that may have occurred during scanning
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	writer.CloseTo(outFile)	
    fmt.Println("Done.")
}