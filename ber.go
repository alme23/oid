package oid

import (
	"errors"
	"fmt"
)

const (
	berTagOID = 0x06 // Тег ASN.1 для Object Identifier
)

// Статические ошибки пакета для BER/DER кодирования
var (
	ErrInvalidLength        = errors.New("некорректная кодировка длины в BER")
	ErrEmptyContent         = errors.New("пустой контент BER OID")
	ErrFirstComponentFailed = errors.New("некорректный первый компонент BER OID")
	ErrComponentFailed      = errors.New("некорректный компонент BER OID")
	ErrComponentOverflow    = errors.New("переполнение компонента OID при декодировании BER")
)

// AppendBER кодирует OID в формате BER/DER без тега и длины
func (o OID) AppendBER(dst []byte) ([]byte, error) {
	if err := o.Validate(); err != nil {
		return dst, err
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

	firstCombined, err := combinedFirstComponents(o[0], o[1])
	if err != nil {
		return nil, err
	}

	contentSize := base128Size(firstCombined)
	for _, v := range o[2:] {
		contentSize += base128Size(v)
	}

	headerSize := 2
	if contentSize >= 128 {
		headerSize = 2 + lengthSize(contentSize)
	}

	result := make([]byte, 0, headerSize+contentSize)
	result = append(result, berTagOID)

	if contentSize < 128 {
		result = append(result, byte(contentSize))
	} else {
		var lenBuf [4]byte
		lIdx := 4
		tempLen := contentSize
		for tempLen > 0 && lIdx > 0 {
			lIdx--
			lenBuf[lIdx] = byte(tempLen & 0xFF)
			tempLen >>= 8
		}
		numLenBytes := 4 - lIdx
		result = append(result, byte(0x80|numLenBytes))
		result = append(result, lenBuf[lIdx:]...)
	}

	return o.AppendBER(result)
}

// UnmarshalBER декодирует OID из полного BER/DER TLV-пакета
func (o *OID) UnmarshalBER(data []byte) (err error) {
	// Минимальная длина: тег (1) + длина (1) = 2 байта
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

	// Проверка пустого контента
	if length == 0 {
		return ErrEmptyContent
	}

	// Проверка, что данных достаточно
	if len(data) < headerLen+length {
		return ErrInsufficientData
	}

	// Проверка, что нет лишних данных
	if len(data) != headerLen+length {
		return fmt.Errorf("%w: длина данных (%d) не совпадает с ASN.1 длиной (%d)",
			ErrInvalidLength, len(data), headerLen+length)
	}

	content := data[headerLen : headerLen+length]
	return o.UnmarshalBERContent(content)
}

// UnmarshalBERContent декодирует OID из BER/DER Content Bytes
func (o *OID) UnmarshalBERContent(content []byte) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}

	// Декодируем первый компонент (base-128)
	firstCombined, bytesRead := readBase128FromBytes(content)
	if bytesRead == 0 {
		return ErrFirstComponentFailed
	}

	// Разделяем первые два компонента согласно RFC 2578
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

	// Декодируем остальные компоненты
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
func (o OID) MarshalDER() ([]byte, error) {
	return o.MarshalBER()
}

// UnmarshalDER идентичен UnmarshalBER для OID
func (o *OID) UnmarshalDER(data []byte) error {
	return o.UnmarshalBER(data)
}

// SizeBER возвращает размер BER-кодированного OID в байтах
func (o OID) SizeBER() (size int, err error) {
	if err := o.Validate(); err != nil {
		return 0, err
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
	// Для короткой формы: 1 байт длины
	// Для длинной формы: 1 байт (0x80|numBytes) + numBytes байт
	if contentSize < 128 {
		return 2 + contentSize, nil // тег + короткая длина + контент
	}

	// Длинная форма длины
	lengthBytes := lengthSize(contentSize)    // количество байт для кодирования длины
	return 1 + lengthBytes + contentSize, nil // тег + длина + контент
}

// readBase128FromBytes читает base-128 значение из байтового среза
func readBase128FromBytes(data []byte) (value uint32, bytesRead int) {
	bytesRead = 0
	value = 0

	for _, b := range data {
		if bytesRead >= 5 {
			return 0, 0
		}

		if value > (0xffffffff >> 7) {
			return 0, 0
		}

		value = (value << 7) | uint32(b&0x7f)
		bytesRead++

		if b&0x80 == 0 {
			return value, bytesRead
		}
	}

	return 0, 0
}

// appendBase128Value кодирует uint32 в base-128 формат
func appendBase128Value(dst []byte, value uint32) (result []byte) {
	if value < 128 {
		return append(dst, byte(value))
	}

	var buf [5]byte
	idx := 4

	// Последний байт
	buf[idx] = byte(value & 0x7F)
	value >>= 7

	for value > 0 && idx > 0 {
		idx--
		// Разделяем операции для избежания предупреждения
		low7bits := byte(value & 0x7F)
		buf[idx] = low7bits | 0x80
		value >>= 7
	}

	return append(dst, buf[idx:]...)
}
