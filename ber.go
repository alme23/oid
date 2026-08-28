// oid/ber.go - исправленная версия с правильным BER кодированием
package oid

import (
	"errors"
	"fmt"
	"sync"
)

const (
	berTagOID = 0x06 // Тег ASN.1 для Object Identifier
)

// bytePool для переиспользования буферов при BER кодировании
var bytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 64)
		return &b
	},
}

// AppendBER кодирует OID в формате BER/DER без тега и длины
// Следует стандарту RFC 2578 для кодирования OID
func (o OID) AppendBER(dst []byte) ([]byte, error) {
	if err := o.Validate(); err != nil {
		return dst, err
	}

	// Кодируем первые два компонента согласно RFC 2578:
	// - Если first < 2: значение = 40 * first + second (всегда < 80)
	// - Если first == 2: значение = 80 + second (может быть >= 80, тогда base-128)
	firstCombined := uint32(o[0]*40 + o[1])

	// Кодируем первое значение в base-128
	dst = appendBase128Value(dst, firstCombined)

	// Кодируем остальные компоненты
	for _, v := range o[2:] {
		dst = appendBase128Value(dst, v)
	}

	return dst, nil
}

// appendBase128Value кодирует uint32 в base-128 формат
func appendBase128Value(dst []byte, value uint32) []byte {
	if value < 128 {
		return append(dst, byte(value))
	}

	// Буфер для разворота 7-битных групп (максимум 5 байт для uint32)
	var buf [5]byte
	idx := 4

	buf[idx] = byte(value & 0x7f) // Последний байт (старший бит 0)
	value >>= 7

	for value > 0 {
		idx--
		buf[idx] = byte((value & 0x7f) | 0x80) // Промежуточные байты (старший бит 1)
		value >>= 7
	}

	return append(dst, buf[idx:]...)
}

// MarshalBER реализует BER-кодирование полного ASN.1 TLV-пакета
func (o OID) MarshalBER() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	// Вычисляем размер контента
	contentSize := base128Size(uint32(o[0]*40 + o[1]))
	for _, v := range o[2:] {
		contentSize += base128Size(v)
	}

	// Вычисляем размер заголовка
	headerSize := 2 // тег + короткая длина
	if contentSize >= 128 {
		headerSize = 2 + lengthSize(contentSize)
	}

	// Одна аллокация под весь результат
	result := make([]byte, 0, headerSize+contentSize)

	// Добавляем тег
	result = append(result, berTagOID)

	// Добавляем длину
	if contentSize < 128 {
		result = append(result, byte(contentSize))
	} else {
		var lenBuf [4]byte
		lIdx := 4
		tempLen := contentSize
		for tempLen > 0 {
			lIdx--
			lenBuf[lIdx] = byte(tempLen & 0xff)
			tempLen >>= 8
		}
		numLenBytes := 4 - lIdx

		result = append(result, 0x80|byte(numLenBytes))
		result = append(result, lenBuf[lIdx:]...)
	}

	// Добавляем контент
	var err error
	result, err = o.AppendBER(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UnmarshalBER декодирует OID из полного BER/DER TLV-пакета
func (o *OID) UnmarshalBER(data []byte) error {
	if len(data) < 3 {
		return errors.New("недостаточно данных для BER OID")
	}

	// Проверяем ASN.1 тег
	if data[0] != berTagOID {
		return fmt.Errorf("некорректный ASN.1 тег: ожидался 0x06, получен 0x%02x", data[0])
	}

	// Декодируем длину контента
	var length int
	var headerLen int

	if data[1] < 128 {
		length = int(data[1])
		headerLen = 2
	} else {
		numLenBytes := int(data[1] & 0x7f)
		if numLenBytes == 0 || numLenBytes > 4 {
			return errors.New("некорректная кодировка длины в BER")
		}
		if len(data) < 2+numLenBytes {
			return errors.New("недостаточно данных для чтения длины BER")
		}

		for i := 0; i < numLenBytes; i++ {
			length = (length << 8) | int(data[2+i])
		}
		headerLen = 2 + numLenBytes
	}

	if len(data) != headerLen+length {
		return fmt.Errorf("длина данных (%d) не совпадает с ASN.1 длиной (%d)",
			len(data), headerLen+length)
	}

	content := data[headerLen:]
	if len(content) == 0 {
		return errors.New("пустой контент BER OID")
	}

	return o.UnmarshalBERContent(content)
}

// UnmarshalBERContent декодирует OID из BER/DER Content Bytes
func (o *OID) UnmarshalBERContent(content []byte) error {
	if len(content) == 0 {
		return errors.New("пустой контент BER OID")
	}

	// Декодируем первый компонент (base-128)
	firstCombined, bytesRead := readBase128FromBytes(content)
	if bytesRead == 0 {
		return errors.New("некорректный первый компонент BER OID")
	}

	// Разделяем первые два компонента согласно RFC 2578
	var first, second uint32
	if firstCombined < 40 {
		first = 0
		second = firstCombined
	} else if firstCombined < 80 {
		first = 1
		second = firstCombined - 40
	} else {
		first = 2
		second = firstCombined - 80
	}

	// Создаем результирующий OID
	res := make(OID, 0, bytesRead+2)
	res = append(res, first, second)

	// Декодируем остальные компоненты
	pos := bytesRead
	for pos < len(content) {
		component, read := readBase128FromBytes(content[pos:])
		if read == 0 {
			return errors.New("некорректный компонент BER OID")
		}
		res = append(res, component)
		pos += read
	}

	if err := res.Validate(); err != nil {
		return err
	}

	*o = res
	return nil
}

// readBase128FromBytes читает base-128 значение из байтового среза
func readBase128FromBytes(data []byte) (uint32, int) {
	var result uint32
	bytesRead := 0

	for _, b := range data {
		if bytesRead >= 5 {
			return 0, 0 // Слишком длинное число
		}

		result = (result << 7) | uint32(b&0x7f)
		bytesRead++

		if b&0x80 == 0 {
			return result, bytesRead
		}
	}

	return 0, 0 // Незавершенная последовательность
}

// MarshalDER идентичен MarshalBER для OID
func (o OID) MarshalDER() ([]byte, error) {
	return o.MarshalBER()
}

// UnmarshalDER идентичен UnmarshalBER для OID
func (o *OID) UnmarshalDER(data []byte) error {
	return o.UnmarshalBER(data)
}

// SizeBER возвращает размер BER-кодированного OID в байтах
func (o OID) SizeBER() (int, error) {
	if err := o.Validate(); err != nil {
		return 0, err
	}

	contentSize := base128Size(uint32(o[0]*40 + o[1]))
	for _, v := range o[2:] {
		contentSize += base128Size(v)
	}

	if contentSize < 128 {
		return 2 + contentSize, nil
	}

	return 2 + lengthSize(contentSize) + contentSize, nil
}
