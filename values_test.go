package ptrie

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testValues_Decode[T comparable](t *testing.T, useCaseDescription string, useCaseValues []T, useCaseHasError bool) {
	values := newValues[T]()
	for _, item := range useCaseValues {
		_, err := values.put(item)
		assert.Nil(t, err, useCaseDescription)
	}
	writer := new(bytes.Buffer)
	err := values.Encode(writer)
	if useCaseHasError {
		assert.NotNil(t, err, useCaseDescription)
		cloned := newValues[T]()
		err = cloned.Decode(writer)
		assert.NotNil(t, err)
		return
	}
	if !assert.Nil(t, err, useCaseDescription) {
		log.Print(err)
		return
	}
	cloned := newValues[T]()
	err = cloned.Decode(writer)
	assert.Nil(t, err, useCaseDescription)
	assert.EqualValues(t, len(values.data), len(cloned.data))
	for i := range values.data {
		assert.EqualValues(t, values.data[i], cloned.data[i], fmt.Sprintf("[%d]: %v", i, useCaseDescription))
	}
}

func TestValues_Decode(t *testing.T) {
	testValues_Decode[string](t, "string coding", []string{"abc", "xyz", "klm", "xyz", "eee"}, false)
	testValues_Decode[int](t, "int coding", []int{int(0), int(10), int(30), int(300), int(4)}, false)
	testValues_Decode[int8](t, "int8 coding", []int8{int8(3), int8(10), int8(30), int8(121), int8(4)}, false)
	testValues_Decode[bool](t, "bool coding", []bool{true, false}, false)
	/**
	testValues_Decode[[]byte](t, "[]byte coding", [][]byte{[]byte("abc"), []byte("xyz")}, false)
	testValues_Decode[*foo](t, "custom type coding", []*foo{&foo{ID: 10, Name: "abc"}, &foo{ID: 20, Name: "xyz"}}, false)
	testValues_Decode[*bar](t, "custom type error coding", []*bar{&bar{ID: 10, Name: "abc"}, &bar{ID: 20, Name: "xyz"}}, false)
	*/
}

type foo struct {
	ID   int
	Name string
}

type bar foo

func (c *bar) Key() interface{} {
	h := fnv.New32a()
	_, _ = h.Write([]byte(c.Name))
	return c.ID + 100000*int(h.Sum32())
}

func (c *foo) Key() interface{} {
	h := fnv.New32a()
	_, _ = h.Write([]byte(c.Name))
	return c.ID + 100000*int(h.Sum32())
}

func (c *foo) Decode(reader io.Reader) error {
	id := int64(0)
	if err := binary.Read(reader, binary.BigEndian, &id); err != nil {
		return err
	}
	c.ID = int(id)
	length := uint16(0)
	err := binary.Read(reader, binary.BigEndian, &length)
	if err == nil {
		name := make([]byte, length)
		if err = binary.Read(reader, binary.BigEndian, name); err == nil {
			c.Name = string(name)
		}
	}
	return err
}

func (c *foo) Encode(writer io.Writer) error {
	err := binary.Write(writer, binary.BigEndian, int64(c.ID))
	if err != nil {
		return err
	}
	length := uint16(len(c.Name))
	if err = binary.Write(writer, binary.BigEndian, length); err == nil {
		err = binary.Write(writer, binary.BigEndian, []byte(c.Name))
	}
	return err
}
