package oid

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Статические ошибки для database
var (
	ErrUnsupportedScanType = errors.New("неподдерживаемый тип для конвертации в OID")
	ErrInvalidArrayFormat  = errors.New("некорректный формат массива OID")
	ErrSaveValidation      = errors.New("невалидный OID перед сохранением в БД")
	ErrDatabaseParse       = errors.New("ошибка парсинга OID из БД")
	ErrInvalidArrayOID     = errors.New("невалидный OID в массиве")
	ErrJSONDecodeArray     = errors.New("ошибка декодирования JSON массива")
)

// Scan реализует интерфейс sql.Scanner для базового OID
func (o *OID) Scan(value any) error {
	if value == nil {
		*o = nil
		return nil
	}

	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedScanType, value)
	}

	parsed, err := ParseOID(s)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDatabaseParse, err)
	}

	*o = parsed
	return nil
}

// Value реализует интерфейс driver.Valuer для базового OID
func (o OID) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	if err := o.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSaveValidation, err)
	}
	return o.String(), nil
}

// NullOID представляет OID, который может быть NULL в базе данных.
type NullOID struct {
	OID   OID
	Valid bool
}

// FromOID создает NullOID из обычного OID
func FromOID(oid OID) NullOID {
	if len(oid) == 0 {
		return NullOID{Valid: false}
	}
	return NullOID{OID: oid, Valid: true}
}

// FromString создает NullOID из строки
func FromString(s string) (NullOID, error) {
	if s == "" {
		return NullOID{Valid: false}, nil
	}

	oid, err := ParseOID(s)
	if err != nil {
		return NullOID{}, err
	}

	return NullOID{OID: oid, Valid: true}, nil
}

// MustFromString создает NullOID из строки и паникует при ошибке
func MustFromString(s string) NullOID {
	n, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return n
}

// Value реализует интерфейс driver.Valuer для NullOID
func (n NullOID) Value() (driver.Value, error) {
	if !n.Valid || len(n.OID) == 0 {
		return nil, nil
	}
	return n.OID.Value()
}

// Scan реализует интерфейс sql.Scanner для NullOID
func (n *NullOID) Scan(value any) error {
	if value == nil {
		n.OID = nil
		n.Valid = false
		return nil
	}

	n.Valid = true
	return n.OID.Scan(value)
}

// String возвращает строковое представление NullOID
func (n NullOID) String() string {
	if !n.Valid {
		return ""
	}
	return n.OID.String()
}

// Equal проверяет равенство двух NullOID
func (n NullOID) Equal(other NullOID) bool {
	if !n.Valid && !other.Valid {
		return true
	}
	if n.Valid != other.Valid {
		return false
	}
	return n.OID.Equal(other.OID)
}

// MarshalJSON реализует json.Marshaler для NullOID
func (n NullOID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.OID.MarshalJSON()
}

// UnmarshalJSON реализует json.Unmarshaler для NullOID
func (n *NullOID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		n.OID = nil
		n.Valid = false
		return nil
	}

	if err := n.OID.UnmarshalJSON(data); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// Array представляет массив OID для работы с PostgreSQL массивами
type Array []OID

// Value реализует driver.Valuer для OIDArray
func (oa Array) Value() (driver.Value, error) {
	if len(oa) == 0 {
		return "{}", nil
	}

	size := 2 // фигурные скобки
	for i, oid := range oa {
		if i > 0 {
			size++ // запятая
		}
		if err := oid.Validate(); err != nil {
			return nil, fmt.Errorf("%w: позиция %d: %w", ErrInvalidArrayOID, i, err)
		}
		size += len(oid.String())
	}

	buf := make([]byte, 0, size)
	buf = append(buf, '{')
	for i, oid := range oa {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, oid.String()...)
	}
	buf = append(buf, '}')

	return string(buf), nil
}

// Scan реализует интерфейс sql.Scanner для OIDArray
func (oa *Array) Scan(value any) error {
	if value == nil {
		*oa = nil
		return nil
	}

	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("%w (OIDArray): %T", ErrUnsupportedScanType, value)
	}

	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("%w: %s", ErrInvalidArrayFormat, s)
	}

	if s == "{}" {
		*oa = Array{}
		return nil
	}

	content := s[1 : len(s)-1]
	parts := splitPostgresArray(content)

	result := make(Array, 0, len(parts))
	for _, part := range parts {
		oid, err := ParseOID(part)
		if err != nil {
			return fmt.Errorf("%w: '%s': %w", ErrDatabaseParse, part, err)
		}
		result = append(result, oid)
	}

	*oa = result
	return nil
}

// MarshalJSON реализует json.Marshaler для OIDArray
func (oa Array) MarshalJSON() ([]byte, error) {
	if len(oa) == 0 {
		return []byte("[]"), nil
	}

	buf := make([]byte, 0, len(oa)*10)
	buf = append(buf, '[')
	for i, oid := range oa {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = append(buf, oid.String()...)
		buf = append(buf, '"')
	}
	buf = append(buf, ']')

	return buf, nil
}

// UnmarshalJSON реализует json.Unmarshaler для OIDArray
func (oa *Array) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" || string(data) == "null" {
		*oa = Array{}
		return nil
	}

	var strArray []string
	if err := json.Unmarshal(data, &strArray); err != nil {
		return fmt.Errorf("%w: %w", ErrJSONDecodeArray, err)
	}

	result := make(Array, 0, len(strArray))
	for _, s := range strArray {
		oid, err := ParseOID(s)
		if err != nil {
			return fmt.Errorf("%w: '%s': %w", ErrDatabaseParse, s, err)
		}
		result = append(result, oid)
	}

	*oa = result
	return nil
}

// String возвращает строковое представление OIDArray
func (oa Array) String() string {
	if len(oa) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteByte('[')
	for i, oid := range oa {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(oid.String())
	}
	builder.WriteByte(']')

	return builder.String()
}

// Equal проверяет равенство двух OIDArray
func (oa Array) Equal(other Array) bool {
	if len(oa) != len(other) {
		return false
	}
	for i := range oa {
		if !oa[i].Equal(other[i]) {
			return false
		}
	}
	return true
}

// Contains проверяет наличие OID в массиве
func (oa Array) Contains(target OID) bool {
	for _, oid := range oa {
		if oid.Equal(target) {
			return true
		}
	}
	return false
}

// Append добавляет OID в массив
func (oa Array) Append(oids ...OID) Array {
	result := make(Array, len(oa), len(oa)+len(oids))
	copy(result, oa)
	return append(result, oids...)
}

// splitPostgresArray разделяет строку PostgreSQL массива на элементы
func splitPostgresArray(s string) []string {
	var parts []string
	var current []byte
	inQuotes := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			current = append(current, ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			current = append(current, ch)
			continue
		}

		switch ch {
		case '"':
			inQuotes = !inQuotes
			current = append(current, ch)
		case ',':
			if !inQuotes {
				part := string(current)
				if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
					part = part[1 : len(part)-1]
				}
				parts = append(parts, part)
				current = current[:0]
			} else {
				current = append(current, ch)
			}
		default:
			current = append(current, ch)
		}
	}

	if len(current) > 0 {
		part := string(current)
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}
		parts = append(parts, part)
	}

	return parts
}
