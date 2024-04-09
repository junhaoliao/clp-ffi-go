package ir

/*
#include <ffi_go/defs.h>
#include <ffi_go/ir/serializer.h>
*/
import "C"

import (
	"unsafe"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/y-scope/clp-ffi-go/ffi"
)

// A Serializer exports functions to serialize log events into a CLP IR byte
// stream. Serialization functions only return views (slices) of IR bytes,
// leaving their use to the user. Each Serializer owns its own unique underlying
// memory for the views it produces/returns. This memory is reused for each
// view, so to persist the contents the memory must be copied into another
// object. Close must be called to free the underlying memory and failure to do
// so will result in a memory leak.
type Serializer interface {
	SerializeUnstructuredLogEvent(event ffi.UnstructuredLogEvent) (BufView, error)
	SerializeStructuredLogEvent(event ffi.StructuredLogEvent) (BufView, error)
	TimestampInfo() TimestampInfo
	Close() error
}

type UnstructuredLogEventJson struct {
	LogMessage string          `json:"message"`
	Timestamp  ffi.EpochTimeMs `json:"timestamp"`
}

// EightByteSerializer creates and returns a new Serializer that writes eight
// byte encoded CLP IR and serializes a IR preamble into a BufView using it. On
// error returns:
//   - nil Serializer
//   - nil BufView
//   - [IrError] error: CLP failed to successfully serialize
func EightByteSerializer(
	tsPattern string,
	tsPatternSyntax string,
	timeZoneId string,
) (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := eightByteSerializer{
		commonSerializer{TimestampInfo{tsPattern, tsPatternSyntax, timeZoneId}, nil},
	}
	if err := IrError(C.ir_serializer_serialize_eight_byte_preamble(
		newCStringView(tsPattern),
		newCStringView(tsPatternSyntax),
		newCStringView(timeZoneId),
		&irs.cptr,
		&irView,
	)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// FourByteSerializer creates and returns a new Serializer that writes four byte
// encoded CLP IR and serializes a IR preamble into a BufView using it. On error
// returns:
//   - nil Serializer
//   - nil BufView
//   - [IrError] error: CLP failed to successfully serialize
func FourByteSerializer(
	tsPattern string,
	tsPatternSyntax string,
	timeZoneId string,
	referenceTs ffi.EpochTimeMs,
) (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := fourByteSerializer{
		commonSerializer{TimestampInfo{tsPattern, tsPatternSyntax, timeZoneId}, nil},
		referenceTs,
	}
	if err := IrError(C.ir_serializer_serialize_four_byte_preamble(
		newCStringView(tsPattern),
		newCStringView(tsPatternSyntax),
		newCStringView(timeZoneId),
		C.int64_t(referenceTs),
		&irs.cptr,
		&irView,
	)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// TODO: complete doc str.
func KeyValuePairSerializer() (Serializer, BufView, error) {
	var irView C.ByteSpan
	irs := keyValuePairSerializer{commonSerializer{TimestampInfo{"", "", ""}, nil}}
	if err := IrError(C.ir_serializer_serialize_kv_preamble(&irs.cptr, &irView)); Success != err {
		return nil, nil, err
	}
	return &irs, unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

// commonSerializer contains fields common to all types of CLP IR encoding.
// TimestampInfo stores information common to all timestamps found in the IR.
// cptr holds a reference to the underlying C++ objected used as backing storage
// for the Views returned by the serializer. Close must be called to free this
// underlying memory and failure to do so will result in a memory leak.
type commonSerializer struct {
	tsInfo TimestampInfo
	cptr   unsafe.Pointer
}

// Returns the TimestampInfo of the Serializer.
func (self commonSerializer) TimestampInfo() TimestampInfo {
	return self.tsInfo
}

type eightByteSerializer struct {
	commonSerializer
}

// SerializeUnstructuredLogEvent attempts to serialize the log event, event,
// into a eight byte encoded CLP IR byte stream. On error returns:
//   - a nil BufView
//   - [IrError] based on the failure of the Cgo call
func (self *eightByteSerializer) SerializeUnstructuredLogEvent(
	event ffi.UnstructuredLogEvent,
) (BufView, error) {
	return serializeUnstructuredLogEvent(self, event)
}

func (self *eightByteSerializer) SerializeStructuredLogEvent(
	event ffi.StructuredLogEvent,
) (BufView, error) {
	return serializeStructuredLogEvent(self, event)
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (self *eightByteSerializer) Close() error {
	if nil != self.cptr {
		C.ir_serializer_close(self.cptr)
		self.cptr = nil
	}
	return nil
}

// fourByteSerializer contains both a common CLP IR serializer and stores the
// previously seen log event's timestamp. The previous timestamp is necessary to
// calculate the current timestamp as four byte encoding only encodes the
// timestamp delta between the current log event and the previous.
type fourByteSerializer struct {
	commonSerializer
	prevTimestamp ffi.EpochTimeMs
}

// SerializeUnstructuredLogEvent attempts to serialize the log event, event,
// into a four byte encoded CLP IR byte stream. On error returns:
//   - nil BufView
//   - [IrError] based on the failure of the Cgo call
func (self *fourByteSerializer) SerializeUnstructuredLogEvent(
	event ffi.UnstructuredLogEvent,
) (BufView, error) {
	return serializeUnstructuredLogEvent(self, event)
}

func (self *fourByteSerializer) SerializeStructuredLogEvent(
	event ffi.StructuredLogEvent,
) (BufView, error) {
	return serializeStructuredLogEvent(self, event)
}

// Close will delete the underlying C++ allocated memory used by the
// deserializer. Failure to call Close will result in a memory leak.
func (self *fourByteSerializer) Close() error {
	if nil != self.cptr {
		C.ir_serializer_close(self.cptr)
		self.cptr = nil
	}
	return nil
}

// TODO: add doc str.
type keyValuePairSerializer struct {
	commonSerializer
}

// TODO: add doc str.
func (self *keyValuePairSerializer) SerializeUnstructuredLogEvent(
	event ffi.UnstructuredLogEvent,
) (BufView, error) {
	return serializeUnstructuredLogEvent(self, event)
}

func (self *keyValuePairSerializer) SerializeStructuredLogEvent(
	event ffi.StructuredLogEvent,
) (BufView, error) {
	return serializeStructuredLogEvent(self, event)
}

// TODO: add doc str.
func (self *keyValuePairSerializer) Close() error {
	if nil != self.cptr {
		C.ir_kv_serializer_close(self.cptr)
		self.cptr = nil
	}
	return nil
}

func serializeUnstructuredLogEvent(
	serializer Serializer,
	event ffi.UnstructuredLogEvent,
) (BufView, error) {
	var irView C.ByteSpan
	var err error
	switch irs := serializer.(type) {
	case *eightByteSerializer:
		err = IrError(C.ir_serializer_serialize_eight_byte_log_event(
			newCStringView(event.LogMessage),
			C.int64_t(event.Timestamp),
			irs.cptr,
			&irView,
		))
	case *fourByteSerializer:
		err = IrError(C.ir_serializer_serialize_four_byte_log_event(
			newCStringView(event.LogMessage),
			C.int64_t(event.Timestamp-irs.prevTimestamp),
			irs.cptr,
			&irView,
		))
		if Success == err {
			irs.prevTimestamp = event.Timestamp
		}
	case *keyValuePairSerializer:
		structured_event := map[string]interface{}{
			"timestamp": event.Timestamp,
			"message":   event.LogMessage,
		}
		msgpackBytes, msgpackErr := msgpack.Marshal(structured_event)
		if nil != msgpackErr {
			return nil, msgpackErr
		}
		err = IrError(C.ir_serializer_serialize_kv_log_event(
			newCByteSpan(msgpackBytes),
			irs.cptr,
			&irView,
		))
	}
	if Success != err {
		return nil, err
	}
	return unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}

func serializeStructuredLogEvent(
	serializer Serializer,
	event ffi.StructuredLogEvent,
) (BufView, error) {
	var irView C.ByteSpan
	var err error
	switch irs := serializer.(type) {
	case *eightByteSerializer, *fourByteSerializer:
		err = UnsupportedVersion
	case *keyValuePairSerializer:
		err = IrError(C.ir_serializer_serialize_kv_log_event(
			newCByteSpan(event.MsgpackRecord),
			irs.cptr,
			&irView,
		))
	}
	if Success != err {
		return nil, err
	}
	return unsafe.Slice((*byte)(irView.m_data), irView.m_size), nil
}
