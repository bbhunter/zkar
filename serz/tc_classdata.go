package serz

import (
	"github.com/phith0n/zkar/commons"
)

type ReferenceClassInformation struct {
	ClassName  string
	Attributes []string
}

type TCClassData struct {
	HasAnnotation    bool
	ReferenceClass   *ReferenceClassInformation
	FieldDatas       []*TCValue
	ObjectAnnotation []*TCContent
}

func (cd *TCClassData) ToBytes() []byte {
	var bs []byte
	for _, data := range cd.FieldDatas {
		bs = append(bs, data.ToBytes()...)
	}

	if !cd.HasAnnotation {
		return bs
	}

	for _, content := range cd.ObjectAnnotation {
		bs = append(bs, content.ToBytes()...)
	}

	bs = append(bs, JAVA_TC_ENDBLOCKDATA)
	return bs
}

func (cd *TCClassData) ToString() string {
	var b = commons.NewPrinter()
	b.Printf("@ClassName - %s", cd.ReferenceClass.ClassName)
	b.IncreaseIndent()
	b.Print("{}Attributes")
	b.IncreaseIndent()
	for i := 0; i < len(cd.FieldDatas); i++ {
		b.Printf("%s", cd.ReferenceClass.Attributes[i])
		b.IncreaseIndent()
		b.Print(cd.FieldDatas[i].ToString())
		b.DecreaseIndent()
	}
	b.DecreaseIndent()

	if !cd.HasAnnotation {
		return b.String()
	}

	b.Print("@ObjectAnnotation")
	b.IncreaseIndent()
	for _, content := range cd.ObjectAnnotation {
		b.Print(content.ToString())
	}
	b.Printf("TC_ENDBLOCKDATA - %s", commons.Hexify(JAVA_TC_ENDBLOCKDATA))
	return b.String()
}

func (cd *TCClassData) Walk(callback WalkCallback) error {
	for _, data := range cd.FieldDatas {
		if err := callback(data); err != nil {
			return err
		}

		if err := data.Walk(callback); err != nil {
			return err
		}
	}

	if !cd.HasAnnotation {
		return nil
	}

	for _, anno := range cd.ObjectAnnotation {
		if err := callback(anno); err != nil {
			return err
		}

		if err := anno.Walk(callback); err != nil {
			return err
		}
	}

	return nil
}

func readTCClassData(stream *ObjectStream, desc *TCClassDesc) (*TCClassData, error) {
	var err error
	var classData = &TCClassData{
		ReferenceClass: &ReferenceClassInformation{
			ClassName: desc.ClassName.Data,
		},
	}

	current := stream.CurrentIndex()
	if desc.HasFlag(JAVA_SC_SERIALIZABLE) {
		for _, field := range desc.Fields {
			fieldData, err := readTCFieldData(stream, field)
			if err == NoFieldError {
				// When java.io.Serializable#defaultWriteObject is not invoke, no built-in field data is written.
				// So we should clear the classData.FieldDatas and reset the position of stream
				// Then everything will be read from objectAnnotation
				// Example: ysoserial C3O0
				stream.Seek(current)
				classData.FieldDatas = []*TCValue{}
				break
			} else if err != nil {
				return nil, err
			}

			classData.FieldDatas = append(classData.FieldDatas, fieldData)
			classData.ReferenceClass.Attributes = append(classData.ReferenceClass.Attributes, field.FieldName.Data)
		}
	}

	if (desc.HasFlag(JAVA_SC_SERIALIZABLE) && desc.HasFlag(JAVA_SC_WRITE_METHOD)) ||
		(desc.HasFlag(JAVA_SC_EXTERNALIZABLE) && desc.HasFlag(JAVA_SC_BLOCK_DATA)) {
		classData.HasAnnotation = true
		classData.ObjectAnnotation, err = readTCAnnotation(stream)
		if err != nil {
			return nil, err
		}
	}

	return classData, nil
}

func readTCFieldData(stream *ObjectStream, field *TCFieldDesc) (*TCValue, error) {
	return readTCValue(stream, field.TypeCode)
}
