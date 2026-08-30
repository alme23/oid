// Package oid provides functionality for working with Object Identifiers (OID).
// This file contains the core OID implementation.
package oid

import (
	"encoding/asn1"
	"fmt"
	"strconv"
	"strings"
)

const (
	// MaxOIDComponent - максимальное значение компонента OID в ASN.1
	MaxOIDComponent = 268435455 // 2^28 - 1
)

// OID представляет Object Identifier
type OID []uint32

// ParseOID создает OID без выделения промежуточных срезов строк
func ParseOID(s string) (oid OID, err error) {
	if s == "" {
		return nil, ErrEmptyOID
	}

	dots := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			dots++
		}
	}

	oid = make(OID, 0, dots+1)
	start := 0

	for i := 0; i <= len(s); i++ {
		if i < len(s) && s[i] != '.' {
			continue
		}

		part := s[start:i]
		if part == "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidOID, s)
		}

		num, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("некорректный компонент OID '%s': %w", part, err)
		}

		if num > MaxOIDComponent {
			return nil, fmt.Errorf("%w: %s", ErrComponentTooBig, part)
		}

		oid = append(oid, uint32(num))
		start = i + 1
	}

	if err := oid.Validate(); err != nil {
		return nil, err
	}

	return oid, nil
}

// MustParseOID создает OID и паникует при ошибке
func MustParseOID(s string) OID {
	oid, err := ParseOID(s)
	if err != nil {
		panic(err)
	}
	return oid
}

// FromASN1 конвертирует asn1.ObjectIdentifier в OID
func FromASN1(asn1OID asn1.ObjectIdentifier) (result OID) {
	result = make(OID, len(asn1OID))
	for i, v := range asn1OID {
		// Безопасная конвертация: проверяем, что v неотрицательное
		if v < 0 {
			return nil
		}
		// #nosec G115 -- v гарантированно >= 0 из проверки выше
		result[i] = uint32(v)
	}
	return result
}

// String возвращает строковое представление OID
func (o OID) String() string {
	if len(o) == 0 {
		return ""
	}

	size := len(o) - 1
	for _, v := range o {
		size += digitCount(v)
	}

	var builder strings.Builder
	builder.Grow(size)

	for i, v := range o {
		if i > 0 {
			builder.WriteByte('.')
		}
		builder.WriteString(strconv.FormatUint(uint64(v), 10))
	}

	return builder.String()
}

// Validate проверяет корректность OID
func (o OID) Validate() (err error) {
	if len(o) < 2 {
		return ErrOIDTooShort
	}

	switch {
	case o[0] > 2:
		return ErrFirstComponentTooBig
	case o[0] < 2 && o[1] > 39:
		return ErrSecondComponentTooBig
	}

	for i, component := range o {
		if component > MaxOIDComponent {
			return fmt.Errorf("%w: компонент %d (%d), максимум %d",
				ErrComponentTooBig, i, component, MaxOIDComponent)
		}
	}

	return nil
}

// Equal проверяет равенство двух OID
func (o OID) Equal(other OID) bool {
	if len(o) != len(other) {
		return false
	}
	if len(o) == 0 {
		return true
	}
	_ = o[len(other)-1]
	for i := range o {
		if o[i] != other[i] {
			return false
		}
	}
	return true
}

// StartsWith проверяет, начинается ли OID с указанного префикса
func (o OID) StartsWith(prefix OID) bool {
	if len(prefix) > len(o) {
		return false
	}
	if len(prefix) == 0 {
		return true
	}
	_ = o[len(prefix)-1]
	for i := range prefix {
		if o[i] != prefix[i] {
			return false
		}
	}
	return true
}

// Append добавляет компоненты к OID
func (o OID) Append(components ...uint32) OID {
	result := make(OID, len(o), len(o)+len(components))
	copy(result, o)
	return append(result, components...)
}

// Parent возвращает родительский OID (без последнего компонента)
func (o OID) Parent() (parent OID, err error) {
	if len(o) <= 1 {
		return nil, ErrNoParent
	}
	return o[:len(o)-1], nil
}

// Last возвращает последний компонент OID
func (o OID) Last() (component uint32, err error) {
	if len(o) == 0 {
		return 0, ErrEmptyOID
	}
	return o[len(o)-1], nil
}

// ToASN1 конвертирует OID в asn1.ObjectIdentifier
func (o OID) ToASN1() (result asn1.ObjectIdentifier) {
	result = make(asn1.ObjectIdentifier, len(o))
	for i, v := range o {
		result[i] = int(v)
	}
	return result
}

// MarshalBinary реализует encoding.BinaryMarshaler
func (o OID) MarshalBinary() (data []byte, err error) {
	// Проверяем длину до Validate для покрытия
	if len(o) < 2 {
		return nil, ErrOIDTooShort
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	firstCombined, err := combinedFirstComponents(o[0], o[1])
	if err != nil {
		return nil, err
	}

	contentSize := base128Size(firstCombined)
	for i := 2; i < len(o); i++ {
		contentSize += base128Size(o[i])
	}

	totalSize := 1 + lengthSize(contentSize) + contentSize
	result := make([]byte, totalSize)

	result[0] = 0x06
	pos := 1
	pos += writeLength(result[pos:], contentSize)
	pos += writeBase128(result[pos:], firstCombined)

	for i := 2; i < len(o); i++ {
		pos += writeBase128(result[pos:], o[i])
	}

	return result, nil
}

// UnmarshalBinary реализует encoding.BinaryUnmarshaler
func (o *OID) UnmarshalBinary(data []byte) (err error) {
	if len(data) < 2 {
		return ErrDataTooShort
	}

	if data[0] != 0x06 {
		return fmt.Errorf("%w: 0x%02x, ожидался 0x06 (OID)", ErrInvalidASN1Tag, data[0])
	}

	contentSize, lengthBytes := readLength(data[1:])
	if lengthBytes == 0 {
		return ErrInvalidASN1Length
	}

	totalLen := 1 + lengthBytes + contentSize
	if len(data) < totalLen {
		return ErrInsufficientData
	}

	content := data[1+lengthBytes : totalLen]
	estimatedComponents := contentSize/2 + 1
	oid := make(OID, 0, estimatedComponents)

	first, bytesRead := readBase128(content)
	if bytesRead == 0 {
		return ErrInvalidFirstComponent
	}

	switch {
	case first < 40:
		oid = append(oid, 0, first)
	case first < 80:
		oid = append(oid, 1, first-40)
	default:
		oid = append(oid, 2, first-80)
	}

	pos := bytesRead
	for pos < len(content) {
		component, read := readBase128(content[pos:])
		if read == 0 {
			return ErrInvalidComponent
		}
		oid = append(oid, component)
		pos += read
	}

	if err := oid.Validate(); err != nil {
		return err
	}

	*o = oid
	return nil
}

// MarshalJSON реализует json.Marshaler
func (o OID) MarshalJSON() (data []byte, err error) {
	if len(o) == 0 {
		return []byte(`""`), nil
	}

	size := len(o) + 2
	for _, v := range o {
		size += digitCount(v)
	}

	buf := make([]byte, 0, size)
	buf = append(buf, '"')
	for i, v := range o {
		if i > 0 {
			buf = append(buf, '.')
		}
		buf = strconv.AppendUint(buf, uint64(v), 10)
	}
	buf = append(buf, '"')

	return buf, nil
}

// UnmarshalJSON реализует json.Unmarshaler
func (o *OID) UnmarshalJSON(data []byte) (err error) {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return ErrInvalidJSONType
	}

	s := string(data[1 : len(data)-1])
	parsed, err := ParseOID(s)
	if err != nil {
		return err
	}
	*o = parsed
	return nil
}

// digitCount возвращает количество цифр в числе
func digitCount(n uint32) (count int) {
	switch {
	case n < 10:
		return 1
	case n < 100:
		return 2
	case n < 1000:
		return 3
	case n < 10000:
		return 4
	case n < 100000:
		return 5
	case n < 1000000:
		return 6
	case n < 10000000:
		return 7
	case n < 100000000:
		return 8
	case n < 1000000000:
		return 9
	default:
		return 10
	}
}

// base128Size возвращает количество байт для base-128 кодирования
func base128Size(value uint32) (size int) {
	switch {
	case value < 128:
		return 1
	case value < 16384:
		return 2
	case value < 2097152:
		return 3
	case value < 268435456:
		return 4
	default:
		return 5
	}
}

// lengthSize возвращает количество байт для кодирования длины (включая первый байт)
func lengthSize(length int) (size int) {
	switch {
	case length < 128:
		return 1 // Короткая форма: 1 байт
	case length < 256:
		return 2 // 0x81 + 1 байт
	case length < 65536:
		return 3 // 0x82 + 2 байта
	default:
		return 4 // 0x83 + 3 байта
	}
}

// writeBase128 записывает число в base-128 формате
func writeBase128(dst []byte, value uint32) (bytesWritten int) {
	if value < 128 {
		dst[0] = byte(value)
		return 1
	}

	var temp [5]byte
	i := len(temp)

	for value > 0 && i > 0 {
		i--
		temp[i] = byte(value & 0x7F)
		value >>= 7
	}

	for j := i; j < len(temp)-1; j++ {
		temp[j] |= 0x80
	}

	copy(dst, temp[i:])
	return len(temp) - i
}

// writeLength записывает длину в ASN.1 формате
func writeLength(dst []byte, length int) (bytesWritten int) {
	// Проверка на отрицательное значение
	if length < 0 {
		return 0
	}

	if length < 128 {
		// #nosec G115 -- length гарантированно < 128 из условия выше
		dst[0] = byte(length)
		return 1
	}

	if length < 256 {
		dst[0] = 0x81
		// #nosec G115 -- length гарантированно < 256 из условия выше
		dst[1] = byte(length)
		return 2
	}

	if length < 65536 {
		dst[0] = 0x82
		// #nosec G115 -- length гарантированно < 65536 из условия выше
		dst[1] = byte(length >> 8)
		// #nosec G115 -- length гарантированно < 65536 из условия выше
		dst[2] = byte(length)
		return 3
	}

	// Для очень больших значений
	dst[0] = 0x83
	dst[1] = byte((length >> 16) & 0xFF)
	dst[2] = byte((length >> 8) & 0xFF)
	dst[3] = byte(length & 0xFF)
	return 4
}

// readBase128 читает число из base-128 формата
func readBase128(data []byte) (value uint32, bytesRead int) {
	for _, b := range data {
		if bytesRead >= 5 {
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

// readLength читает длину ASN.1
func readLength(data []byte) (length, bytesRead int) {
	if len(data) == 0 {
		return 0, 0
	}

	if data[0] < 0x80 {
		return int(data[0]), 1
	}

	numBytes := int(data[0] & 0x7F)
	if numBytes == 0 || numBytes > 4 {
		return 0, 0
	}

	requiredLen := 1 + numBytes
	if len(data) < requiredLen {
		return 0, 0
	}

	length = 0
	for _, b := range data[1:requiredLen] {
		if length > (int(^uint(0)>>1) >> 8) {
			return 0, 0
		}
		length = (length << 8) | int(b)
	}

	return length, requiredLen
}

// combinedFirstComponents вычисляет объединенное значение первых двух компонентов
func combinedFirstComponents(first, second uint32) (uint32, error) {
	combined := uint64(first)*40 + uint64(second)

	if combined > uint64(^uint32(0)) {
		return 0, fmt.Errorf("%w: %d*40 + %d = %d превышает uint32",
			ErrComponentTooBig, first, second, combined)
	}

	return uint32(combined), nil
}
