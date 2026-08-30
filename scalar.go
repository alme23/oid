package oid

import (
	"database/sql/driver"
)

// ScalarOID представляет скалярный OID (заканчивается на .0)
type ScalarOID OID

// NewScalarOID создает скалярный OID (добавляет .0 если нужно)
func NewScalarOID(base OID) ScalarOID {
	if len(base) == 0 {
		return nil
	}

	// Если уже заканчивается на 0 - возвращаем как есть
	if base[len(base)-1] == 0 {
		return ScalarOID(base)
	}

	// Добавляем .0
	return ScalarOID(base.Append(0))
}

// MustScalarOID создает скалярный OID из строки
func MustScalarOID(s string) ScalarOID {
	oid := MustParseOID(s)
	return NewScalarOID(oid)
}

// ParseScalarOID парсит скалярный OID из строки
func ParseScalarOID(s string) (ScalarOID, error) {
	oid, err := ParseOID(s)
	if err != nil {
		return nil, err
	}
	return NewScalarOID(oid), nil
}

// OID возвращает базовый OID
func (s ScalarOID) OID() OID {
	return OID(s)
}

// IsScalar проверяет, является ли OID скалярным (заканчивается на .0)
func (s ScalarOID) IsScalar() bool {
	if len(s) == 0 {
		return false
	}
	return s[len(s)-1] == 0
}

// Base возвращает OID без последнего .0
func (s ScalarOID) Base() OID {
	if len(s) == 0 {
		return nil
	}
	if s[len(s)-1] == 0 {
		return OID(s[:len(s)-1])
	}
	return OID(s)
}

// String возвращает строковое представление
func (s ScalarOID) String() string {
	return OID(s).String()
}

// Validate проверяет корректность
func (s ScalarOID) Validate() error {
	return OID(s).Validate()
}

// Equal проверяет равенство
func (s ScalarOID) Equal(other ScalarOID) bool {
	return OID(s).Equal(OID(other))
}

// StartsWith проверяет, начинается ли с указанного префикса
func (s ScalarOID) StartsWith(prefix OID) bool {
	return OID(s).StartsWith(prefix)
}

// Append добавляет компоненты
func (s ScalarOID) Append(components ...uint32) ScalarOID {
	return ScalarOID(OID(s).Append(components...))
}

// Parent возвращает родительский OID
func (s ScalarOID) Parent() (OID, error) {
	return OID(s).Parent()
}

// Last возвращает последний компонент
func (s ScalarOID) Last() (uint32, error) {
	return OID(s).Last()
}

// MarshalBinary кодирует в DER
func (s ScalarOID) MarshalBinary() ([]byte, error) {
	return OID(s).MarshalBinary()
}

// UnmarshalBinary декодирует из DER
func (s *ScalarOID) UnmarshalBinary(data []byte) error {
	var o OID
	if err := o.UnmarshalBinary(data); err != nil {
		return err
	}
	*s = ScalarOID(o)
	return nil
}

// MarshalBER кодирует в BER
func (s ScalarOID) MarshalBER() ([]byte, error) {
	return OID(s).MarshalBER()
}

// UnmarshalBER декодирует из BER
func (s *ScalarOID) UnmarshalBER(data []byte) error {
	var o OID
	if err := o.UnmarshalBER(data); err != nil {
		return err
	}
	*s = ScalarOID(o)
	return nil
}

// MarshalJSON кодирует в JSON
func (s ScalarOID) MarshalJSON() ([]byte, error) {
	return OID(s).MarshalJSON()
}

// UnmarshalJSON декодирует из JSON
func (s *ScalarOID) UnmarshalJSON(data []byte) error {
	var o OID
	if err := o.UnmarshalJSON(data); err != nil {
		return err
	}
	*s = ScalarOID(o)
	return nil
}

// Value реализует driver.Valuer
func (s ScalarOID) Value() (driver.Value, error) {
	return OID(s).Value()
}

// Scan реализует sql.Scanner
func (s *ScalarOID) Scan(value any) error {
	var o OID
	if err := o.Scan(value); err != nil {
		return err
	}
	*s = ScalarOID(o)
	return nil
}
