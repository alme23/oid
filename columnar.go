package oid

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// ColumnarOID представляет колумнарный OID (табличный)
// Формат: base.column.index1[.index2...]
type ColumnarOID struct {
	Base    OID      // Базовый OID таблицы
	Column  uint32   // Номер колонки
	Indexes []uint32 // Индексы строки
}

// NewColumnarOID создает колумнарный OID
func NewColumnarOID(base OID, column uint32, indexes ...uint32) ColumnarOID {
	return ColumnarOID{
		Base:    base,
		Column:  column,
		Indexes: indexes,
	}
}

// ParseColumnarOID парсит колумнарный OID из полного OID
func ParseColumnarOID(base, fullOID OID) (ColumnarOID, error) {
	if !fullOID.StartsWith(base) {
		return ColumnarOID{}, fmt.Errorf("%w: %s не принадлежит базе %s",
			ErrOIDNotInBase, fullOID, base)
	}

	rest := fullOID[len(base):]
	if len(rest) < 1 {
		return ColumnarOID{}, ErrNotEnoughComponents
	}

	column := rest[0]
	indexes := make([]uint32, 0, len(rest)-1)
	if len(rest) > 1 {
		indexes = append(indexes, rest[1:]...)
	}

	return ColumnarOID{
		Base:    base,
		Column:  column,
		Indexes: indexes,
	}, nil
}

// MustColumnarOID создает колумнарный OID из строк
func MustColumnarOID(baseStr string, column uint32, indexes ...uint32) ColumnarOID {
	base := MustParseOID(baseStr)
	return NewColumnarOID(base, column, indexes...)
}

// FullOID возвращает полный OID
func (c ColumnarOID) FullOID() OID {
	result := make(OID, 0, len(c.Base)+1+len(c.Indexes))
	result = append(result, c.Base...)
	result = append(result, c.Column)
	result = append(result, c.Indexes...)
	return result
}

// String возвращает строковое представление
func (c ColumnarOID) String() string {
	return c.FullOID().String()
}

// IsValid проверяет корректность
func (c ColumnarOID) IsValid() bool {
	return c.Base.Validate() == nil && len(c.Base) >= 2
}

// HasIndexes проверяет наличие индексов
func (c ColumnarOID) HasIndexes() bool {
	return len(c.Indexes) > 0
}

// IndexString возвращает индексы как строку
func (c ColumnarOID) IndexString() string {
	if len(c.Indexes) == 0 {
		return ""
	}

	parts := make([]string, len(c.Indexes))
	for i, idx := range c.Indexes {
		parts[i] = fmt.Sprintf("%d", idx)
	}
	return strings.Join(parts, ".")
}

// WithIndexes создает новый ColumnarOID с указанными индексами
func (c ColumnarOID) WithIndexes(indexes ...uint32) ColumnarOID {
	return ColumnarOID{
		Base:    c.Base,
		Column:  c.Column,
		Indexes: indexes,
	}
}

// AppendIndex добавляет индекс к существующим
func (c ColumnarOID) AppendIndex(index uint32) ColumnarOID {
	newIndexes := make([]uint32, 0, len(c.Indexes)+1)
	newIndexes = append(newIndexes, c.Indexes...)
	newIndexes = append(newIndexes, index)

	return ColumnarOID{
		Base:    c.Base,
		Column:  c.Column,
		Indexes: newIndexes,
	}
}

// Parent возвращает ColumnarOID без последнего индекса
func (c ColumnarOID) Parent() ColumnarOID {
	if len(c.Indexes) <= 1 {
		return ColumnarOID{
			Base:    c.Base,
			Column:  c.Column,
			Indexes: nil,
		}
	}
	return ColumnarOID{
		Base:    c.Base,
		Column:  c.Column,
		Indexes: c.Indexes[:len(c.Indexes)-1],
	}
}

// LastIndex возвращает последний индекс
func (c ColumnarOID) LastIndex() (uint32, error) {
	if len(c.Indexes) == 0 {
		return 0, ErrNoIndexes
	}
	return c.Indexes[len(c.Indexes)-1], nil
}

// Validate проверяет корректность
func (c ColumnarOID) Validate() error {
	if err := c.Base.Validate(); err != nil {
		return err
	}
	return nil
}

// Equal проверяет равенство
func (c ColumnarOID) Equal(other ColumnarOID) bool {
	if !c.Base.Equal(other.Base) {
		return false
	}
	if c.Column != other.Column {
		return false
	}
	if len(c.Indexes) != len(other.Indexes) {
		return false
	}
	for i := range c.Indexes {
		if c.Indexes[i] != other.Indexes[i] {
			return false
		}
	}
	return true
}

// MarshalBinary encodes to DER
func (c ColumnarOID) MarshalBinary() ([]byte, error) {
	return c.FullOID().MarshalBinary()
}

// UnmarshalBinary decodes  from DER
// IMPORTANT: base OID MUST be set
func (c *ColumnarOID) UnmarshalBinary(data []byte, base OID) error {
	var o OID
	if err := o.UnmarshalBinary(data); err != nil {
		return err
	}

	parsed, err := ParseColumnarOID(base, o)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// MarshalBER encodes to BER
func (c ColumnarOID) MarshalBER() ([]byte, error) {
	return c.FullOID().MarshalBER()
}

// MarshalJSON encodes to JSON
func (c ColumnarOID) MarshalJSON() ([]byte, error) {
	return c.FullOID().MarshalJSON()
}

// UnmarshalJSON decodes from JSON
// IMPORTANT: uses Base from current object
func (c *ColumnarOID) UnmarshalJSON(data []byte) error {
	var o OID
	if err := o.UnmarshalJSON(data); err != nil {
		return err
	}

	// if Base not set, try to use full OID
	if len(c.Base) == 0 {
		// keep as full OID
		parsed := ColumnarOID{
			Base:    o,
			Column:  0,
			Indexes: nil,
		}
		*c = parsed
		return nil
	}

	parsed, err := ParseColumnarOID(c.Base, o)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// Value реализует driver.Valuer
func (c ColumnarOID) Value() (driver.Value, error) {
	return c.FullOID().Value()
}

// Scan реализует sql.Scanner
func (c *ColumnarOID) Scan(value any) error {
	var o OID
	if err := o.Scan(value); err != nil {
		return err
	}

	if len(c.Base) == 0 {
		*c = ColumnarOID{
			Base:    o,
			Column:  0,
			Indexes: nil,
		}
		return nil
	}

	parsed, err := ParseColumnarOID(c.Base, o)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}
