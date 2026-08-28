// oid/database.go
package oid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Value реализует интерфейс driver.Valuer для базового OID
func (o OID) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	if err := o.Validate(); err != nil {
		return nil, fmt.Errorf("невалидный OID перед сохранением в БД: %w", err)
	}
	return o.String(), nil
}

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
		return fmt.Errorf("неподдерживаемый тип для конвертации в OID: %T", value)
	}

	parsed, err := ParseOID(s)
	if err != nil {
		return fmt.Errorf("ошибка парсинга OID из БД: %w", err)
	}

	*o = parsed
	return nil
}

// NullOID представляет OID, который может быть NULL в базе данных.
// Он реализует интерфейсы sql.Scanner и driver.Valuer аналогично sql.NullString.
type NullOID struct {
	OID   OID
	Valid bool // Valid равен true, если OID не NULL
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

// String возвращает строковое представление NullOID.
// Для NULL возвращает пустую строку.
func (n NullOID) String() string {
	if !n.Valid {
		return ""
	}
	return n.OID.String()
}

// Equal проверяет равенство двух NullOID.
// Два NULL значения считаются равными.
func (n NullOID) Equal(other NullOID) bool {
	if !n.Valid && !other.Valid {
		return true
	}
	if n.Valid != other.Valid {
		return false
	}
	return n.OID.Equal(other.OID)
}

// MarshalJSON реализует json.Marshaler для NullOID.
// NULL сериализуется как null, иначе как строка OID.
func (n NullOID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.OID.MarshalJSON()
}

// UnmarshalJSON реализует json.Unmarshaler для NullOID.
// Принимает null или строку с OID.
func (n *NullOID) UnmarshalJSON(data []byte) error {
	// Проверяем на null
	if string(data) == "null" || string(data) == `""` {
		n.OID = nil
		n.Valid = false
		return nil
	}

	// Декодируем как обычный OID
	if err := n.OID.UnmarshalJSON(data); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// FromOID создает NullOID из обычного OID.
func FromOID(oid OID) NullOID {
	if len(oid) == 0 {
		return NullOID{Valid: false}
	}
	return NullOID{OID: oid, Valid: true}
}

// FromString создает NullOID из строки.
// Пустая строка интерпретируется как NULL.
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

// MustFromString создает NullOID из строки и паникует при ошибке.
func MustFromString(s string) NullOID {
	n, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return n
}

// OIDArray представляет массив OID для работы с PostgreSQL массивами.
type OIDArray []OID

// Value реализует driver.Valuer для OIDArray.
// Возвращает строку в формате PostgreSQL массива: {1.3.6.1,2.100.3}
func (oa OIDArray) Value() (driver.Value, error) {
	if len(oa) == 0 {
		return "{}", nil
	}

	// Предварительно вычисляем размер
	size := 2 // фигурные скобки
	for i, oid := range oa {
		if i > 0 {
			size++ // запятая
		}
		if err := oid.Validate(); err != nil {
			return nil, fmt.Errorf("невалидный OID в массиве на позиции %d: %w", i, err)
		}
		size += len(oid.String())
	}

	// Собираем строку
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

// Scan реализует sql.Scanner для OIDArray.
// Поддерживает PostgreSQL формат массива: {1.3.6.1,2.100.3}
func (oa *OIDArray) Scan(value any) error {
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
		return fmt.Errorf("неподдерживаемый тип для конвертации в OIDArray: %T", value)
	}

	// Проверяем формат
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("некорректный формат массива OID: %s", s)
	}

	// Пустой массив
	if s == "{}" {
		*oa = OIDArray{}
		return nil
	}

	// Парсим элементы
	content := s[1 : len(s)-1]
	parts := splitPostgresArray(content)

	result := make(OIDArray, 0, len(parts))
	for _, part := range parts {
		oid, err := ParseOID(part)
		if err != nil {
			return fmt.Errorf("ошибка парсинга OID '%s' в массиве: %w", part, err)
		}
		result = append(result, oid)
	}

	*oa = result
	return nil
}

// MarshalJSON реализует json.Marshaler для OIDArray.
func (oa OIDArray) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON реализует json.Unmarshaler для OIDArray.
func (oa *OIDArray) UnmarshalJSON(data []byte) error {
	// Пустой массив
	if string(data) == "[]" || string(data) == "null" {
		*oa = OIDArray{}
		return nil
	}

	var strArray []string
	if err := json.Unmarshal(data, &strArray); err != nil {
		return fmt.Errorf("ошибка декодирования JSON массива: %w", err)
	}

	result := make(OIDArray, 0, len(strArray))
	for _, s := range strArray {
		oid, err := ParseOID(s)
		if err != nil {
			return fmt.Errorf("ошибка парсинга OID '%s': %w", s, err)
		}
		result = append(result, oid)
	}

	*oa = result
	return nil
}

// String возвращает строковое представление OIDArray.
func (oa OIDArray) String() string {
	if len(oa) == 0 {
		return "[]"
	}

	result := "["
	for i, oid := range oa {
		if i > 0 {
			result += ", "
		}
		result += oid.String()
	}
	result += "]"

	return result
}

// Equal проверяет равенство двух OIDArray.
func (oa OIDArray) Equal(other OIDArray) bool {
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

// Contains проверяет наличие OID в массиве.
func (oa OIDArray) Contains(target OID) bool {
	for _, oid := range oa {
		if oid.Equal(target) {
			return true
		}
	}
	return false
}

// Append добавляет OID в массив.
func (oa OIDArray) Append(oids ...OID) OIDArray {
	result := make(OIDArray, len(oa), len(oa)+len(oids))
	copy(result, oa)
	return append(result, oids...)
}

// splitPostgresArray разделяет строку PostgreSQL массива на элементы.
// Учитывает экранирование и кавычки.
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
				// Убираем кавычки, если есть
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
		// Убираем кавычки, если есть
		part := string(current)
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}
		parts = append(parts, part)
	}

	return parts
}
