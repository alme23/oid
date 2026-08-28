// oid.go - с полностью ручной ASN.1 кодировкой
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
func ParseOID(s string) (OID, error) {
	if s == "" {
		return nil, fmt.Errorf("пустой OID")
	}

	// Считаем точки
	dots := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			dots++
		}
	}

	oid := make(OID, 0, dots+1)
	start := 0

	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '.' {
			part := s[start:i]
			if part == "" {
				return nil, fmt.Errorf("некорректный OID: %s", s)
			}

			num, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("некорректный компонент OID '%s': %v", part, err)
			}

			oid = append(oid, uint32(num))
			start = i + 1
		}
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

// String возвращает строковое представление OID
func (o OID) String() string {
	if len(o) == 0 {
		return ""
	}

	size := len(o) - 1 // точки
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

// digitCount возвращает количество цифр в числе
func digitCount(n uint32) int {
	if n < 10 {
		return 1
	}
	if n < 100 {
		return 2
	}
	if n < 1000 {
		return 3
	}
	if n < 10000 {
		return 4
	}
	if n < 100000 {
		return 5
	}
	if n < 1000000 {
		return 6
	}
	if n < 10000000 {
		return 7
	}
	if n < 100000000 {
		return 8
	}
	if n < 1000000000 {
		return 9
	}
	return 10
}

// Validate проверяет корректность OID
func (o OID) Validate() error {
	if len(o) < 2 {
		return fmt.Errorf("OID должен содержать минимум 2 компонента")
	}

	if o[0] > 2 {
		return fmt.Errorf("первый компонент OID должен быть 0, 1 или 2")
	}

	if o[0] < 2 && o[1] > 39 {
		return fmt.Errorf("второй компонент должен быть <= 39 при первом компоненте 0 или 1")
	}

	for i, component := range o {
		if component > MaxOIDComponent {
			return fmt.Errorf("компонент %d (%d) превышает максимальное значение ASN.1 (%d)",
				i, component, MaxOIDComponent)
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
	_ = o[len(other)-1] // BCE оптимизация
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
	_ = o[len(prefix)-1] // BCE оптимизация
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
func (o OID) Parent() (OID, error) {
	if len(o) <= 1 {
		return nil, fmt.Errorf("у OID нет родителя")
	}
	return o[:len(o)-1], nil
}

// Last возвращает последний компонент OID
func (o OID) Last() (uint32, error) {
	if len(o) == 0 {
		return 0, fmt.Errorf("пустой OID")
	}
	return o[len(o)-1], nil
}

// ToASN1 конвертирует OID в asn1.ObjectIdentifier
func (o OID) ToASN1() asn1.ObjectIdentifier {
	result := make(asn1.ObjectIdentifier, len(o))
	for i, v := range o {
		result[i] = int(v)
	}
	return result
}

// FromASN1 конвертирует asn1.ObjectIdentifier в OID
func FromASN1(asn1OID asn1.ObjectIdentifier) OID {
	result := make(OID, len(asn1OID))
	for i, v := range asn1OID {
		result[i] = uint32(v)
	}
	return result
}

// MarshalBinary реализует encoding.BinaryMarshaler
// Полностью ручная ASN.1 кодировка для максимальной производительности
func (o OID) MarshalBinary() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	// Вычисляем размер содержимого
	contentSize := 0

	// Первые два компонента
	first := 40*int(o[0]) + int(o[1])
	contentSize += base128Size(uint32(first))

	// Остальные компоненты
	for i := 2; i < len(o); i++ {
		contentSize += base128Size(o[i])
	}

	// Вычисляем общий размер: tag(1) + length(1-2) + content
	totalSize := 1 + lengthSize(contentSize) + contentSize

	// Выделяем память один раз
	result := make([]byte, totalSize)

	// Записываем тег
	result[0] = 0x06 // OID tag

	// Записываем длину
	pos := 1
	pos += writeLength(result[pos:], contentSize)

	// Записываем первые два компонента
	pos += writeBase128(result[pos:], uint32(first))

	// Записываем остальные компоненты
	for i := 2; i < len(o); i++ {
		pos += writeBase128(result[pos:], o[i])
	}

	return result, nil
}

// base128Size возвращает количество байт для base-128 кодирования
func base128Size(value uint32) int {
	if value < 128 {
		return 1
	}
	if value < 16384 {
		return 2
	}
	if value < 2097152 {
		return 3
	}
	if value < 268435456 {
		return 4
	}
	return 5
}

// lengthSize возвращает количество байт для кодирования длины
func lengthSize(length int) int {
	if length < 128 {
		return 1
	}
	if length < 256 {
		return 2
	}
	return 3
}

// writeBase128 записывает число в base-128 формате
func writeBase128(dst []byte, value uint32) int {
	if value < 128 {
		dst[0] = byte(value)
		return 1
	}

	// Временный буфер
	var temp [5]byte
	i := len(temp)

	for value > 0 {
		i--
		temp[i] = byte(value & 0x7F)
		value >>= 7
	}

	// Устанавливаем continuation bit для всех кроме последнего
	for j := i; j < len(temp)-1; j++ {
		temp[j] |= 0x80
	}

	// Копируем в destination
	copy(dst, temp[i:])
	return len(temp) - i
}

// writeLength записывает длину в ASN.1 формате
func writeLength(dst []byte, length int) int {
	if length < 128 {
		dst[0] = byte(length)
		return 1
	}

	if length < 256 {
		dst[0] = 0x81
		dst[1] = byte(length)
		return 2
	}

	dst[0] = 0x82
	dst[1] = byte(length >> 8)
	dst[2] = byte(length)
	return 3
}

// UnmarshalBinary реализует encoding.BinaryUnmarshaler
// Полностью ручная ASN.1 декодировка для максимальной производительности
func (o *OID) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("данные слишком короткие для OID")
	}

	if data[0] != 0x06 {
		return fmt.Errorf("некорректный тег ASN.1: 0x%02x, ожидался 0x06 (OID)", data[0])
	}

	// Декодируем длину
	contentSize, lengthBytes := readLength(data[1:])
	if lengthBytes == 0 {
		return fmt.Errorf("некорректная длина ASN.1")
	}

	if len(data) < 1+lengthBytes+contentSize {
		return fmt.Errorf("недостаточно данных для OID")
	}

	content := data[1+lengthBytes : 1+lengthBytes+contentSize]

	// Оцениваем количество компонентов
	estimatedComponents := contentSize/2 + 1
	oid := make(OID, 0, estimatedComponents)

	// Декодируем первый компонент (объединенные первые два)
	first, bytesRead := readBase128(content)
	if bytesRead == 0 {
		return fmt.Errorf("некорректный первый компонент OID")
	}

	// Разделяем первые два компонента
	if first < 40 {
		oid = append(oid, 0, first)
	} else if first < 80 {
		oid = append(oid, 1, first-40)
	} else {
		oid = append(oid, 2, first-80)
	}

	// Декодируем остальные компоненты
	pos := bytesRead
	for pos < len(content) {
		component, read := readBase128(content[pos:])
		if read == 0 {
			return fmt.Errorf("некорректный компонент OID")
		}
		oid = append(oid, component)
		pos += read
	}

	// Проверяем валидность
	if err := oid.Validate(); err != nil {
		return err
	}

	*o = oid
	return nil
}

// readBase128 читает число из base-128 формата
func readBase128(data []byte) (uint32, int) {
	var result uint32
	bytesRead := 0

	for _, b := range data {
		if bytesRead >= 5 {
			return 0, 0 // Слишком длинное число
		}

		result = (result << 7) | uint32(b&0x7F)
		bytesRead++

		if b&0x80 == 0 {
			return result, bytesRead
		}
	}

	return 0, 0 // Незавершенная последовательность
}

// readLength читает длину ASN.1
func readLength(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}

	if data[0] < 0x80 {
		return int(data[0]), 1
	}

	numBytes := int(data[0] & 0x7F)
	if numBytes == 0 || numBytes > 4 || len(data) < 1+numBytes {
		return 0, 0
	}

	length := 0
	for i := 1; i <= numBytes; i++ {
		length = (length << 8) | int(data[i])
	}

	return length, 1 + numBytes
}

// MarshalJSON реализует json.Marshaler
func (o OID) MarshalJSON() ([]byte, error) {
	if len(o) == 0 {
		return []byte(`""`), nil
	}

	size := len(o) + 2 // кавычки и точки
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
func (o *OID) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("некорректный JSON тип для OID")
	}

	s := string(data[1 : len(data)-1])
	parsed, err := ParseOID(s)
	if err != nil {
		return err
	}
	*o = parsed
	return nil
}
