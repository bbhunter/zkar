package serz

import (
	"encoding/binary"
	"fmt"
	"github.com/phith0n/zkar/commons"
	"github.com/samber/lo"
	"math"
)

var NoFieldError = fmt.Errorf("Oops!")
var SizeTable = map[string]int{
	"B": 1,
	"C": 2,
	"D": 8,
	"F": 4,
	"I": 4,
	"J": 8,
	"S": 2,
	"Z": 1,
}

type TCValue struct {
	TypeCode string
	Byte     byte    // byte in Java
	Char     uint16  // char in Java
	Double   float64 // double in Java
	Float    float32 // float in Java
	Integer  int32   // int in Java
	Long     int64   // long in Java
	Short    int16   // short in Java
	Boolean  bool    // bool in Java
	Object   Object  // object in Java
}

func (t *TCValue) ToBytes() []byte {
	if t.TypeCode == "L" || t.TypeCode == "[" {
		return t.Object.ToBytes()
	}

	switch t.TypeCode {
	case "B":
		return []byte{t.Byte}
	case "C":
		return commons.NumberToBytes(t.Char)
	case "D":
		return commons.NumberToBytes(t.Double)
	case "F":
		return commons.NumberToBytes(t.Float)
	case "I":
		return commons.NumberToBytes(t.Integer)
	case "J":
		return commons.NumberToBytes(t.Long)
	case "S":
		return commons.NumberToBytes(t.Short)
	case "Z":
		if t.Boolean {
			return []byte{0x01}
		} else {
			return []byte{0x00}
		}
	}

	return nil
}

func (t *TCValue) ToString() string {
	var b = commons.NewPrinter()
	switch t.TypeCode {
	case "B":
		b.Printf("(byte)%v - %s", t.Byte, commons.Hexify(t.Byte))
	case "C":
		b.Printf("(char)%v - %s", t.Char, commons.Hexify(t.Char))
	case "D":
		b.Printf("(double)%v - %s", t.Double, commons.Hexify(t.Double))
	case "F":
		b.Printf("(float)%v - %s", t.Float, commons.Hexify(t.Float))
	case "I":
		b.Printf("(integer)%v - %s", t.Integer, commons.Hexify(t.Integer))
	case "J":
		b.Printf("(long)%v - %s", t.Long, commons.Hexify(t.Long))
	case "S":
		b.Printf("(short)%v - %s", t.Short, commons.Hexify(t.Short))
	case "Z":
		b.Printf("(boolean)%v - %s", t.Boolean, commons.Hexify(t.Boolean))
	case "L", "[":
		b.Print(t.Object.ToString())
	}

	return b.String()
}

func (t *TCValue) Walk(callback WalkCallback) error {
	if t.TypeCode == "L" || t.TypeCode == "[" {
		if err := callback(t.Object); err != nil {
			return err
		}

		if err := t.Object.Walk(callback); err != nil {
			return err
		}
	}

	return nil
}

func readTCValue(stream *ObjectStream, typeCode string) (*TCValue, error) {
	if lo.Contains(PrimitiveTypecode, typeCode) {
		return readTCValueFromPrimitive(stream, typeCode)
	} else {
		return readTCValueFromObject(stream, typeCode)
	}
}

func readTCValueFromPrimitive(stream *ObjectStream, typeCode string) (*TCValue, error) {
	var bs []byte
	var err error

	var size = SizeTable[typeCode]
	bs, err = stream.ReadN(size)
	if err != nil {
		return nil, fmt.Errorf("read primitive field value failed on index %v", stream.CurrentIndex())
	}

	var fieldData = &TCValue{TypeCode: typeCode}
	switch typeCode {
	case "B": // byte
		fieldData.Byte = bs[0]
	case "C": // char
		fieldData.Char = binary.BigEndian.Uint16(bs)
	case "D": // double
		bits := binary.BigEndian.Uint64(bs)
		fieldData.Double = math.Float64frombits(bits)
	case "F": // float
		bits := binary.BigEndian.Uint32(bs)
		fieldData.Float = math.Float32frombits(bits)
	case "I": // int
		fieldData.Integer = int32(binary.BigEndian.Uint32(bs))
	case "J": // long
		fieldData.Long = int64(binary.BigEndian.Uint64(bs))
	case "S": // short
		fieldData.Short = int16(binary.BigEndian.Uint16(bs))
	case "Z": // boolean
		fieldData.Boolean = bs[0] != 0x00
	}

	return fieldData, nil
}

func readTCValueFromObject(stream *ObjectStream, typeCode string) (*TCValue, error) {
	flag, err := stream.PeekN(1)
	if err != nil {
		return nil, fmt.Errorf("read object field value failed on index %v", stream.CurrentIndex())
	}

	var fieldData = &TCValue{TypeCode: typeCode}
	switch flag[0] {
	case JAVA_TC_OBJECT:
		fieldData.Object, err = readTCObject(stream)
	case JAVA_TC_NULL:
		fieldData.Object = readTCNull(stream)
	case JAVA_TC_STRING:
		fieldData.Object, err = readTCString(stream)
	case JAVA_TC_REFERENCE:
		fieldData.Object, err = readTCReference(stream)
	case JAVA_TC_CLASS:
		fieldData.Object, err = readTCClass(stream)
	case JAVA_TC_ARRAY:
		fieldData.Object, err = readTCArray(stream)
	case JAVA_TC_ENUM:
		fieldData.Object, err = readTCEnum(stream)
	default:
		err = NoFieldError
	}

	if err != nil {
		return nil, err
	}

	return fieldData, nil
}

//func readTCValueFromArray(stream *ObjectStream, typeCode string) (*TCValue, error) {
//	flag, err := stream.PeekN(1)
//	if err != nil {
//		sugar.Error(err)
//		return nil, fmt.Errorf("read array field value failed on index %v", stream.CurrentIndex())
//	}
//
//	switch flag[0] {
//	case JAVA_TC_STRING:
//
//	}
//
//	return nil, nil
//}
