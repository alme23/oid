package oid

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// TableOID представляет таблицу SNMP
type TableOID struct {
	Base    OID               // Базовый OID таблицы
	Columns map[string]uint32 // Имя колонки -> номер
}

// NewTableOID создает таблицу
func NewTableOID(base OID) *TableOID {
	return &TableOID{
		Base:    base,
		Columns: make(map[string]uint32),
	}
}

// MustTableOID создает таблицу из строки
func MustTableOID(baseStr string) *TableOID {
	base := MustParseOID(baseStr)
	return NewTableOID(base)
}

// AddColumn добавляет колонку
func (t *TableOID) AddColumn(name string, column uint32) {
	t.Columns[name] = column
}

// AddColumns добавляет несколько колонок
func (t *TableOID) AddColumns(columns map[string]uint32) {
	// Используем maps.Copy вместо цикла
	maps.Copy(t.Columns, columns)
}

// GetColumnOID возвращает OID колонки с индексом
func (t *TableOID) GetColumnOID(columnName string, index uint32) (OID, error) {
	column, exists := t.Columns[columnName]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrColumnNotFound, columnName)
	}

	result := make(OID, 0, len(t.Base)+2)
	result = append(result, t.Base...)
	result = append(result, column, index)
	return result, nil
}

// GetColumnOIDWithIndexes возвращает OID колонки с несколькими индексами
func (t *TableOID) GetColumnOIDWithIndexes(columnName string, indexes ...uint32) (OID, error) {
	column, exists := t.Columns[columnName]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrColumnNotFound, columnName)
	}

	result := make(OID, 0, len(t.Base)+1+len(indexes))
	result = append(result, t.Base...)
	result = append(result, column)
	result = append(result, indexes...)
	return result, nil
}

// GetRowOID возвращает все OID колонок для индекса
func (t *TableOID) GetRowOID(index uint32) (map[string]OID, error) {
	result := make(map[string]OID)

	for name, column := range t.Columns {
		oid := make(OID, 0, len(t.Base)+2)
		oid = append(oid, t.Base...)
		oid = append(oid, column, index)
		result[name] = oid
	}

	return result, nil
}

// GetRowOIDWithIndexes возвращает все OID колонок для нескольких индексов
func (t *TableOID) GetRowOIDWithIndexes(indexes ...uint32) (map[string]OID, error) {
	result := make(map[string]OID)

	for name, column := range t.Columns {
		oid := make(OID, 0, len(t.Base)+1+len(indexes))
		oid = append(oid, t.Base...)
		oid = append(oid, column)
		oid = append(oid, indexes...)
		result[name] = oid
	}

	return result, nil
}

// ParseRowOID парсит OID строки таблицы
func (t *TableOID) ParseRowOID(fullOID OID) (column, index uint32, err error) {
	if !fullOID.StartsWith(t.Base) {
		return 0, 0, fmt.Errorf("%w: %s не принадлежит таблице %s",
			ErrNotInTable, fullOID, t.Base)
	}

	rest := fullOID[len(t.Base):]
	if len(rest) < 2 {
		return 0, 0, ErrNotEnoughComponents
	}

	return rest[0], rest[1], nil
}

// ParseRowOIDWithIndexes парсит OID с несколькими индексами
func (t *TableOID) ParseRowOIDWithIndexes(fullOID OID) (column uint32, indexes []uint32, err error) {
	if !fullOID.StartsWith(t.Base) {
		return 0, nil, fmt.Errorf("%w: %s не принадлежит таблице %s",
			ErrNotInTable, fullOID, t.Base)
	}

	rest := fullOID[len(t.Base):]
	if len(rest) < 1 {
		return 0, nil, ErrNotEnoughComponents
	}

	column = rest[0]
	if len(rest) > 1 {
		indexes = make([]uint32, 0, len(rest)-1)
		indexes = append(indexes, rest[1:]...)
	}

	return column, indexes, nil
}

// GetColumnName возвращает имя колонки по номеру
func (t *TableOID) GetColumnName(column uint32) (string, bool) {
	for name, col := range t.Columns {
		if col == column {
			return name, true
		}
	}
	return "", false
}

// GetColumnNames возвращает отсортированные имена колонок
func (t *TableOID) GetColumnNames() []string {
	names := make([]string, 0, len(t.Columns))
	for name := range t.Columns {
		names = append(names, name)
	}
	slices.Sort(names) // Используем slices.Sort вместо sort.Strings
	return names
}

// GetColumnNumbers возвращает отсортированные номера колонок
func (t *TableOID) GetColumnNumbers() []uint32 {
	numbers := make([]uint32, 0, len(t.Columns))
	for _, col := range t.Columns {
		numbers = append(numbers, col)
	}
	slices.Sort(numbers) // Используем slices.Sort вместо sort.Slice
	return numbers
}

// Validate проверяет корректность таблицы
func (t *TableOID) Validate() error {
	if err := t.Base.Validate(); err != nil {
		return err
	}
	if len(t.Columns) == 0 {
		return ErrTableEmpty
	}
	return nil
}

// String возвращает строковое представление
func (t *TableOID) String() string {
	var builder strings.Builder
	builder.WriteString(t.Base.String())
	builder.WriteString(" [")

	names := t.GetColumnNames()
	for i, name := range names {
		if i > 0 {
			builder.WriteString(", ")
		}
		// Используем fmt.Fprintf вместо WriteString(fmt.Sprintf(...))
		fmt.Fprintf(&builder, "%s=%d", name, t.Columns[name])
	}
	builder.WriteString("]")

	return builder.String()
}

// WalkRoot возвращает OID для начала walk
func (t *TableOID) WalkRoot() OID {
	return t.Base
}

// IsColumnOID проверяет, является ли OID колонкой этой таблицы
func (t *TableOID) IsColumnOID(fullOID OID) bool {
	if !fullOID.StartsWith(t.Base) {
		return false
	}

	rest := fullOID[len(t.Base):]
	if len(rest) < 1 {
		return false
	}

	_, exists := t.GetColumnName(rest[0])
	return exists
}
