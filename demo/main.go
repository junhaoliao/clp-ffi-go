package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/y-scope/clp-ffi-go/ffi"
	"github.com/y-scope/clp-ffi-go/ir"
)

func test_structured(file_path string) {
	writer, _ := ir.NewKeyValuePairWriter()

	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()

	outPath := file_path + ".clp"
	fmt.Println("Serialized IR to path: ", outPath)
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file: ", err)
		return
	}
	defer outFile.Close()
	writer.WriteTo(outFile)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonLine := scanner.Text()

		var data interface{}
		err := json.Unmarshal([]byte(jsonLine), &data)
		if err != nil {
			fmt.Println("Error unmarshalling JSON: ", err)
			return
		}

		msgPackBytes, err := msgpack.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling to MessagePack: ", err)
			return
		}
		if 0 == len(msgPackBytes) {
			fmt.Println("Encoded msgpack bytes empty: ", jsonLine)
			return
		}
		event := ffi.StructuredLogEvent{MsgpackRecord: msgPackBytes}
		numBytes, err := writer.WriteStructured(event)
		if err != nil {
			fmt.Println("Failed to serialize msg: ", err)
			return
		}
		if 0 == numBytes {
			fmt.Println("0 bytes written: ", jsonLine)
			return
		}
		writer.WriteTo(outFile)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}

	writer.CloseTo(outFile)
}

func test_unstructured() {
	writer, _ := ir.NewKeyValuePairWriter()

	outPath := "/Users/lzh/go-ffi/clp-ffi-go/unstructured.clp"
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file: ", err)
	}
	defer outFile.Close()
	writer.WriteTo(outFile)

	for i := 0; i < 10; i++ {
		event := ffi.UnstructuredLogEvent{
			LogMessage: fmt.Sprintf("This is message: %d", i),
			Timestamp:  ffi.EpochTimeMs(time.Now().UnixMilli()),
		}
		numBytes, err := writer.WriteUnstructured(event)
		if err != nil {
			fmt.Println("Failed to serialize msg: ", err)
		}
		if 0 == numBytes {
			fmt.Println("0 bytes written.")
		}
		writer.WriteTo(outFile)
	}

	writer.CloseTo(outFile)

}

func main() {
	if 2 != len(os.Args) {
		fmt.Println("Please give the path of the JSON file for serialization.")
		return
	}
	file_path := os.Args[1]
	// test_unstructured()
	test_structured(file_path)
}
