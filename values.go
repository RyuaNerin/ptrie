package ptrie

import (
	"encoding/binary"
	"io"
	"sync"
)

type KeyValueType interface {
	comparable
}

type values[T comparable] struct {
	data []T
	*sync.RWMutex
	registry map[T]uint32
}

func (v *values[T]) put(value T) (uint32, error) {
	key := value
	v.RLock()
	result, ok := v.registry[key]
	v.RUnlock()
	if ok {
		return result, nil
	}
	v.Lock()
	defer v.Unlock()

	result = uint32(len(v.registry))
	v.registry[key] = result
	v.data = append(v.data, value)
	return result, nil
}

func (v *values[T]) Decode(reader io.Reader) error {
	var err error
	control := uint8(0)
	if err = binary.Read(reader, binary.LittleEndian, &control); err == nil {
		length := uint32(0)
		if err = binary.Read(reader, binary.LittleEndian, &length); err == nil {
			if length == 0 {
				return nil
			}

			v.data = make([]T, length)

			switch arrTyped := (any(v.data)).(type) {
			case []string:
				for i := range arrTyped {
					var length uint32
					err = binary.Read(reader, binary.LittleEndian, &length)
					if err == nil {
						var item = make([]byte, length)
						if err = binary.Read(reader, binary.LittleEndian, item); err == nil {
							arrTyped[i] = string(item)
						}
					}
					if err != nil {
						return err
					}
				}

			case []int:
				for i := range arrTyped {
					var item int64
					err = binary.Read(reader, binary.LittleEndian, &item)
					if err != nil {
						return err
					}
					arrTyped[i] = int(item)
				}

			case []uint:
				for i := range arrTyped {
					var item uint64
					err = binary.Read(reader, binary.LittleEndian, &item)
					if err != nil {
						return err
					}
					arrTyped[i] = uint(item)
				}

			default:
				for i := range v.data {
					err = binary.Read(reader, binary.LittleEndian, &v.data[i])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (v *values[T]) Encode(writer io.Writer) error {
	var err error
	if err = binary.Write(writer, binary.LittleEndian, controlByte); err == nil {
		if err = binary.Write(writer, binary.LittleEndian, uint32(len(v.data))); err == nil {
			if len(v.data) == 0 {
				return nil
			}

			switch arrTyped := (any(v.data)).(type) {
			case []string:
				for i := range arrTyped {
					length := uint32(len(arrTyped[i]))
					if err = binary.Write(writer, binary.LittleEndian, length); err == nil {
						err = binary.Write(writer, binary.LittleEndian, []byte(arrTyped[i]))
					}
					if err != nil {
						return err
					}
				}

			case []int:
				for i := range arrTyped {
					item := int64(arrTyped[i])
					err = binary.Write(writer, binary.LittleEndian, item)
					if err != nil {
						return err
					}
				}

			case []uint:
				for i := range arrTyped {
					item := uint64(arrTyped[i])
					err = binary.Write(writer, binary.LittleEndian, item)
					if err != nil {
						return err
					}
				}

			default:
				for i := range v.data {
					err = binary.Write(writer, binary.LittleEndian, v.data[i])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (v *values[T]) value(index uint32) T {
	v.RLock()
	defer v.RUnlock()
	return v.data[index]
}

func newValues[T comparable]() *values[T] {
	return &values[T]{
		data:     make([]T, 0),
		registry: make(map[T]uint32),
		RWMutex:  &sync.RWMutex{},
	}
}
