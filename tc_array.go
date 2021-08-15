package zkar

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type TCArray struct {
	TypeCode     byte
	ClassPointer *TCClassPointer
	ArrayData    []*TCValue
	Handler uint32
}

func (t *TCArray) ToBytes() []byte {
	var bs = []byte{JAVA_TC_ARRAY}
	bs = append(bs, t.ClassPointer.ToBytes()...)
	bs = append(bs, NumberToBytes(uint32(len(t.ArrayData)))...)
	for _, value := range t.ArrayData {
		bs = append(bs, value.ToBytes()...)
	}

	return bs
}

func (t *TCArray) ToString() string {
	var b = NewPrinter()
	b.Printf("TC_ARRAY - %s", Hexify(JAVA_TC_ARRAY))
	b.IncreaseIndent()
	b.Print(t.ClassPointer.ToString())
	b.Printf("@Handler - %v", t.Handler)
	b.Printf("@ArraySize - %d - %s", len(t.ArrayData), Hexify(uint32(len(t.ArrayData))))
	b.Printf("[]Values")
	b.IncreaseIndent()


	var className = ""
	if t.ClassPointer.Flag == JAVA_TC_CLASSDESC {
		className = t.ClassPointer.NormalClassDesc.ClassName.Data
	} else if t.ClassPointer.Flag == JAVA_TC_REFERENCE && t.ClassPointer.Reference.Flag == JAVA_TC_CLASSDESC {
		className = t.ClassPointer.Reference.NormalClassDesc.ClassName.Data
	}
	if className == "[B" {
		b.Print(t.DumpByteArray(t.ArrayData))
		return b.String()
	}

	for index, data := range t.ArrayData {
		b.Printf("Index %d", index)
		b.IncreaseIndent()
		b.Print(data.ToString())
		b.DecreaseIndent()
	}

	return b.String()
}

func (t *TCArray) DumpByteArray(arr []*TCValue) string {
	var b strings.Builder
	var dumper = hex.Dumper(&b)
	for _, v := range arr {
		_, _ = dumper.Write([]byte{v.Byte})
	}
	dumper.Close()
	return b.String()
}

func readTCArray(stream *ObjectStream) (*TCArray, error) {
	var array = new(TCArray)
	var err error

	_, _ = stream.ReadN(1)
	array.ClassPointer, err = readTCClassPointer(stream)
	if err != nil {
		return nil, err
	}

	stream.AddReference(array)
	bs, err := stream.ReadN(4)
	if err != nil {
		return nil, fmt.Errorf("read JAVA_TC_ARRAY object failed on index %v", stream.CurrentIndex())
	}

	var className string
	if array.ClassPointer.Flag == JAVA_TC_NULL || array.ClassPointer.Flag == JAVA_TC_PROXYCLASSDESC {
		return nil, fmt.Errorf("JAVA_TC_NULL and JAVA_TC_PROXYCLASSDESC is not allowed on index %v", stream.CurrentIndex())
	} else if array.ClassPointer.Flag == JAVA_TC_CLASSDESC {
		className = array.ClassPointer.NormalClassDesc.ClassName.Data
	} else {
		if array.ClassPointer.Reference.Flag == JAVA_TC_CLASSDESC {
			className = array.ClassPointer.Reference.NormalClassDesc.ClassName.Data
		} else {
			return nil, fmt.Errorf("JAVA_TC_PROXYCLASSDESC is not allowed on index %v", stream.CurrentIndex())
		}
	}

	if !strings.HasPrefix(className, "[") || len(className) < 2 {
		return nil, fmt.Errorf("JAVA_TC_ARRAY ClassName %v is error in %v", className, stream.CurrentIndex())
	}

	size := binary.BigEndian.Uint32(bs)
	for i := uint32(0); i < size; i++ {
		value, err := readTCValue(stream, className[1:2])
		if err != nil {
			return nil, err
		}

		array.ArrayData = append(array.ArrayData, value)
	}

	return array, nil
}
