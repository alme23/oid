package oid

import (
	"fmt"
)

const (
	berTagOID = 0x06 // Тег ASN.1 для Object Identifier
)

// AppendBER кодирует OID в формате BER/DER без тега и длины
func (o OID) AppendBER(dst []byte) (result []byte, err error) {
	if err := o.Validate(); err != nil {
		return dst, err
	}

	// Проверка границ
	if len(o) < 2 {
		return dst, ErrOIDTooShort
	}

	firstCombined := uint32(o[0]*40 + o[1])
	dst = appendBase128Value(dst, firstCombined)

	for _, v := range o[2:] {
		dst = appendBase128Value(dst, v)
	}

	return dst, nil
}

// MarshalBER реализует BER-кодирование полного ASN.1 TLV-пакета
func (o OID) MarshalBER() (data []byte, err error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	if len(o) < 2 {
		return nil, ErrOIDTooShort
	}

	firstCombined, err := combinedFirstComponents(o[0], o[1])
	if err != nil {
		return nil, err
	}

	contentSize := base128Size(firstCombined)
	for _, v := range o[2:] {
		contentSize += base128Size(v)
	}

	// Создаем результат с правильным размером
	var result []byte

	if contentSize < 128 {
		// Короткая форма: тег + 1 байт длины + контент
		result = make([]byte, 0, 1+1+contentSize)
		result = append(result, berTagOID)
		result = append(result, byte(contentSize))
	} else {
		// Длинная форма
		var lenBuf [4]byte
		lIdx := 4
		tempLen := contentSize
		for tempLen > 0 && lIdx > 0 {
			lIdx--
			lenBuf[lIdx] = byte(tempLen & 0xFF)
			tempLen >>= 8
		}
		numLenBytes := 4 - lIdx

		// Размер: тег + 1 байт длины + numLenBytes байт длины + контент
		result = make([]byte, 0, 1+1+numLenBytes+contentSize)
		result = append(result, berTagOID)
		result = append(result, byte(0x80|numLenBytes))
		result = append(result, lenBuf[lIdx:]...)
	}

	return o.AppendBER(result)
}

// UnmarshalBER декодирует OID из полного BER/DER TLV-пакета
func (o *OID) UnmarshalBER(data []byte) (err error) {
	if len(data) < 2 {
		return ErrInsufficientData
	}

	if data[0] != berTagOID {
		return fmt.Errorf("%w: ожидался 0x06, получен 0x%02x", ErrInvalidASN1Tag, data[0])
	}

	var length int
	var headerLen int

	if data[1] < 128 {
		length = int(data[1])
		headerLen = 2
	} else {
		numLenBytes := int(data[1] & 0x7F)
		if numLenBytes == 0 || numLenBytes > 4 {
			return ErrInvalidLength
		}

		if len(data) < 2+numLenBytes {
			return ErrInsufficientData
		}

		for _, b := range data[2 : 2+numLenBytes] {
			length = (length << 8) | int(b)
		}
		headerLen = 2 + numLenBytes
	}

	if length == 0 {
		return ErrEmptyContent
	}

	if len(data) < headerLen+length {
		return ErrInsufficientData
	}

	if len(data) != headerLen+length {
		return fmt.Errorf("%w: длина данных (%d) не совпадает с ASN.1 длиной (%d)",
			ErrInvalidLength, len(data), headerLen+length)
	}

	content := data[headerLen : headerLen+length]
	return o.UnmarshalBERContent(content)
}

// UnmarshalBERContent декодирует OID из BER/DER Content Bytes
func (o *OID) UnmarshalBERContent(content []byte) (err error) {
	if len(content) == 0 {
		return ErrEmptyContent
	}

	firstCombined, bytesRead := readBase128FromBytes(content)
	if bytesRead == 0 {
		return ErrFirstComponentFailed
	}

	var first, second uint32
	switch {
	case firstCombined < 40:
		first = 0
		second = firstCombined
	case firstCombined < 80:
		first = 1
		second = firstCombined - 40
	default:
		first = 2
		second = firstCombined - 80
	}

	res := make(OID, 0, bytesRead+2)
	res = append(res, first, second)

	pos := bytesRead
	for pos < len(content) {
		component, read := readBase128FromBytes(content[pos:])
		if read == 0 {
			return ErrComponentFailed
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

// MarshalDER идентичен MarshalBER для OID
func (o OID) MarshalDER() (data []byte, err error) {
	return o.MarshalBER()
}

// UnmarshalDER идентичен UnmarshalBER для OID
func (o *OID) UnmarshalDER(data []byte) (err error) {
	return o.UnmarshalBER(data)
}

// SizeBER возвращает размер BER-кодированного OID в байтах
func (o OID) SizeBER() (size int, err error) {
	if err := o.Validate(); err != nil {
		return 0, err
	}

	if len(o) < 2 {
		return 0, ErrOIDTooShort
	}

	firstCombined, err := combinedFirstComponents(o[0], o[1])
	if err != nil {
		return 0, err
	}

	contentSize := base128Size(firstCombined)
	for _, v := range o[2:] {
		contentSize += base128Size(v)
	}

	// Заголовок: тег (1 байт) + длина
	if contentSize < 128 {
		// Короткая форма: 1 байт тега + 1 байт длины
		return 1 + 1 + contentSize, nil
	}

	// Длинная форма: 1 байт тега + 1 байт (0x80|numBytes) + numBytes байт длины
	lengthBytes := lengthSize(contentSize)
	return 1 + lengthBytes + contentSize, nil
}

// safeByte конвертирует int в byte с проверкой диапазона
func safeByte(value int) byte {
	// Маскируем до 8 бит для гарантии безопасности
	return byte(value & 0xFF)
}

// readBase128FromBytes читает base-128 значение из байтового среза
func readBase128FromBytes(data []byte) (value uint32, bytesRead int) {
	for _, b := range data {
		if bytesRead >= 5 {
			return 0, 0
		}

		if value > (0xFFFFFFFF >> 7) {
			return 0, 0
		}

		value = (value << 7) | uint32(b&0x7F)
		bytesRead++

		if b&0x80 == 0 {
			return value, bytesRead
		}
	}

	return 0, 0
}

// appendBase128Value кодирует uint32 в base-128 формат
func appendBase128Value(dst []byte, value uint32) []byte {
	if value < 128 {
		return append(dst, byte(value))
	}

	var buf [5]byte
	idx := 4

	buf[idx] = byte(value & 0x7F)
	value >>= 7

	for value > 0 && idx > 0 {
		idx--
		low7bits := byte(value & 0x7F)
		buf[idx] = low7bits | 0x80
		value >>= 7
	}

	return append(dst, buf[idx:]...)
}
