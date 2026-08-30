// oid/table_test.go
package oid

import (
	"testing"
)

func TestNewTableOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)

	if table == nil {
		t.Fatal("NewTableOID: nil")
	}
	if !table.Base.Equal(base) {
		t.Error("NewTableOID: неверная база")
	}
	if len(table.Columns) != 0 {
		t.Error("NewTableOID: колонки должны быть пустыми")
	}
}

func TestTableOIDColumns(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)

	// Добавляем колонки
	table.AddColumn("ifIndex", 1)
	table.AddColumn("ifDescr", 2)
	table.AddColumn("ifType", 3)

	// Проверяем
	if len(table.Columns) != 3 {
		t.Error("AddColumn: неверное количество колонок")
	}

	// GetColumnOID
	oid, err := table.GetColumnOID("ifDescr", 1)
	if err != nil {
		t.Errorf("GetColumnOID: %v", err)
	}
	if oid.String() != "1.3.6.1.2.1.2.2.1.2.1" {
		t.Errorf("GetColumnOID = %s", oid)
	}

	// GetColumnOID с несуществующей колонкой
	_, err = table.GetColumnOID("nonexistent", 1)
	if err == nil {
		t.Error("GetColumnOID: ожидалась ошибка")
	}

	// GetColumnName
	name, exists := table.GetColumnName(2)
	if !exists || name != "ifDescr" {
		t.Errorf("GetColumnName = %s, %v", name, exists)
	}

	// GetColumnNames
	names := table.GetColumnNames()
	if len(names) != 3 {
		t.Error("GetColumnNames: неверное количество")
	}
}

func TestTableOIDRow(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)
	table.AddColumn("ifDescr", 2)
	table.AddColumn("ifType", 3)

	// GetRowOID
	row, err := table.GetRowOID(1)
	if err != nil {
		t.Errorf("GetRowOID: %v", err)
	}
	if len(row) != 2 {
		t.Error("GetRowOID: неверное количество")
	}
	if row["ifDescr"].String() != "1.3.6.1.2.1.2.2.1.2.1" {
		t.Errorf("GetRowOID[ifDescr] = %s", row["ifDescr"])
	}

	// GetRowOIDWithIndexes
	row2, err := table.GetRowOIDWithIndexes(1, 2)
	if err != nil {
		t.Errorf("GetRowOIDWithIndexes: %v", err)
	}
	if row2["ifDescr"].String() != "1.3.6.1.2.1.2.2.1.2.1.2" {
		t.Errorf("GetRowOIDWithIndexes[ifDescr] = %s", row2["ifDescr"])
	}
}

func TestTableOIDParse(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)
	table.AddColumn("ifDescr", 2)

	// ParseRowOID
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")
	column, index, err := table.ParseRowOID(fullOID)
	if err != nil {
		t.Errorf("ParseRowOID: %v", err)
	}
	if column != 2 || index != 1 {
		t.Errorf("ParseRowOID = %d, %d", column, index)
	}

	// ParseRowOIDWithIndexes
	fullOID2 := MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2")
	column2, indexes, err := table.ParseRowOIDWithIndexes(fullOID2)
	if err != nil {
		t.Errorf("ParseRowOIDWithIndexes: %v", err)
	}
	if column2 != 2 || len(indexes) != 2 {
		t.Errorf("ParseRowOIDWithIndexes = %d, %v", column2, indexes)
	}

	// Не принадлежит таблице
	notTableOID := MustParseOID("1.3.6.1.2.1.1.1.0")
	_, _, err = table.ParseRowOID(notTableOID)
	if err == nil {
		t.Error("ParseRowOID: ожидалась ошибка")
	}
}

func TestTableOIDValidate(t *testing.T) {
	// Пустая таблица
	table := NewTableOID(MustParseOID("1.3.6.1"))
	if err := table.Validate(); err == nil {
		t.Error("Validate: ожидалась ошибка для пустой таблицы")
	}

	// С колонками
	table.AddColumn("test", 1)
	if err := table.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}

	// Невалидная база
	invalidTable := NewTableOID(OID{3, 1})
	invalidTable.AddColumn("test", 1)
	if err := invalidTable.Validate(); err == nil {
		t.Error("Validate: ожидалась ошибка для невалидной базы")
	}
}

func TestTableOIDString(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)
	table.AddColumn("ifDescr", 2)
	table.AddColumn("ifType", 3)

	str := table.String()
	if str == "" {
		t.Error("String: пустая строка")
	}
	t.Logf("String: %s", str)
}

func TestTableOIDIsColumn(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	table := NewTableOID(base)
	table.AddColumn("ifDescr", 2)

	// Валидная колонка
	if !table.IsColumnOID(MustParseOID("1.3.6.1.2.1.2.2.1.2.1")) {
		t.Error("IsColumnOID: должно быть true")
	}

	// Не колонка
	if table.IsColumnOID(MustParseOID("1.3.6.1.2.1.2.2.1.5.1")) {
		t.Error("IsColumnOID: должно быть false")
	}

	// Не принадлежит таблице
	if table.IsColumnOID(MustParseOID("1.3.6.1.2.1.1.1.0")) {
		t.Error("IsColumnOID: должно быть false")
	}
}
