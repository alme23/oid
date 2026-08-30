// oid/table_test.go
package oid

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestTableOIDStructure(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name         string
		table        *TableOID
		expectedBase OID
		expectedCols int
	}{
		{
			name: "Пустая таблица",
			table: &TableOID{
				Base:    base,
				Columns: make(map[string]uint32),
			},
			expectedBase: base,
			expectedCols: 0,
		},
		{
			name: "С колонками",
			table: &TableOID{
				Base: base,
				Columns: map[string]uint32{
					"ifIndex": 1,
					"ifDescr": 2,
				},
			},
			expectedBase: base,
			expectedCols: 2,
		},
		{
			name: "Nil Columns",
			table: &TableOID{
				Base:    base,
				Columns: nil,
			},
			expectedBase: base,
			expectedCols: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.table.Base.Equal(tt.expectedBase) {
				t.Errorf("Base = %v, want %v", tt.table.Base, tt.expectedBase)
			}

			if len(tt.table.Columns) != tt.expectedCols {
				t.Errorf("len(Columns) = %d, want %d",
					len(tt.table.Columns), tt.expectedCols)
			}
		})
	}
}

func TestTableOIDZeroValue(t *testing.T) {
	var table TableOID

	if table.Base != nil {
		t.Error("Base должно быть nil для нулевого значения")
	}

	if table.Columns != nil {
		t.Error("Columns должны быть nil для нулевого значения")
	}
}

func TestTableOIDCopyIndependence(t *testing.T) {
	base := MustParseOID("1.3.6.1")

	table1 := &TableOID{
		Base:    base,
		Columns: map[string]uint32{"test": 1},
	}

	// Копируем структуру
	table2 := *table1

	// Изменяем копию
	table2.Columns["test"] = 99
	table2.Base[0] = 99

	// Оригинал не должен измениться (Columns - map, копируется ссылка)
	if table1.Columns["test"] != 99 {
		t.Error("Columns должны быть общими (map)")
	}

	// Base - слайс, общий
	if table1.Base[0] != 99 {
		t.Error("Base должен быть общим (слайс)")
	}
}

// Пример использования
func ExampleTableOID() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	table := &TableOID{
		Base: base,
		Columns: map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
		},
	}

	fmt.Println(table.Base)
	fmt.Println(len(table.Columns))
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
}

// Бенчмарк
func BenchmarkTableOIDCreation(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = &TableOID{
			Base:    base,
			Columns: make(map[string]uint32),
		}
	}
}

func TestNewTableOID(t *testing.T) {
	tests := []struct {
		name     string
		base     OID
		expected *TableOID
	}{
		{
			name: "Стандартная база",
			base: OID{1, 3, 6, 1, 2, 1, 2, 2, 1},
			expected: &TableOID{
				Base:    OID{1, 3, 6, 1, 2, 1, 2, 2, 1},
				Columns: map[string]uint32{},
			},
		},
		{
			name: "Короткая база",
			base: OID{1, 3, 6, 1},
			expected: &TableOID{
				Base:    OID{1, 3, 6, 1},
				Columns: map[string]uint32{},
			},
		},
		{
			name: "Пустая база",
			base: OID{},
			expected: &TableOID{
				Base:    OID{},
				Columns: map[string]uint32{},
			},
		},
		{
			name: "Nil база",
			base: nil,
			expected: &TableOID{
				Base:    nil,
				Columns: map[string]uint32{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTableOID(tt.base)

			if result == nil {
				t.Fatal("NewTableOID: nil результат")
			}

			// Проверяем базу
			if !result.Base.Equal(tt.expected.Base) {
				t.Errorf("Base = %v, ожидалось %v", result.Base, tt.expected.Base)
			}

			// Проверяем, что Columns инициализирована
			if result.Columns == nil {
				t.Error("Columns должен быть инициализирован")
			}

			// Проверяем, что Columns пустая
			if len(result.Columns) != 0 {
				t.Errorf("len(Columns) = %d, ожидалось 0", len(result.Columns))
			}
		})
	}
}

// Тест с проверкой свойств
func TestNewTableOIDProperties(t *testing.T) {
	t.Run("Создает независимую копию базы", func(t *testing.T) {
		base := OID{1, 3, 6, 1, 2, 1, 2, 2, 1}

		table := NewTableOID(base)

		// Изменяем оригинальную базу
		base[0] = 2

		// Проверяем, что таблица не изменилась
		if table.Base[0] != 1 {
			t.Error("Таблица должна хранить независимую копию базы")
		}
	})

	t.Run("Columns можно добавлять", func(t *testing.T) {
		table := NewTableOID(OID{1, 3, 6, 1})

		table.AddColumn("test", 1)

		if len(table.Columns) != 1 {
			t.Error("AddColumn не работает")
		}
	})

	t.Run("Каждый вызов создает новый экземпляр", func(t *testing.T) {
		base := OID{1, 3, 6, 1}

		table1 := NewTableOID(base)
		table2 := NewTableOID(base)

		if table1 == table2 {
			t.Error("NewTableOID должен создавать разные экземпляры")
		}

		// Изменяем первый
		table1.AddColumn("test", 1)

		// Второй не должен измениться
		if len(table2.Columns) != 0 {
			t.Error("Экземпляры не должны влиять друг на друга")
		}
	})
}

// Тест с подтестами
func TestNewTableOIDCategories(t *testing.T) {
	t.Run("Валидные базы", func(t *testing.T) {
		bases := []OID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2, 1, 2, 2, 1},
			{2, 100, 3},
			{0, 39, 1},
		}

		for _, base := range bases {
			table := NewTableOID(base)

			if table == nil {
				t.Errorf("NewTableOID(%v): nil", base)
				continue
			}

			if !table.Base.Equal(base) {
				t.Errorf("NewTableOID(%v).Base = %v", base, table.Base)
			}

			if table.Columns == nil {
				t.Errorf("NewTableOID(%v).Columns = nil", base)
			}
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		tests := []struct {
			name string
			base OID
		}{
			{"Пустой", OID{}},
			{"Nil", nil},
			{"Один компонент", OID{1}},
			{"Два компонента", OID{1, 3}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				table := NewTableOID(tt.base)

				if table == nil {
					t.Fatal("NewTableOID: nil")
				}

				if table.Columns == nil {
					t.Error("Columns должен быть инициализирован")
				}
			})
		}
	})
}

// Тест с round trip
func TestNewTableOIDRoundTrip(t *testing.T) {
	bases := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 2, 2, 1},
		{2, 100, 3},
	}

	for _, base := range bases {
		t.Run(base.String(), func(t *testing.T) {
			table := NewTableOID(base)

			// Проверяем, что Base сохранилась
			if !table.Base.Equal(base) {
				t.Errorf("Base = %v, ожидалось %v", table.Base, base)
			}

			// Проверяем, что Columns можно использовать
			table.AddColumn("col1", 1)
			table.AddColumn("col2", 2)

			if len(table.Columns) != 2 {
				t.Errorf("len(Columns) = %d, ожидалось 2", len(table.Columns))
			}
		})
	}
}

// Пример использования
func ExampleNewTableOID() {
	// Создаем таблицу интерфейсов
	ifTable := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Добавляем колонки
	ifTable.AddColumn("ifIndex", 1)
	ifTable.AddColumn("ifDescr", 2)

	fmt.Println(ifTable.Base)
	fmt.Println(len(ifTable.Columns))
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
}

// Пример с пустой базой
func ExampleNewTableOID_empty() {
	table := NewTableOID(OID{})

	fmt.Println(table.Base)
	fmt.Println(table.Columns)
	// Output:
	//
	// map[]
}

// Бенчмарк
func BenchmarkNewTableOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	b.ReportAllocs()
	for b.Loop() {
		_ = NewTableOID(base)
	}
}

func TestMustTableOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OID
	}{
		{
			name:     "Стандартная база",
			input:    "1.3.6.1.2.1.2.2.1",
			expected: OID{1, 3, 6, 1, 2, 1, 2, 2, 1},
		},
		{
			name:     "Короткая база",
			input:    "1.3.6.1",
			expected: OID{1, 3, 6, 1},
		},
		{
			name:     "С первым 2",
			input:    "2.100.3",
			expected: OID{2, 100, 3},
		},
		{
			name:     "С первым 0",
			input:    "0.39.1",
			expected: OID{0, 39, 1},
		},
		{
			name:     "Максимальный компонент",
			input:    "1.3.268435455",
			expected: OID{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := MustTableOID(tt.input)

			if table == nil {
				t.Fatal("MustTableOID: nil результат")
			}

			if !table.Base.Equal(tt.expected) {
				t.Errorf("Base = %v, ожидалось %v", table.Base, tt.expected)
			}

			// Проверяем, что Columns инициализирована
			if table.Columns == nil {
				t.Error("Columns должен быть инициализирован")
			}

			// Проверяем, что Columns пустая
			if len(table.Columns) != 0 {
				t.Errorf("len(Columns) = %d, ожидалось 0", len(table.Columns))
			}
		})
	}
}

// Тест с паникой
func TestMustTableOIDPanic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Пустая строка", ""},
		{"Невалидный", "invalid"},
		{"Один компонент", "1"},
		{"Первый > 2", "3.1"},
		{"Второй > 39", "1.40"},
		{"Спецсимволы", "1.3#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustTableOID(%q): ожидалась паника", tt.input)
				}
			}()

			MustTableOID(tt.input)
		})
	}
}

func TestMustTableOIDProperties(t *testing.T) {
	t.Run("NewTableOID создает копию базы", func(t *testing.T) {
		// Создаем оригинальный OID
		originalBase := MustParseOID("1.3.6.1.2.1.2.2.1")

		// Создаем таблицу
		table := NewTableOID(originalBase)

		// Изменяем ОРИГИНАЛ, а не таблицу
		originalBase[0] = 99

		// Таблица должна сохранить оригинальное значение
		if table.Base[0] != 1 {
			t.Errorf("table.Base[0] = %d, ожидалось 1 (независимая копия)", table.Base[0])
		}
	})

	t.Run("MustTableOID создает копию базы", func(t *testing.T) {
		// Создаем оригинальный OID
		originalBase := MustParseOID("1.3.6.1.2.1.2.2.1")

		// Создаем таблицу из СТРОКИ (не из originalBase)
		table := MustTableOID("1.3.6.1.2.1.2.2.1")

		// Изменяем оригинальный OID
		originalBase[0] = 99

		// Таблица должна сохранить оригинальное значение
		if table.Base[0] != 1 {
			t.Errorf("table.Base[0] = %d, ожидалось 1", table.Base[0])
		}
	})

	t.Run("Эквивалентна NewTableOID + MustParseOID", func(t *testing.T) {
		input := "1.3.6.1.2.1.2.2.1"

		table1 := MustTableOID(input)
		table2 := NewTableOID(MustParseOID(input))

		if !table1.Base.Equal(table2.Base) {
			t.Error("MustTableOID и NewTableOID должны давать одинаковый результат")
		}
	})
}

// Тест с подтестами
func TestMustTableOIDCategories(t *testing.T) {
	t.Run("Валидные базы", func(t *testing.T) {
		inputs := []string{
			"1.3.6.1",
			"1.3.6.1.2.1.2.2.1",
			"2.100.3",
			"0.39.1",
		}

		for _, input := range inputs {
			table := MustTableOID(input)

			if table == nil {
				t.Errorf("MustTableOID(%q): nil", input)
				continue
			}

			if table.Columns == nil {
				t.Errorf("MustTableOID(%q): Columns nil", input)
			}
		}
	})

	t.Run("Паника при невалидных", func(t *testing.T) {
		inputs := []string{"", "invalid", "1", "3.1"}

		for _, input := range inputs {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MustTableOID(%q): ожидалась паника", input)
					}
				}()
				MustTableOID(input)
			}()
		}
	})
}

// Тест с round trip
func TestMustTableOIDRoundTrip(t *testing.T) {
	inputs := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.2.2.1",
		"2.100.3",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			table := MustTableOID(input)

			// Проверяем, что Base.String() совпадает с входом
			if table.Base.String() != input {
				t.Errorf("Base.String() = %s, ожидалось %s", table.Base.String(), input)
			}

			// Добавляем колонки и проверяем
			table.AddColumn("test", 1)

			if len(table.Columns) != 1 {
				t.Error("AddColumn не работает")
			}
		})
	}
}

// Пример использования
func ExampleMustTableOID() {
	// Создаем таблицу из строки
	ifTable := MustTableOID("1.3.6.1.2.1.2.2.1")

	// Добавляем колонки
	ifTable.AddColumn("ifIndex", 1)
	ifTable.AddColumn("ifDescr", 2)

	fmt.Println(ifTable.Base)
	fmt.Println(len(ifTable.Columns))
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
}

// Пример с паникой
func ExampleMustTableOID_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustTableOID("invalid")
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustTableOID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = MustTableOID("1.3.6.1.2.1.2.2.1")
	}
}

// Сравнение с NewTableOID
func BenchmarkMustTableOIDVsNewTableOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	b.Run("MustTableOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = MustTableOID("1.3.6.1.2.1.2.2.1")
		}
	})

	b.Run("NewTableOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = NewTableOID(base)
		}
	})
}

func TestTableOIDAddColumn(t *testing.T) {
	// Создаем таблицу один раз
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	tests := []struct {
		name       string
		columnName string
		columnNum  uint32
		wantLen    int
	}{
		{"ifIndex", "ifIndex", 1, 1},
		{"ifDescr", "ifDescr", 2, 2},
		{"ifType", "ifType", 3, 3},
		{"ifMtu", "ifMtu", 4, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table.AddColumn(tt.columnName, tt.columnNum)

			if len(table.Columns) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(table.Columns), tt.wantLen)
			}

			if table.Columns[tt.columnName] != tt.columnNum {
				t.Errorf("Columns[%q] = %d, want %d",
					tt.columnName, table.Columns[tt.columnName], tt.columnNum)
			}
		})
	}
}

// Тест с перезаписью
func TestTableOIDAddColumnOverwrite(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	// Добавляем колонку
	table.AddColumn("test", 1)

	// Проверяем
	if table.Columns["test"] != 1 {
		t.Error("Колонка не добавлена")
	}

	// Перезаписываем
	table.AddColumn("test", 2)

	// Проверяем перезапись
	if table.Columns["test"] != 2 {
		t.Errorf("Колонка не перезаписана: %d", table.Columns["test"])
	}

	// Количество не должно измениться
	if len(table.Columns) != 1 {
		t.Errorf("len = %d, ожидалось 1", len(table.Columns))
	}
}

// Тест с проверкой свойств
func TestTableOIDAddColumnProperties(t *testing.T) {
	t.Run("Добавляет уникальные колонки", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		table.AddColumn("col1", 1)
		table.AddColumn("col2", 2)
		table.AddColumn("col3", 3)

		if len(table.Columns) != 3 {
			t.Errorf("len = %d, ожидалось 3", len(table.Columns))
		}
	})

	t.Run("Одинаковые имена перезаписываются", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		table.AddColumn("same", 1)
		table.AddColumn("same", 2)

		if len(table.Columns) != 1 {
			t.Errorf("len = %d, ожидалось 1", len(table.Columns))
		}

		if table.Columns["same"] != 2 {
			t.Errorf("Columns[same] = %d, ожидалось 2", table.Columns["same"])
		}
	})

	t.Run("Одинаковые номера с разными именами", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		table.AddColumn("name1", 1)
		table.AddColumn("name2", 1)

		if len(table.Columns) != 2 {
			t.Errorf("len = %d, ожидалось 2", len(table.Columns))
		}
	})
}

// Тест с подтестами
func TestTableOIDAddColumnCategories(t *testing.T) {
	t.Run("Стандартные колонки", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

		columns := map[string]uint32{
			"ifIndex":       1,
			"ifDescr":       2,
			"ifType":        3,
			"ifMtu":         4,
			"ifSpeed":       5,
			"ifAdminStatus": 7,
			"ifOperStatus":  8,
		}

		for name, num := range columns {
			table.AddColumn(name, num)
		}

		if len(table.Columns) != len(columns) {
			t.Errorf("len = %d, ожидалось %d", len(table.Columns), len(columns))
		}

		for name, num := range columns {
			if table.Columns[name] != num {
				t.Errorf("Columns[%q] = %d, ожидалось %d", name, table.Columns[name], num)
			}
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		// Пустое имя
		table.AddColumn("", 1)
		if table.Columns[""] != 1 {
			t.Error("Пустое имя не добавлено")
		}

		// Номер 0
		table.AddColumn("zero", 0)
		if table.Columns["zero"] != 0 {
			t.Error("Номер 0 не добавлен")
		}

		// Максимальный номер
		table.AddColumn("max", MaxOIDComponent)
		if table.Columns["max"] != MaxOIDComponent {
			t.Error("Максимальный номер не добавлен")
		}
	})
}

// Тест с round trip
func TestTableOIDAddColumnRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Добавляем колонки
	table.AddColumn("ifIndex", 1)
	table.AddColumn("ifDescr", 2)

	// Проверяем через GetColumnOID
	oid, err := table.GetColumnOID("ifDescr", 1)
	if err != nil {
		t.Fatalf("GetColumnOID: %v", err)
	}

	expected := "1.3.6.1.2.1.2.2.1.2.1"
	if oid.String() != expected {
		t.Errorf("GetColumnOID = %s, ожидалось %s", oid, expected)
	}
}

// Пример использования
func ExampleTableOID_AddColumn() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Добавляем колонки
	table.AddColumn("ifIndex", 1)
	table.AddColumn("ifDescr", 2)
	table.AddColumn("ifType", 3)

	fmt.Println(len(table.Columns))
	fmt.Println(table.Columns["ifDescr"])
	// Output:
	// 3
	// 2
}

// Пример с перезаписью
func ExampleTableOID_AddColumn_overwrite() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	table.AddColumn("test", 1)
	table.AddColumn("test", 2)

	fmt.Println(table.Columns["test"])
	fmt.Println(len(table.Columns))
	// Output:
	// 2
	// 1
}

// Бенчмарк
func BenchmarkTableOIDAddColumn(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		table.AddColumn("test", 1)
	}
}

func TestTableOIDAddColumns(t *testing.T) {
	tests := []struct {
		name        string
		initial     map[string]uint32
		add         map[string]uint32
		expectedLen int
	}{
		{
			name:        "Пустая таблица",
			initial:     nil,
			add:         map[string]uint32{"ifIndex": 1, "ifDescr": 2},
			expectedLen: 2,
		},
		{
			name:        "Добавление к существующим",
			initial:     map[string]uint32{"ifIndex": 1},
			add:         map[string]uint32{"ifDescr": 2, "ifType": 3},
			expectedLen: 3,
		},
		{
			name:        "Перезапись существующих",
			initial:     map[string]uint32{"ifIndex": 1, "ifDescr": 2},
			add:         map[string]uint32{"ifIndex": 10, "ifDescr": 20},
			expectedLen: 2,
		},
		{
			name:        "Пустой map",
			initial:     map[string]uint32{"ifIndex": 1},
			add:         map[string]uint32{},
			expectedLen: 1,
		},
		{
			name:        "Nil map",
			initial:     map[string]uint32{"ifIndex": 1},
			add:         nil,
			expectedLen: 1,
		},
		{
			name:    "Много колонок",
			initial: nil,
			add: map[string]uint32{
				"ifIndex":       1,
				"ifDescr":       2,
				"ifType":        3,
				"ifMtu":         4,
				"ifSpeed":       5,
				"ifAdminStatus": 7,
				"ifOperStatus":  8,
			},
			expectedLen: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

			// Добавляем начальные колонки
			if tt.initial != nil {
				for name, num := range tt.initial {
					table.AddColumn(name, num)
				}
			}

			// Добавляем через AddColumns
			table.AddColumns(tt.add)

			// Проверяем количество
			if len(table.Columns) != tt.expectedLen {
				t.Errorf("len(Columns) = %d, ожидалось %d", len(table.Columns), tt.expectedLen)
			}

			// Проверяем, что все добавленные колонки есть
			for name, num := range tt.add {
				got, exists := table.Columns[name]
				if !exists {
					t.Errorf("Колонка '%s' не найдена", name)
					continue
				}
				if got != num {
					t.Errorf("Columns[%q] = %d, ожидалось %d", name, got, num)
				}
			}
		})
	}
}

// Тест с проверкой перезаписи
func TestTableOIDAddColumnsOverwrite(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	// Добавляем начальные колонки
	table.AddColumns(map[string]uint32{
		"col1": 1,
		"col2": 2,
	})

	// Перезаписываем
	table.AddColumns(map[string]uint32{
		"col1": 10,
		"col3": 3,
	})

	// Проверяем результаты
	if table.Columns["col1"] != 10 {
		t.Errorf("col1 = %d, ожидалось 10 (перезапись)", table.Columns["col1"])
	}
	if table.Columns["col2"] != 2 {
		t.Errorf("col2 = %d, ожидалось 2 (не изменена)", table.Columns["col2"])
	}
	if table.Columns["col3"] != 3 {
		t.Errorf("col3 = %d, ожидалось 3 (новая)", table.Columns["col3"])
	}

	// Проверяем количество
	if len(table.Columns) != 3 {
		t.Errorf("len = %d, ожидалось 3", len(table.Columns))
	}
}

// Тест с проверкой свойств
func TestTableOIDAddColumnsProperties(t *testing.T) {
	t.Run("Не изменяет исходный map", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		columns := map[string]uint32{"col1": 1, "col2": 2}
		originalLen := len(columns)

		table.AddColumns(columns)

		// Исходный map не должен измениться
		if len(columns) != originalLen {
			t.Error("Исходный map изменен")
		}

		// Изменяем исходный map
		columns["col3"] = 3

		// Таблица не должна измениться
		if len(table.Columns) != 2 {
			t.Error("Таблица должна иметь независимую копию")
		}
	})

	t.Run("Эквивалентна множеству AddColumn", func(t *testing.T) {
		columns := map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
			"ifType":  3,
		}

		table1 := NewTableOID(MustParseOID("1.3.6.1"))
		table1.AddColumns(columns)

		table2 := NewTableOID(MustParseOID("1.3.6.1"))
		for name, num := range columns {
			table2.AddColumn(name, num)
		}

		if len(table1.Columns) != len(table2.Columns) {
			t.Error("Разное количество колонок")
		}

		for name, num := range table1.Columns {
			if table2.Columns[name] != num {
				t.Errorf("Columns[%q]: %d != %d", name, num, table2.Columns[name])
			}
		}
	})
}

// Тест с подтестами
func TestTableOIDAddColumnsCategories(t *testing.T) {
	t.Run("Стандартные колонки", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

		table.AddColumns(map[string]uint32{
			"ifIndex":       1,
			"ifDescr":       2,
			"ifType":        3,
			"ifMtu":         4,
			"ifSpeed":       5,
			"ifAdminStatus": 7,
			"ifOperStatus":  8,
		})

		if len(table.Columns) != 7 {
			t.Errorf("len = %d, ожидалось 7", len(table.Columns))
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		// Пустой map
		table.AddColumns(map[string]uint32{})
		if len(table.Columns) != 0 {
			t.Error("Пустой map должен не добавлять колонок")
		}

		// Nil map
		table.AddColumns(nil)
		if len(table.Columns) != 0 {
			t.Error("Nil map должен не добавлять колонок")
		}

		// Map с пустым именем
		table.AddColumns(map[string]uint32{"": 1})
		if table.Columns[""] != 1 {
			t.Error("Пустое имя не добавлено")
		}
	})
}

// Тест с round trip
func TestTableOIDAddColumnsRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Добавляем колонки
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Проверяем через GetColumnOID
	oid, err := table.GetColumnOID("ifDescr", 1)
	if err != nil {
		t.Fatalf("GetColumnOID: %v", err)
	}

	expected := "1.3.6.1.2.1.2.2.1.2.1"
	if oid.String() != expected {
		t.Errorf("GetColumnOID = %s, ожидалось %s", oid, expected)
	}
}

// Пример использования
func ExampleTableOID_AddColumns() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Добавляем несколько колонок сразу
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	fmt.Println(len(table.Columns))
	fmt.Println(table.Columns["ifDescr"])
	// Output:
	// 3
	// 2
}

// Пример с перезаписью
func ExampleTableOID_AddColumns_overwrite() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	table.AddColumns(map[string]uint32{"col1": 1, "col2": 2})
	table.AddColumns(map[string]uint32{"col1": 10})

	fmt.Println(table.Columns["col1"])
	fmt.Println(table.Columns["col2"])
	fmt.Println(len(table.Columns))
	// Output:
	// 10
	// 2
	// 2
}

// Бенчмарк
func BenchmarkTableOIDAddColumns(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	columns := map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		table.AddColumns(columns)
	}
}

func TestTableOIDGetColumnOID(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex":       1,
		"ifDescr":       2,
		"ifType":        3,
		"ifAdminStatus": 7,
	})

	tests := []struct {
		name       string
		columnName string
		index      uint32
		expected   string
		wantErr    error
	}{
		{
			name:       "Первая колонка",
			columnName: "ifIndex",
			index:      1,
			expected:   "1.3.6.1.2.1.2.2.1.1.1",
			wantErr:    nil,
		},
		{
			name:       "Вторая колонка",
			columnName: "ifDescr",
			index:      1,
			expected:   "1.3.6.1.2.1.2.2.1.2.1",
			wantErr:    nil,
		},
		{
			name:       "Третья колонка",
			columnName: "ifType",
			index:      2,
			expected:   "1.3.6.1.2.1.2.2.1.3.2",
			wantErr:    nil,
		},
		{
			name:       "Индекс 0",
			columnName: "ifIndex",
			index:      0,
			expected:   "1.3.6.1.2.1.2.2.1.1.0",
			wantErr:    nil,
		},
		{
			name:       "Большой индекс",
			columnName: "ifIndex",
			index:      MaxOIDComponent,
			expected:   "1.3.6.1.2.1.2.2.1.1.268435455",
			wantErr:    nil,
		},
		{
			name:       "Несуществующая колонка",
			columnName: "nonexistent",
			index:      1,
			expected:   "",
			wantErr:    ErrColumnNotFound,
		},
		{
			name:       "Пустое имя колонки",
			columnName: "",
			index:      1,
			expected:   "",
			wantErr:    ErrColumnNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid, err := table.GetColumnOID(tt.columnName, tt.index)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetColumnOID: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetColumnOID = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetColumnOID: %v", err)
				return
			}

			if oid.String() != tt.expected {
				t.Errorf("GetColumnOID = %s, ожидалось %s", oid, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDGetColumnOIDProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Возвращает новый OID каждый раз", func(t *testing.T) {
		oid1, err1 := table.GetColumnOID("ifDescr", 1)
		oid2, err2 := table.GetColumnOID("ifDescr", 1)

		if err1 != nil || err2 != nil {
			t.Fatal("Ошибка GetColumnOID")
		}

		if &oid1[0] == &oid2[0] {
			t.Error("GetColumnOID должен возвращать новый OID")
		}
	})

	t.Run("Результат начинается с Base", func(t *testing.T) {
		oid, err := table.GetColumnOID("ifDescr", 1)
		if err != nil {
			t.Fatalf("GetColumnOID: %v", err)
		}

		if !oid.StartsWith(table.Base) {
			t.Error("OID должен начинаться с базы таблицы")
		}
	})

	t.Run("Длина = len(Base) + 2", func(t *testing.T) {
		oid, err := table.GetColumnOID("ifDescr", 1)
		if err != nil {
			t.Fatalf("GetColumnOID: %v", err)
		}

		expectedLen := len(table.Base) + 2
		if len(oid) != expectedLen {
			t.Errorf("len = %d, ожидалось %d", len(oid), expectedLen)
		}
	})
}

// Тест с round trip
func TestTableOIDGetColumnOIDRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Получаем OID колонки
	oid, err := table.GetColumnOID("ifDescr", 1)
	if err != nil {
		t.Fatalf("GetColumnOID: %v", err)
	}

	// Парсим обратно
	column, index, err := table.ParseRowOID(oid)
	if err != nil {
		t.Fatalf("ParseRowOID: %v", err)
	}

	if column != 2 {
		t.Errorf("column = %d, ожидалось 2", column)
	}
	if index != 1 {
		t.Errorf("index = %d, ожидалось 1", index)
	}
}

// Тест с подтестами
func TestTableOIDGetColumnOIDCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	t.Run("Существующие колонки", func(t *testing.T) {
		tests := []struct {
			columnName string
			index      uint32
			expected   string
		}{
			{"ifIndex", 1, "1.3.6.1.2.1.2.2.1.1.1"},
			{"ifDescr", 1, "1.3.6.1.2.1.2.2.1.2.1"},
			{"ifIndex", 2, "1.3.6.1.2.1.2.2.1.1.2"},
		}

		for _, tt := range tests {
			oid, err := table.GetColumnOID(tt.columnName, tt.index)
			if err != nil {
				t.Errorf("GetColumnOID(%q, %d): %v", tt.columnName, tt.index, err)
				continue
			}
			if oid.String() != tt.expected {
				t.Errorf("GetColumnOID(%q, %d) = %s, ожидалось %s",
					tt.columnName, tt.index, oid, tt.expected)
			}
		}
	})

	t.Run("Несуществующие колонки", func(t *testing.T) {
		_, err := table.GetColumnOID("nonexistent", 1)
		if !errors.Is(err, ErrColumnNotFound) {
			t.Errorf("GetColumnOID = %v, ожидалось ErrColumnNotFound", err)
		}
	})
}

// Пример использования
func ExampleTableOID_GetColumnOID() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	oid, err := table.GetColumnOID("ifDescr", 1)
	if err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1.2.1.2.2.1.2.1
}

// Пример с ошибкой
func ExampleTableOID_GetColumnOID_error() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	_, err := table.GetColumnOID("nonexistent", 1)
	fmt.Println(errors.Is(err, ErrColumnNotFound))
	// Output: true
}

// Бенчмарк
func BenchmarkTableOIDGetColumnOID(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	b.ReportAllocs()
	for b.Loop() {
		_, _ = table.GetColumnOID("ifDescr", 1)
	}
}

func TestTableOIDGetColumnOIDWithIndexes(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	tests := []struct {
		name       string
		columnName string
		indexes    []uint32
		expected   string
		wantErr    error
	}{
		{
			name:       "Один индекс",
			columnName: "ifDescr",
			indexes:    []uint32{1},
			expected:   "1.3.6.1.2.1.2.2.1.2.1",
			wantErr:    nil,
		},
		{
			name:       "Два индекса",
			columnName: "ifDescr",
			indexes:    []uint32{1, 2},
			expected:   "1.3.6.1.2.1.2.2.1.2.1.2",
			wantErr:    nil,
		},
		{
			name:       "Три индекса",
			columnName: "ifDescr",
			indexes:    []uint32{1, 2, 3},
			expected:   "1.3.6.1.2.1.2.2.1.2.1.2.3",
			wantErr:    nil,
		},
		{
			name:       "Без индексов",
			columnName: "ifDescr",
			indexes:    []uint32{},
			expected:   "1.3.6.1.2.1.2.2.1.2",
			wantErr:    nil,
		},
		{
			name:       "Nil индексы",
			columnName: "ifDescr",
			indexes:    nil,
			expected:   "1.3.6.1.2.1.2.2.1.2",
			wantErr:    nil,
		},
		{
			name:       "Индекс 0",
			columnName: "ifDescr",
			indexes:    []uint32{0},
			expected:   "1.3.6.1.2.1.2.2.1.2.0",
			wantErr:    nil,
		},
		{
			name:       "Большие индексы",
			columnName: "ifDescr",
			indexes:    []uint32{MaxOIDComponent, MaxOIDComponent},
			expected:   "1.3.6.1.2.1.2.2.1.2.268435455.268435455",
			wantErr:    nil,
		},
		{
			name:       "Несуществующая колонка",
			columnName: "nonexistent",
			indexes:    []uint32{1},
			expected:   "",
			wantErr:    ErrColumnNotFound,
		},
		{
			name:       "Пустое имя колонки",
			columnName: "",
			indexes:    []uint32{1},
			expected:   "",
			wantErr:    ErrColumnNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid, err := table.GetColumnOIDWithIndexes(tt.columnName, tt.indexes...)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetColumnOIDWithIndexes: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetColumnOIDWithIndexes = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetColumnOIDWithIndexes: %v", err)
				return
			}

			if oid.String() != tt.expected {
				t.Errorf("GetColumnOIDWithIndexes = %s, ожидалось %s", oid, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDGetColumnOIDWithIndexesProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Возвращает новый OID каждый раз", func(t *testing.T) {
		oid1, _ := table.GetColumnOIDWithIndexes("ifDescr", 1, 2)
		oid2, _ := table.GetColumnOIDWithIndexes("ifDescr", 1, 2)

		if &oid1[0] == &oid2[0] {
			t.Error("Должен возвращать новый OID")
		}
	})

	t.Run("Результат начинается с Base", func(t *testing.T) {
		oid, err := table.GetColumnOIDWithIndexes("ifDescr", 1, 2)
		if err != nil {
			t.Fatalf("GetColumnOIDWithIndexes: %v", err)
		}

		if !oid.StartsWith(table.Base) {
			t.Error("OID должен начинаться с базы таблицы")
		}
	})

	t.Run("Длина = len(Base) + 1 + len(indexes)", func(t *testing.T) {
		indexes := []uint32{1, 2, 3}
		oid, err := table.GetColumnOIDWithIndexes("ifDescr", indexes...)
		if err != nil {
			t.Fatalf("GetColumnOIDWithIndexes: %v", err)
		}

		expectedLen := len(table.Base) + 1 + len(indexes)
		if len(oid) != expectedLen {
			t.Errorf("len = %d, ожидалось %d", len(oid), expectedLen)
		}
	})
}

// Тест с round trip
func TestTableOIDGetColumnOIDWithIndexesRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Получаем OID с индексами
	oid, err := table.GetColumnOIDWithIndexes("ifDescr", 1, 2, 3)
	if err != nil {
		t.Fatalf("GetColumnOIDWithIndexes: %v", err)
	}

	// Парсим обратно
	column, indexes, err := table.ParseRowOIDWithIndexes(oid)
	if err != nil {
		t.Fatalf("ParseRowOIDWithIndexes: %v", err)
	}

	if column != 2 {
		t.Errorf("column = %d, ожидалось 2", column)
	}

	if len(indexes) != 3 {
		t.Errorf("len(indexes) = %d, ожидалось 3", len(indexes))
	}

	expectedIndexes := []uint32{1, 2, 3}
	for i, idx := range indexes {
		if idx != expectedIndexes[i] {
			t.Errorf("indexes[%d] = %d, ожидалось %d", i, idx, expectedIndexes[i])
		}
	}
}

// Тест с подтестами
func TestTableOIDGetColumnOIDWithIndexesCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Разное количество индексов", func(t *testing.T) {
		tests := []struct {
			indexes  []uint32
			expected string
		}{
			{[]uint32{}, "1.3.6.1.2.1.2.2.1.2"},
			{[]uint32{1}, "1.3.6.1.2.1.2.2.1.2.1"},
			{[]uint32{1, 2}, "1.3.6.1.2.1.2.2.1.2.1.2"},
			{[]uint32{1, 2, 3}, "1.3.6.1.2.1.2.2.1.2.1.2.3"},
		}

		for _, tt := range tests {
			oid, err := table.GetColumnOIDWithIndexes("ifDescr", tt.indexes...)
			if err != nil {
				t.Errorf("GetColumnOIDWithIndexes(%v): %v", tt.indexes, err)
				continue
			}
			if oid.String() != tt.expected {
				t.Errorf("GetColumnOIDWithIndexes(%v) = %s, ожидалось %s",
					tt.indexes, oid, tt.expected)
			}
		}
	})

	t.Run("Несуществующая колонка", func(t *testing.T) {
		_, err := table.GetColumnOIDWithIndexes("nonexistent", 1)
		if !errors.Is(err, ErrColumnNotFound) {
			t.Errorf("GetColumnOIDWithIndexes = %v, ожидалось ErrColumnNotFound", err)
		}
	})
}

// Пример использования
func ExampleTableOID_GetColumnOIDWithIndexes() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Один индекс
	oid1, _ := table.GetColumnOIDWithIndexes("ifDescr", 1)
	fmt.Println(oid1)

	// Два индекса
	oid2, _ := table.GetColumnOIDWithIndexes("ifDescr", 1, 2)
	fmt.Println(oid2)

	// Output:
	// 1.3.6.1.2.1.2.2.1.2.1
	// 1.3.6.1.2.1.2.2.1.2.1.2
}

// Пример с ошибкой
func ExampleTableOID_GetColumnOIDWithIndexes_error() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	_, err := table.GetColumnOIDWithIndexes("nonexistent", 1)
	fmt.Println(errors.Is(err, ErrColumnNotFound))
	// Output: true
}

// Бенчмарк
func BenchmarkTableOIDGetColumnOIDWithIndexes(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	b.ReportAllocs()
	for b.Loop() {
		_, _ = table.GetColumnOIDWithIndexes("ifDescr", 1, 2)
	}
}

func TestTableOIDGetRowOID(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	tests := []struct {
		name        string
		index       uint32
		expectedLen int
	}{
		{
			name:        "Индекс 1",
			index:       1,
			expectedLen: 3,
		},
		{
			name:        "Индекс 2",
			index:       2,
			expectedLen: 3,
		},
		{
			name:        "Индекс 0",
			index:       0,
			expectedLen: 3,
		},
		{
			name:        "Большой индекс",
			index:       MaxOIDComponent,
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := table.GetRowOID(tt.index)

			if err != nil {
				t.Errorf("GetRowOID: %v", err)
				return
			}

			// Проверяем количество колонок
			if len(row) != tt.expectedLen {
				t.Errorf("len(row) = %d, ожидалось %d", len(row), tt.expectedLen)
			}

			// Проверяем каждую колонку
			expectedOIDs := map[string]string{
				"ifIndex": "1.3.6.1.2.1.2.2.1.1." + uintToString(tt.index),
				"ifDescr": "1.3.6.1.2.1.2.2.1.2." + uintToString(tt.index),
				"ifType":  "1.3.6.1.2.1.2.2.1.3." + uintToString(tt.index),
			}

			for name, expectedStr := range expectedOIDs {
				oid, exists := row[name]
				if !exists {
					t.Errorf("Колонка '%s' не найдена", name)
					continue
				}

				if oid.String() != expectedStr {
					t.Errorf("row[%q] = %s, ожидалось %s", name, oid, expectedStr)
				}
			}
		})
	}
}

// Helper функция для конвертации uint32 в строку
func uintToString(v uint32) string {
	return fmt.Sprintf("%d", v)
}

// Тест с пустой таблицей
func TestTableOIDGetRowOIDEmpty(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	row, err := table.GetRowOID(1)

	if err != nil {
		t.Errorf("GetRowOID: %v", err)
	}

	if len(row) != 0 {
		t.Errorf("len(row) = %d, ожидалось 0", len(row))
	}
}

// Тест с проверкой свойств
func TestTableOIDGetRowOIDProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	t.Run("Возвращает новый map каждый раз", func(t *testing.T) {
		row1, _ := table.GetRowOID(1)
		row2, _ := table.GetRowOID(1)

		// Изменяем первый map
		row1["new"] = OID{1, 3}

		// Второй map не должен измениться
		if _, exists := row2["new"]; exists {
			t.Error("GetRowOID должен возвращать новый map")
		}
	})

	t.Run("Каждый OID начинается с Base", func(t *testing.T) {
		row, _ := table.GetRowOID(1)

		for name, oid := range row {
			if !oid.StartsWith(table.Base) {
				t.Errorf("row[%q] не начинается с базы", name)
			}
		}
	})

	t.Run("Длина каждого OID = len(Base) + 2", func(t *testing.T) {
		row, _ := table.GetRowOID(1)

		expectedLen := len(table.Base) + 2
		for name, oid := range row {
			if len(oid) != expectedLen {
				t.Errorf("row[%q] len = %d, ожидалось %d", name, len(oid), expectedLen)
			}
		}
	})
}

// Тест с round trip
func TestTableOIDGetRowOIDRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Получаем строку
	row, err := table.GetRowOID(1)
	if err != nil {
		t.Fatalf("GetRowOID: %v", err)
	}

	// Проверяем каждую колонку через ParseRowOID
	for name, oid := range row {
		column, index, err := table.ParseRowOID(oid)
		if err != nil {
			t.Errorf("ParseRowOID(%s): %v", name, err)
			continue
		}

		if index != 1 {
			t.Errorf("index = %d, ожидалось 1", index)
		}

		expectedColumn := table.Columns[name]
		if column != expectedColumn {
			t.Errorf("column = %d, ожидалось %d", column, expectedColumn)
		}
	}
}

// Тест с подтестами
func TestTableOIDGetRowOIDCategories(t *testing.T) {
	t.Run("С колонками", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
		})

		row, err := table.GetRowOID(1)
		if err != nil {
			t.Fatalf("GetRowOID: %v", err)
		}

		if len(row) != 2 {
			t.Errorf("len = %d, ожидалось 2", len(row))
		}
	})

	t.Run("Без колонок", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		row, err := table.GetRowOID(1)
		if err != nil {
			t.Fatalf("GetRowOID: %v", err)
		}

		if len(row) != 0 {
			t.Errorf("len = %d, ожидалось 0", len(row))
		}
	})

	t.Run("Разные индексы", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
		table.AddColumn("ifIndex", 1)

		for _, index := range []uint32{0, 1, 2, 100} {
			row, _ := table.GetRowOID(index)

			oid := row["ifIndex"]
			lastComponent, _ := oid.Last()

			if lastComponent != index {
				t.Errorf("Last = %d, ожидалось %d", lastComponent, index)
			}
		}
	})
}

// Пример использования
func ExampleTableOID_GetRowOID() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	row, _ := table.GetRowOID(1)

	fmt.Println(row["ifIndex"])
	fmt.Println(row["ifDescr"])
	// Output:
	// 1.3.6.1.2.1.2.2.1.1.1
	// 1.3.6.1.2.1.2.2.1.2.1
}

// Бенчмарк
func BenchmarkTableOIDGetRowOID(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	b.ReportAllocs()
	for b.Loop() {
		_, _ = table.GetRowOID(1)
	}
}

func TestTableOIDGetRowOIDWithIndexes(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	tests := []struct {
		name        string
		indexes     []uint32
		expectedLen int
	}{
		{
			name:        "Один индекс",
			indexes:     []uint32{1},
			expectedLen: 3,
		},
		{
			name:        "Два индекса",
			indexes:     []uint32{1, 2},
			expectedLen: 3,
		},
		{
			name:        "Три индекса",
			indexes:     []uint32{1, 2, 3},
			expectedLen: 3,
		},
		{
			name:        "Без индексов",
			indexes:     []uint32{},
			expectedLen: 3,
		},
		{
			name:        "Nil индексы",
			indexes:     nil,
			expectedLen: 3,
		},
		{
			name:        "Индекс 0",
			indexes:     []uint32{0},
			expectedLen: 3,
		},
		{
			name:        "Большие индексы",
			indexes:     []uint32{MaxOIDComponent, MaxOIDComponent},
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := table.GetRowOIDWithIndexes(tt.indexes...)

			if err != nil {
				t.Errorf("GetRowOIDWithIndexes: %v", err)
				return
			}

			// Проверяем количество колонок
			if len(row) != tt.expectedLen {
				t.Errorf("len(row) = %d, ожидалось %d", len(row), tt.expectedLen)
			}

			// Проверяем каждую колонку
			for name, oid := range row {
				// Проверяем, что OID начинается с базы
				if !oid.StartsWith(table.Base) {
					t.Errorf("row[%q] не начинается с базы", name)
				}

				// Проверяем длину
				expectedLen := len(table.Base) + 1 + len(tt.indexes)
				if len(oid) != expectedLen {
					t.Errorf("row[%q] len = %d, ожидалось %d",
						name, len(oid), expectedLen)
				}

				// Проверяем последние компоненты
				if len(tt.indexes) > 0 {
					lastIndexes := oid[len(oid)-len(tt.indexes):]
					for i, idx := range tt.indexes {
						if lastIndexes[i] != idx {
							t.Errorf("row[%q] index[%d] = %d, ожидалось %d",
								name, i, lastIndexes[i], idx)
						}
					}
				}
			}
		})
	}
}

// Тест с конкретными значениями
func TestTableOIDGetRowOIDWithIndexesValues(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	tests := []struct {
		name     string
		indexes  []uint32
		expected map[string]string
	}{
		{
			name:    "Один индекс",
			indexes: []uint32{1},
			expected: map[string]string{
				"ifIndex": "1.3.6.1.2.1.2.2.1.1.1",
				"ifDescr": "1.3.6.1.2.1.2.2.1.2.1",
			},
		},
		{
			name:    "Два индекса",
			indexes: []uint32{1, 2},
			expected: map[string]string{
				"ifIndex": "1.3.6.1.2.1.2.2.1.1.1.2",
				"ifDescr": "1.3.6.1.2.1.2.2.1.2.1.2",
			},
		},
		{
			name:    "Без индексов",
			indexes: []uint32{},
			expected: map[string]string{
				"ifIndex": "1.3.6.1.2.1.2.2.1.1",
				"ifDescr": "1.3.6.1.2.1.2.2.1.2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := table.GetRowOIDWithIndexes(tt.indexes...)
			if err != nil {
				t.Fatalf("GetRowOIDWithIndexes: %v", err)
			}

			for name, expectedStr := range tt.expected {
				oid, exists := row[name]
				if !exists {
					t.Errorf("Колонка '%s' не найдена", name)
					continue
				}

				if oid.String() != expectedStr {
					t.Errorf("row[%q] = %s, ожидалось %s", name, oid, expectedStr)
				}
			}
		})
	}
}

// Тест с пустой таблицей
func TestTableOIDGetRowOIDWithIndexesEmpty(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	row, err := table.GetRowOIDWithIndexes(1, 2)

	if err != nil {
		t.Errorf("GetRowOIDWithIndexes: %v", err)
	}

	if len(row) != 0 {
		t.Errorf("len(row) = %d, ожидалось 0", len(row))
	}
}

// Тест с проверкой свойств
func TestTableOIDGetRowOIDWithIndexesProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	t.Run("Возвращает новый map каждый раз", func(t *testing.T) {
		row1, _ := table.GetRowOIDWithIndexes(1)
		row2, _ := table.GetRowOIDWithIndexes(1)

		row1["new"] = OID{1, 3}

		if _, exists := row2["new"]; exists {
			t.Error("Должен возвращать новый map")
		}
	})

	t.Run("Эквивалентность с GetRowOID для одного индекса", func(t *testing.T) {
		row1, _ := table.GetRowOID(1)
		row2, _ := table.GetRowOIDWithIndexes(1)

		if len(row1) != len(row2) {
			t.Error("Разное количество колонок")
		}

		for name, oid1 := range row1 {
			oid2, exists := row2[name]
			if !exists {
				t.Errorf("Колонка '%s' не найдена", name)
				continue
			}
			if !oid1.Equal(oid2) {
				t.Errorf("row1[%q] = %v, row2[%q] = %v", name, oid1, name, oid2)
			}
		}
	})
}

// Тест с round trip
func TestTableOIDGetRowOIDWithIndexesRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Получаем строку с индексами
	row, err := table.GetRowOIDWithIndexes(1, 2, 3)
	if err != nil {
		t.Fatalf("GetRowOIDWithIndexes: %v", err)
	}

	oid := row["ifDescr"]

	// Парсим обратно
	column, indexes, err := table.ParseRowOIDWithIndexes(oid)
	if err != nil {
		t.Fatalf("ParseRowOIDWithIndexes: %v", err)
	}

	if column != 2 {
		t.Errorf("column = %d, ожидалось 2", column)
	}

	if len(indexes) != 3 {
		t.Errorf("len(indexes) = %d, ожидалось 3", len(indexes))
	}

	expected := []uint32{1, 2, 3}
	for i, idx := range indexes {
		if idx != expected[i] {
			t.Errorf("indexes[%d] = %d, ожидалось %d", i, idx, expected[i])
		}
	}
}

// Пример использования
func ExampleTableOID_GetRowOIDWithIndexes() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Один индекс
	row1, _ := table.GetRowOIDWithIndexes(1)
	fmt.Println(row1["ifDescr"])

	// Два индекса
	row2, _ := table.GetRowOIDWithIndexes(1, 2)
	fmt.Println(row2["ifDescr"])

	// Output:
	// 1.3.6.1.2.1.2.2.1.2.1
	// 1.3.6.1.2.1.2.2.1.2.1.2
}

// Бенчмарк
func BenchmarkTableOIDGetRowOIDWithIndexes(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	b.ReportAllocs()
	for b.Loop() {
		_, _ = table.GetRowOIDWithIndexes(1, 2)
	}
}

func TestTableOIDParseRowOID(t *testing.T) {
	// Создаем таблицу
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	tests := []struct {
		name       string
		fullOID    OID
		wantColumn uint32
		wantIndex  uint32
		wantErr    error
	}{
		{
			name:       "Колонка 1, индекс 1",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.1.1"),
			wantColumn: 1,
			wantIndex:  1,
			wantErr:    nil,
		},
		{
			name:       "Колонка 2, индекс 1",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			wantColumn: 2,
			wantIndex:  1,
			wantErr:    nil,
		},
		{
			name:       "Колонка 3, индекс 5",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.3.5"),
			wantColumn: 3,
			wantIndex:  5,
			wantErr:    nil,
		},
		{
			name:       "Индекс 0",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.1.0"),
			wantColumn: 1,
			wantIndex:  0,
			wantErr:    nil,
		},
		{
			name:       "Большой индекс",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.1.268435455"),
			wantColumn: 1,
			wantIndex:  MaxOIDComponent,
			wantErr:    nil,
		},
		{
			name:       "Не принадлежит таблице",
			fullOID:    MustParseOID("1.3.6.1.2.1.1.1.0"),
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotInTable,
		},
		{
			name:       "Совсем другой OID",
			fullOID:    MustParseOID("2.100.3.1.1"),
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotInTable,
		},
		{
			name:       "Только база",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1"),
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotEnoughComponents,
		},
		{
			name:       "База + 1 компонент",
			fullOID:    MustParseOID("1.3.6.1.2.1.2.2.1.1"),
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotEnoughComponents,
		},
		{
			name:       "Пустой OID",
			fullOID:    OID{},
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotInTable,
		},
		{
			name:       "Nil OID",
			fullOID:    nil,
			wantColumn: 0,
			wantIndex:  0,
			wantErr:    ErrNotInTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, index, err := table.ParseRowOID(tt.fullOID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseRowOID: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseRowOID = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRowOID: %v", err)
				return
			}

			if column != tt.wantColumn {
				t.Errorf("column = %d, ожидалось %d", column, tt.wantColumn)
			}

			if index != tt.wantIndex {
				t.Errorf("index = %d, ожидалось %d", index, tt.wantIndex)
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDParseRowOIDProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Round trip с GetColumnOID", func(t *testing.T) {
		for _, index := range []uint32{0, 1, 2, 10, 100} {
			// Создаем OID
			oid, err := table.GetColumnOID("ifDescr", index)
			if err != nil {
				t.Fatalf("GetColumnOID: %v", err)
			}

			// Парсим обратно
			column, parsedIndex, err := table.ParseRowOID(oid)
			if err != nil {
				t.Fatalf("ParseRowOID: %v", err)
			}

			if column != 2 {
				t.Errorf("column = %d, ожидалось 2", column)
			}

			if parsedIndex != index {
				t.Errorf("index = %d, ожидалось %d", parsedIndex, index)
			}
		}
	})

	t.Run("Не изменяет входной OID", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		table.ParseRowOID(oid)

		if !oid.Equal(oidCopy) {
			t.Error("ParseRowOID не должен изменять входной OID")
		}
	})
}

// Тест с round trip через GetRowOID
func TestTableOIDParseRowOIDRoundTripWithGetRow(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Получаем строку
	row, err := table.GetRowOID(1)
	if err != nil {
		t.Fatalf("GetRowOID: %v", err)
	}

	// Парсим каждую колонку
	for name, oid := range row {
		column, index, err := table.ParseRowOID(oid)
		if err != nil {
			t.Errorf("ParseRowOID(%s): %v", name, err)
			continue
		}

		if index != 1 {
			t.Errorf("index = %d, ожидалось 1", index)
		}

		expectedColumn := table.Columns[name]
		if column != expectedColumn {
			t.Errorf("column = %d, ожидалось %d", column, expectedColumn)
		}
	}
}

// Тест с подтестами
func TestTableOIDParseRowOIDCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			oid        OID
			wantColumn uint32
			wantIndex  uint32
		}{
			{MustParseOID("1.3.6.1.2.1.2.2.1.1.1"), 1, 1},
			{MustParseOID("1.3.6.1.2.1.2.2.1.2.3"), 2, 3},
			{MustParseOID("1.3.6.1.2.1.2.2.1.7.0"), 7, 0},
		}

		for _, tt := range tests {
			column, index, err := table.ParseRowOID(tt.oid)
			if err != nil {
				t.Errorf("ParseRowOID(%v): %v", tt.oid, err)
				continue
			}
			if column != tt.wantColumn || index != tt.wantIndex {
				t.Errorf("ParseRowOID(%v) = %d, %d; ожидалось %d, %d",
					tt.oid, column, index, tt.wantColumn, tt.wantIndex)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			oid     OID
			wantErr error
		}{
			{MustParseOID("1.3.6.1.2.1.1.1.0"), ErrNotInTable},
			{MustParseOID("1.3.6.1.2.1.2.2.1"), ErrNotEnoughComponents},
			{OID{}, ErrNotInTable},
		}

		for _, tt := range tests {
			_, _, err := table.ParseRowOID(tt.oid)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseRowOID(%v) = %v, ожидалось %v", tt.oid, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleTableOID_ParseRowOID() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// OID строки таблицы
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	column, index, err := table.ParseRowOID(fullOID)
	if err != nil {
		panic(err)
	}

	fmt.Println(column)
	fmt.Println(index)
	// Output:
	// 2
	// 1
}

// Пример с ошибкой
func ExampleTableOID_ParseRowOID_error() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// OID не принадлежит таблице
	_, _, err := table.ParseRowOID(MustParseOID("1.3.6.1.2.1.1.1.0"))
	fmt.Println(errors.Is(err, ErrNotInTable))
	// Output: true
}

// Бенчмарк
func BenchmarkTableOIDParseRowOID(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = table.ParseRowOID(oid)
	}
}

func TestTableOIDParseRowOIDWithIndexes(t *testing.T) {
	// Создаем таблицу
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	tests := []struct {
		name        string
		fullOID     OID
		wantColumn  uint32
		wantIndexes []uint32
		wantErr     error
	}{
		{
			name:        "Один индекс",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			wantColumn:  2,
			wantIndexes: []uint32{1},
			wantErr:     nil,
		},
		{
			name:        "Два индекса",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2"),
			wantColumn:  2,
			wantIndexes: []uint32{1, 2},
			wantErr:     nil,
		},
		{
			name:        "Три индекса",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2.3"),
			wantColumn:  2,
			wantIndexes: []uint32{1, 2, 3},
			wantErr:     nil,
		},
		{
			name:        "Без индексов",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2"),
			wantColumn:  2,
			wantIndexes: nil,
			wantErr:     nil,
		},
		{
			name:        "Индекс 0",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.0"),
			wantColumn:  2,
			wantIndexes: []uint32{0},
			wantErr:     nil,
		},
		{
			name:        "Большие индексы",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.268435455.268435455"),
			wantColumn:  2,
			wantIndexes: []uint32{MaxOIDComponent, MaxOIDComponent},
			wantErr:     nil,
		},
		{
			name:        "Не принадлежит таблице",
			fullOID:     MustParseOID("1.3.6.1.2.1.1.1.0"),
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrNotInTable,
		},
		{
			name:        "Только база",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1"),
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrNotEnoughComponents,
		},
		{
			name:        "Пустой OID",
			fullOID:     OID{},
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrNotInTable,
		},
		{
			name:        "Nil OID",
			fullOID:     nil,
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrNotInTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, indexes, err := table.ParseRowOIDWithIndexes(tt.fullOID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseRowOIDWithIndexes: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseRowOIDWithIndexes = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRowOIDWithIndexes: %v", err)
				return
			}

			if column != tt.wantColumn {
				t.Errorf("column = %d, ожидалось %d", column, tt.wantColumn)
			}

			if len(indexes) != len(tt.wantIndexes) {
				t.Errorf("len(indexes) = %d, ожидалось %d", len(indexes), len(tt.wantIndexes))
				return
			}

			for i, idx := range indexes {
				if idx != tt.wantIndexes[i] {
					t.Errorf("indexes[%d] = %d, ожидалось %d", i, idx, tt.wantIndexes[i])
				}
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDParseRowOIDWithIndexesProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Round trip с GetColumnOIDWithIndexes", func(t *testing.T) {
		testCases := [][]uint32{
			{1},
			{1, 2},
			{1, 2, 3},
			{},
			{0},
		}

		for _, indexes := range testCases {
			// Создаем OID
			oid, err := table.GetColumnOIDWithIndexes("ifDescr", indexes...)
			if err != nil {
				t.Fatalf("GetColumnOIDWithIndexes: %v", err)
			}

			// Парсим обратно
			column, parsedIndexes, err := table.ParseRowOIDWithIndexes(oid)
			if err != nil {
				t.Fatalf("ParseRowOIDWithIndexes: %v", err)
			}

			if column != 2 {
				t.Errorf("column = %d, ожидалось 2", column)
			}

			if len(parsedIndexes) != len(indexes) {
				t.Errorf("len(indexes) = %d, ожидалось %d", len(parsedIndexes), len(indexes))
				continue
			}

			for i, idx := range parsedIndexes {
				if idx != indexes[i] {
					t.Errorf("indexes[%d] = %d, ожидалось %d", i, idx, indexes[i])
				}
			}
		}
	})

	t.Run("Не изменяет входной OID", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2")
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		table.ParseRowOIDWithIndexes(oid)

		if !oid.Equal(oidCopy) {
			t.Error("ParseRowOIDWithIndexes не должен изменять входной OID")
		}
	})

	t.Run("Эквивалентность с ParseRowOID для одного индекса", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

		// ParseRowOID
		column1, index1, err1 := table.ParseRowOID(oid)
		if err1 != nil {
			t.Fatalf("ParseRowOID: %v", err1)
		}

		// ParseRowOIDWithIndexes
		column2, indexes2, err2 := table.ParseRowOIDWithIndexes(oid)
		if err2 != nil {
			t.Fatalf("ParseRowOIDWithIndexes: %v", err2)
		}

		if column1 != column2 {
			t.Error("Колонки должны совпадать")
		}

		if len(indexes2) != 1 || indexes2[0] != index1 {
			t.Error("Индексы должны совпадать")
		}
	})
}

// Тест с round trip через GetRowOIDWithIndexes
func TestTableOIDParseRowOIDWithIndexesRoundTripWithGetRow(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Получаем строку с индексами
	row, err := table.GetRowOIDWithIndexes(1, 2, 3)
	if err != nil {
		t.Fatalf("GetRowOIDWithIndexes: %v", err)
	}

	oid := row["ifDescr"]

	// Парсим обратно
	column, indexes, err := table.ParseRowOIDWithIndexes(oid)
	if err != nil {
		t.Fatalf("ParseRowOIDWithIndexes: %v", err)
	}

	if column != 2 {
		t.Errorf("column = %d, ожидалось 2", column)
	}

	expected := []uint32{1, 2, 3}
	if len(indexes) != len(expected) {
		t.Errorf("len(indexes) = %d, ожидалось %d", len(indexes), len(expected))
	}

	for i, idx := range indexes {
		if idx != expected[i] {
			t.Errorf("indexes[%d] = %d, ожидалось %d", i, idx, expected[i])
		}
	}
}

// Тест с подтестами
func TestTableOIDParseRowOIDWithIndexesCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			oid         OID
			wantColumn  uint32
			wantIndexes []uint32
		}{
			{MustParseOID("1.3.6.1.2.1.2.2.1.1.1"), 1, []uint32{1}},
			{MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2"), 2, []uint32{1, 2}},
			{MustParseOID("1.3.6.1.2.1.2.2.1.3"), 3, nil},
		}

		for _, tt := range tests {
			column, indexes, err := table.ParseRowOIDWithIndexes(tt.oid)
			if err != nil {
				t.Errorf("ParseRowOIDWithIndexes(%v): %v", tt.oid, err)
				continue
			}
			if column != tt.wantColumn {
				t.Errorf("column = %d, ожидалось %d", column, tt.wantColumn)
			}
			if len(indexes) != len(tt.wantIndexes) {
				t.Errorf("len(indexes) = %d, ожидалось %d", len(indexes), len(tt.wantIndexes))
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			oid     OID
			wantErr error
		}{
			{MustParseOID("1.3.6.1.2.1.1.1.0"), ErrNotInTable},
			{MustParseOID("1.3.6.1.2.1.2.2.1"), ErrNotEnoughComponents},
			{OID{}, ErrNotInTable},
		}

		for _, tt := range tests {
			_, _, err := table.ParseRowOIDWithIndexes(tt.oid)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseRowOIDWithIndexes(%v) = %v, ожидалось %v", tt.oid, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleTableOID_ParseRowOIDWithIndexes() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// OID с двумя индексами
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2")

	column, indexes, err := table.ParseRowOIDWithIndexes(fullOID)
	if err != nil {
		panic(err)
	}

	fmt.Println(column)
	fmt.Println(indexes)
	// Output:
	// 2
	// [1 2]
}

// Пример с ошибкой
func ExampleTableOID_ParseRowOIDWithIndexes_error() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	_, _, err := table.ParseRowOIDWithIndexes(MustParseOID("1.3.6.1.2.1.1.1.0"))
	fmt.Println(errors.Is(err, ErrNotInTable))
	// Output: true
}

// Бенчмарк
func BenchmarkTableOIDParseRowOIDWithIndexes(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2")

	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = table.ParseRowOIDWithIndexes(oid)
	}
}

func TestTableOIDGetColumnName(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
		"ifMtu":   4,
	})

	tests := []struct {
		name         string
		column       uint32
		expectedName string
		expectedOk   bool
	}{
		{
			name:         "Колонка 1",
			column:       1,
			expectedName: "ifIndex",
			expectedOk:   true,
		},
		{
			name:         "Колонка 2",
			column:       2,
			expectedName: "ifDescr",
			expectedOk:   true,
		},
		{
			name:         "Колонка 3",
			column:       3,
			expectedName: "ifType",
			expectedOk:   true,
		},
		{
			name:         "Колонка 4",
			column:       4,
			expectedName: "ifMtu",
			expectedOk:   true,
		},
		{
			name:         "Несуществующая колонка 5",
			column:       5,
			expectedName: "",
			expectedOk:   false,
		},
		{
			name:         "Несуществующая колонка 0",
			column:       0,
			expectedName: "",
			expectedOk:   false,
		},
		{
			name:         "Несуществующая колонка 100",
			column:       100,
			expectedName: "",
			expectedOk:   false,
		},
		{
			name:         "Несуществующая колонка MaxOIDComponent",
			column:       MaxOIDComponent,
			expectedName: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := table.GetColumnName(tt.column)

			if ok != tt.expectedOk {
				t.Errorf("GetColumnName(%d) ok = %v, ожидалось %v",
					tt.column, ok, tt.expectedOk)
			}

			if name != tt.expectedName {
				t.Errorf("GetColumnName(%d) = %q, ожидалось %q",
					tt.column, name, tt.expectedName)
			}
		})
	}
}

// Тест с пустой таблицей
func TestTableOIDGetColumnNameEmpty(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	name, ok := table.GetColumnName(1)

	if ok {
		t.Error("GetColumnName: не должен найти колонку в пустой таблице")
	}

	if name != "" {
		t.Errorf("GetColumnName = %q, ожидалась пустая строка", name)
	}
}

// Тест с дубликатами номеров
func TestTableOIDGetColumnNameDuplicateNumbers(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	// Два имени с одинаковым номером
	table.AddColumn("first", 1)
	table.AddColumn("second", 1)

	name, ok := table.GetColumnName(1)

	if !ok {
		t.Error("GetColumnName: должен найти колонку")
	}

	// Должен вернуть одно из имен (порядок map не гарантирован)
	if name != "first" && name != "second" {
		t.Errorf("GetColumnName = %q, ожидалось 'first' или 'second'", name)
	}
}

// Тест с проверкой свойств
func TestTableOIDGetColumnNameProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	t.Run("Round trip с AddColumn", func(t *testing.T) {
		// Для каждой колонки
		for name, num := range table.Columns {
			gotName, ok := table.GetColumnName(num)
			if !ok {
				t.Errorf("GetColumnName(%d): не найдена", num)
				continue
			}
			if gotName != name {
				t.Errorf("GetColumnName(%d) = %q, ожидалось %q", num, gotName, name)
			}
		}
	})

	t.Run("Обратный поиск работает", func(t *testing.T) {
		// Для каждого номера
		columnNums := table.GetColumnNumbers()
		for _, num := range columnNums {
			name, ok := table.GetColumnName(num)
			if !ok {
				t.Errorf("GetColumnName(%d): не найдена", num)
				continue
			}

			// Проверяем, что имя соответствует номеру
			if table.Columns[name] != num {
				t.Errorf("Columns[%q] = %d, ожидалось %d", name, table.Columns[name], num)
			}
		}
	})
}

// Тест с round trip через GetColumnNumbers
func TestTableOIDGetColumnNameRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	// Получаем все номера колонок
	numbers := table.GetColumnNumbers()

	// Для каждого номера получаем имя
	for _, num := range numbers {
		name, ok := table.GetColumnName(num)
		if !ok {
			t.Errorf("GetColumnName(%d): не найдена", num)
			continue
		}

		// Проверяем, что имя правильное
		expectedNum := table.Columns[name]
		if expectedNum != num {
			t.Errorf("Columns[%q] = %d, ожидалось %d", name, expectedNum, num)
		}
	}
}

// Тест с подтестами
func TestTableOIDGetColumnNameCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	t.Run("Существующие колонки", func(t *testing.T) {
		tests := []struct {
			column   uint32
			expected string
		}{
			{1, "ifIndex"},
			{2, "ifDescr"},
		}

		for _, tt := range tests {
			name, ok := table.GetColumnName(tt.column)
			if !ok {
				t.Errorf("GetColumnName(%d): не найдена", tt.column)
				continue
			}
			if name != tt.expected {
				t.Errorf("GetColumnName(%d) = %q, ожидалось %q",
					tt.column, name, tt.expected)
			}
		}
	})

	t.Run("Несуществующие колонки", func(t *testing.T) {
		for _, column := range []uint32{0, 3, 10, 100} {
			name, ok := table.GetColumnName(column)
			if ok {
				t.Errorf("GetColumnName(%d): не должна найти", column)
			}
			if name != "" {
				t.Errorf("GetColumnName(%d) = %q, ожидалась пустая строка", column, name)
			}
		}
	})
}

// Пример использования
func ExampleTableOID_GetColumnName() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Поиск по номеру
	name, ok := table.GetColumnName(2)
	if ok {
		fmt.Println(name)
	}

	// Несуществующая колонка
	_, ok = table.GetColumnName(99)
	fmt.Println(ok)
	// Output:
	// ifDescr
	// false
}

// Пример с пустой таблицей
func ExampleTableOID_GetColumnName_empty() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	_, ok := table.GetColumnName(1)
	fmt.Println(ok)
	// Output: false
}

// Бенчмарк
func BenchmarkTableOIDGetColumnName(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	b.ReportAllocs()
	for b.Loop() {
		_, _ = table.GetColumnName(2)
	}
}

func TestTableOIDGetColumnNames(t *testing.T) {
	tests := []struct {
		name     string
		columns  map[string]uint32
		expected []string
	}{
		{
			name:     "Пустая таблица",
			columns:  nil,
			expected: []string{},
		},
		{
			name:     "Одна колонка",
			columns:  map[string]uint32{"ifIndex": 1},
			expected: []string{"ifIndex"},
		},
		{
			name: "Две колонки",
			columns: map[string]uint32{
				"ifIndex": 1,
				"ifDescr": 2,
			},
			expected: []string{"ifDescr", "ifIndex"},
		},
		{
			name: "Три колонки",
			columns: map[string]uint32{
				"ifIndex": 1,
				"ifDescr": 2,
				"ifType":  3,
			},
			expected: []string{"ifDescr", "ifIndex", "ifType"},
		},
		{
			name: "Много колонок",
			columns: map[string]uint32{
				"ifOperStatus":  8,
				"ifAdminStatus": 7,
				"ifSpeed":       5,
				"ifMtu":         4,
				"ifType":        3,
				"ifDescr":       2,
				"ifIndex":       1,
			},
			expected: []string{
				"ifAdminStatus",
				"ifDescr",
				"ifIndex",
				"ifMtu",
				"ifOperStatus",
				"ifSpeed",
				"ifType",
			},
		},
		{
			name: "Сортировка по алфавиту",
			columns: map[string]uint32{
				"zebra":  1,
				"apple":  2,
				"mango":  3,
				"banana": 4,
			},
			expected: []string{"apple", "banana", "mango", "zebra"},
		},
		{
			name: "Сортировка с цифрами",
			columns: map[string]uint32{
				"col10": 1,
				"col2":  2,
				"col1":  3,
			},
			expected: []string{"col1", "col10", "col2"},
		},
		{
			name: "Пустые имена",
			columns: map[string]uint32{
				"":     1,
				"test": 2,
			},
			expected: []string{"", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTableOID(MustParseOID("1.3.6.1"))

			if tt.columns != nil {
				table.AddColumns(tt.columns)
			}

			names := table.GetColumnNames()

			// Проверяем длину
			if len(names) != len(tt.expected) {
				t.Errorf("len = %d, ожидалось %d", len(names), len(tt.expected))
				return
			}

			// Проверяем каждый элемент
			for i, name := range names {
				if name != tt.expected[i] {
					t.Errorf("names[%d] = %q, ожидалось %q", i, name, tt.expected[i])
				}
			}

			// Проверяем, что список отсортирован
			if !slices.IsSorted(names) {
				t.Error("Имена должны быть отсортированы")
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDGetColumnNamesProperties(t *testing.T) {
	t.Run("Возвращает новый слайс каждый раз", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumn("test", 1)

		names1 := table.GetColumnNames()
		names2 := table.GetColumnNames()

		// Изменяем первый слайс
		names1[0] = "modified"

		// Второй не должен измениться
		if names2[0] != "test" {
			t.Error("GetColumnNames должен возвращать новый слайс")
		}
	})

	t.Run("Всегда отсортирован", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"z": 1,
			"a": 2,
			"m": 3,
			"b": 4,
		})

		names := table.GetColumnNames()

		if !slices.IsSorted(names) {
			t.Error("Имена должны быть отсортированы")
		}
	})

	t.Run("Соответствует ключам Columns", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
			"ifType":  3,
		})

		names := table.GetColumnNames()

		if len(names) != len(table.Columns) {
			t.Errorf("len = %d, ожидалось %d", len(names), len(table.Columns))
		}

		// Проверяем, что все имена из Columns есть в result
		for name := range table.Columns {
			found := false
			for _, n := range names {
				if n == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Имя %q не найдено в результате", name)
			}
		}
	})
}

// Тест с round trip
func TestTableOIDGetColumnNamesRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	// Получаем имена
	names := table.GetColumnNames()

	// Проверяем, что каждое имя можно использовать
	for _, name := range names {
		// Проверяем через GetColumnOID
		oid, err := table.GetColumnOID(name, 1)
		if err != nil {
			t.Errorf("GetColumnOID(%q): %v", name, err)
			continue
		}

		if len(oid) == 0 {
			t.Errorf("GetColumnOID(%q): пустой OID", name)
		}

		// Проверяем через Columns map
		columnNum, exists := table.Columns[name]
		if !exists {
			t.Errorf("Columns[%q] не найдена", name)
			continue
		}

		// Проверяем, что OID содержит правильный номер колонки
		if len(oid) >= len(table.Base)+1 {
			if oid[len(table.Base)] != columnNum {
				t.Errorf("OID колонки = %v, ожидался номер %d", oid, columnNum)
			}
		}
	}
}

// Тест с подтестами
func TestTableOIDGetColumnNamesCategories(t *testing.T) {
	t.Run("Пустая таблица", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		names := table.GetColumnNames()

		if len(names) != 0 {
			t.Errorf("len = %d, ожидалось 0", len(names))
		}
	})

	t.Run("С колонками", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
		})

		names := table.GetColumnNames()

		if len(names) != 2 {
			t.Errorf("len = %d, ожидалось 2", len(names))
		}

		if !slices.IsSorted(names) {
			t.Error("Имена должны быть отсортированы")
		}
	})
}

// Пример использования
func ExampleTableOID_GetColumnNames() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	names := table.GetColumnNames()

	for _, name := range names {
		fmt.Println(name)
	}
	// Output:
	// ifDescr
	// ifIndex
	// ifType
}

// Пример с пустой таблицей
func ExampleTableOID_GetColumnNames_empty() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	names := table.GetColumnNames()

	fmt.Println(len(names))
	// Output: 0
}

// Бенчмарк
func BenchmarkTableOIDGetColumnNames(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	b.ReportAllocs()
	for b.Loop() {
		_ = table.GetColumnNames()
	}
}

func TestTableOIDGetColumnNumbers(t *testing.T) {
	tests := []struct {
		name     string
		columns  map[string]uint32
		expected []uint32
	}{
		{
			name:     "Пустая таблица",
			columns:  nil,
			expected: []uint32{},
		},
		{
			name:     "Одна колонка",
			columns:  map[string]uint32{"ifIndex": 1},
			expected: []uint32{1},
		},
		{
			name: "Две колонки",
			columns: map[string]uint32{
				"ifIndex": 1,
				"ifDescr": 2,
			},
			expected: []uint32{1, 2},
		},
		{
			name: "Три колонки не по порядку",
			columns: map[string]uint32{
				"ifType":  3,
				"ifIndex": 1,
				"ifDescr": 2,
			},
			expected: []uint32{1, 2, 3},
		},
		{
			name: "Много колонок",
			columns: map[string]uint32{
				"ifOperStatus":  8,
				"ifAdminStatus": 7,
				"ifSpeed":       5,
				"ifMtu":         4,
				"ifType":        3,
				"ifDescr":       2,
				"ifIndex":       1,
			},
			expected: []uint32{1, 2, 3, 4, 5, 7, 8},
		},
		{
			name: "С нулем",
			columns: map[string]uint32{
				"zero": 0,
				"one":  1,
				"two":  2,
			},
			expected: []uint32{0, 1, 2},
		},
		{
			name: "С большими числами",
			columns: map[string]uint32{
				"small": 1,
				"big":   MaxOIDComponent,
				"mid":   1000,
			},
			expected: []uint32{1, 1000, MaxOIDComponent},
		},
		{
			name: "Дубликаты номеров",
			columns: map[string]uint32{
				"first":  1,
				"second": 1,
				"third":  2,
			},
			expected: []uint32{1, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTableOID(MustParseOID("1.3.6.1"))

			if tt.columns != nil {
				table.AddColumns(tt.columns)
			}

			numbers := table.GetColumnNumbers()

			// Проверяем длину
			if len(numbers) != len(tt.expected) {
				t.Errorf("len = %d, ожидалось %d", len(numbers), len(tt.expected))
				return
			}

			// Проверяем каждый элемент
			for i, num := range numbers {
				if num != tt.expected[i] {
					t.Errorf("numbers[%d] = %d, ожидалось %d", i, num, tt.expected[i])
				}
			}

			// Проверяем, что список отсортирован
			if !slices.IsSorted(numbers) {
				t.Error("Номера должны быть отсортированы")
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDGetColumnNumbersProperties(t *testing.T) {
	t.Run("Возвращает новый слайс каждый раз", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumn("test", 1)

		numbers1 := table.GetColumnNumbers()
		numbers2 := table.GetColumnNumbers()

		// Изменяем первый слайс
		numbers1[0] = 99

		// Второй не должен измениться
		if numbers2[0] != 1 {
			t.Error("GetColumnNumbers должен возвращать новый слайс")
		}
	})

	t.Run("Всегда отсортирован", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"z": 100,
			"a": 1,
			"m": 50,
			"b": 10,
		})

		numbers := table.GetColumnNumbers()

		if !slices.IsSorted(numbers) {
			t.Error("Номера должны быть отсортированы")
		}
	})

	t.Run("Соответствует значениям Columns", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
			"ifType":  3,
		})

		numbers := table.GetColumnNumbers()

		if len(numbers) != len(table.Columns) {
			t.Errorf("len = %d, ожидалось %d", len(numbers), len(table.Columns))
		}

		// Проверяем, что все номера из Columns есть в result
		for _, colNum := range table.Columns {
			found := false
			for _, n := range numbers {
				if n == colNum {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Номер %d не найден в результате", colNum)
			}
		}
	})
}

// Тест с round trip
func TestTableOIDGetColumnNumbersRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	// Получаем номера
	numbers := table.GetColumnNumbers()

	// Проверяем, что каждый номер можно использовать
	for _, num := range numbers {
		// Проверяем через GetColumnName
		name, ok := table.GetColumnName(num)
		if !ok {
			t.Errorf("GetColumnName(%d): не найдена", num)
			continue
		}

		// Проверяем, что имя соответствует номеру
		if table.Columns[name] != num {
			t.Errorf("Columns[%q] = %d, ожидалось %d", name, table.Columns[name], num)
		}
	}
}

// Тест с подтестами
func TestTableOIDGetColumnNumbersCategories(t *testing.T) {
	t.Run("Пустая таблица", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		numbers := table.GetColumnNumbers()

		if len(numbers) != 0 {
			t.Errorf("len = %d, ожидалось 0", len(numbers))
		}
	})

	t.Run("С колонками", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
		})

		numbers := table.GetColumnNumbers()

		if len(numbers) != 2 {
			t.Errorf("len = %d, ожидалось 2", len(numbers))
		}

		if !slices.IsSorted(numbers) {
			t.Error("Номера должны быть отсортированы")
		}
	})
}

// Пример использования
func ExampleTableOID_GetColumnNumbers() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	numbers := table.GetColumnNumbers()

	for _, num := range numbers {
		fmt.Println(num)
	}
	// Output:
	// 1
	// 2
	// 3
}

// Пример с пустой таблицей
func ExampleTableOID_GetColumnNumbers_empty() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	numbers := table.GetColumnNumbers()

	fmt.Println(len(numbers))
	// Output: 0
}

// Бенчмарк
func BenchmarkTableOIDGetColumnNumbers(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	b.ReportAllocs()
	for b.Loop() {
		_ = table.GetColumnNumbers()
	}
}

func TestTableOIDValidate(t *testing.T) {
	tests := []struct {
		name    string
		table   *TableOID
		wantErr error
	}{
		{
			name: "Валидная таблица",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
				t.AddColumn("ifIndex", 1)
				return t
			}(),
			wantErr: nil,
		},
		{
			name: "Валидная с несколькими колонками",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
				t.AddColumns(map[string]uint32{
					"ifIndex": 1,
					"ifDescr": 2,
					"ifType":  3,
				})
				return t
			}(),
			wantErr: nil,
		},
		{
			name: "Пустая таблица без колонок",
			table: func() *TableOID {
				return NewTableOID(MustParseOID("1.3.6.1"))
			}(),
			wantErr: ErrTableEmpty,
		},
		{
			name: "Невалидная база",
			table: func() *TableOID {
				t := NewTableOID(OID{3, 1})
				t.AddColumn("test", 1)
				return t
			}(),
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name: "Короткая база",
			table: func() *TableOID {
				t := NewTableOID(OID{1})
				t.AddColumn("test", 1)
				return t
			}(),
			wantErr: ErrOIDTooShort,
		},
		{
			name: "Пустая база",
			table: func() *TableOID {
				t := NewTableOID(OID{})
				t.AddColumn("test", 1)
				return t
			}(),
			wantErr: ErrOIDTooShort,
		},
		{
			name: "Nil база с колонками",
			table: func() *TableOID {
				t := NewTableOID(nil)
				t.AddColumn("test", 1)
				return t
			}(),
			wantErr: ErrOIDTooShort,
		},
		{
			name: "База с вторым компонентом > 39",
			table: func() *TableOID {
				t := NewTableOID(OID{1, 40})
				t.AddColumn("test", 1)
				return t
			}(),
			wantErr: ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.table.Validate()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate: %v", err)
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDValidateProperties(t *testing.T) {
	t.Run("Порядок проверок: база, затем колонки", func(t *testing.T) {
		// Невалидная база И пустые колонки
		table := NewTableOID(OID{3, 1})

		err := table.Validate()

		// Должна вернуться ошибка базы, а не пустой таблицы
		if !errors.Is(err, ErrFirstComponentTooBig) {
			t.Errorf("Validate = %v, ожидалось ErrFirstComponentTooBig", err)
		}
	})

	t.Run("Валидная таблица проходит", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumn("test", 1)

		if err := table.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
	})

	t.Run("После добавления колонок проходит", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		// Без колонок - ошибка
		if err := table.Validate(); !errors.Is(err, ErrTableEmpty) {
			t.Errorf("Validate = %v, ожидалось ErrTableEmpty", err)
		}

		// Добавляем колонку
		table.AddColumn("test", 1)

		// Теперь проходит
		if err := table.Validate(); err != nil {
			t.Errorf("Validate после AddColumn: %v", err)
		}
	})
}

// Тест с подтестами
func TestTableOIDValidateCategories(t *testing.T) {
	t.Run("Валидные таблицы", func(t *testing.T) {
		tables := []*TableOID{
			func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("test", 1)
				return t
			}(),
			func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
				t.AddColumns(map[string]uint32{"ifIndex": 1, "ifDescr": 2})
				return t
			}(),
		}

		for i, table := range tables {
			if err := table.Validate(); err != nil {
				t.Errorf("Table %d: %v", i, err)
			}
		}
	})

	t.Run("Невалидные базы", func(t *testing.T) {
		tests := []struct {
			base    OID
			wantErr error
		}{
			{OID{}, ErrOIDTooShort},
			{OID{1}, ErrOIDTooShort},
			{OID{3, 1}, ErrFirstComponentTooBig},
			{OID{1, 40}, ErrSecondComponentTooBig},
		}

		for _, tt := range tests {
			table := NewTableOID(tt.base)
			table.AddColumn("test", 1)

			err := table.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%v) = %v, ожидалось %v", tt.base, err, tt.wantErr)
			}
		}
	})

	t.Run("Пустые колонки", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		err := table.Validate()
		if !errors.Is(err, ErrTableEmpty) {
			t.Errorf("Validate = %v, ожидалось ErrTableEmpty", err)
		}
	})
}

// Тест с round trip
func TestTableOIDValidateRoundTrip(t *testing.T) {
	// Создаем таблицу
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	// Изначально невалидна (нет колонок)
	if err := table.Validate(); !errors.Is(err, ErrTableEmpty) {
		t.Errorf("Validate = %v, ожидалось ErrTableEmpty", err)
	}

	// Добавляем колонки
	table.AddColumn("ifIndex", 1)

	// Теперь валидна
	if err := table.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}

	// Удаляем колонки (очищаем map)
	table.Columns = make(map[string]uint32)

	// Снова невалидна
	if err := table.Validate(); !errors.Is(err, ErrTableEmpty) {
		t.Errorf("Validate = %v, ожидалось ErrTableEmpty", err)
	}
}

// Пример использования
func ExampleTableOID_Validate() {
	// Валидная таблица
	validTable := NewTableOID(MustParseOID("1.3.6.1"))
	validTable.AddColumn("test", 1)

	fmt.Println(validTable.Validate() == nil)

	// Невалидная (нет колонок)
	invalidTable := NewTableOID(MustParseOID("1.3.6.1"))

	fmt.Println(errors.Is(invalidTable.Validate(), ErrTableEmpty))
	// Output:
	// true
	// true
}

// Пример с невалидной базой
func ExampleTableOID_Validate_invalidBase() {
	table := NewTableOID(OID{3, 1})
	table.AddColumn("test", 1)

	err := table.Validate()
	fmt.Println(errors.Is(err, ErrFirstComponentTooBig))
	// Output: true
}

// Бенчмарк
func BenchmarkTableOIDValidate(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifIndex", 1)

	b.ReportAllocs()
	for b.Loop() {
		_ = table.Validate()
	}
}

func TestTableOIDString(t *testing.T) {
	tests := []struct {
		name     string
		table    *TableOID
		expected string
	}{
		{
			name: "Пустая таблица",
			table: func() *TableOID {
				return NewTableOID(MustParseOID("1.3.6.1"))
			}(),
			expected: "1.3.6.1 []",
		},
		{
			name: "Одна колонка",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("ifIndex", 1)
				return t
			}(),
			expected: "1.3.6.1 [ifIndex=1]",
		},
		{
			name: "Две колонки",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("ifIndex", 1)
				t.AddColumn("ifDescr", 2)
				return t
			}(),
			expected: "1.3.6.1 [ifDescr=2, ifIndex=1]",
		},
		{
			name: "Три колонки",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumns(map[string]uint32{
					"ifType":  3,
					"ifIndex": 1,
					"ifDescr": 2,
				})
				return t
			}(),
			expected: "1.3.6.1 [ifDescr=2, ifIndex=1, ifType=3]",
		},
		{
			name: "Стандартная таблица интерфейсов",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
				t.AddColumns(map[string]uint32{
					"ifIndex":       1,
					"ifDescr":       2,
					"ifType":        3,
					"ifAdminStatus": 7,
				})
				return t
			}(),
			expected: "1.3.6.1.2.1.2.2.1 [ifAdminStatus=7, ifDescr=2, ifIndex=1, ifType=3]",
		},
		{
			name: "Пустое имя колонки",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("", 1)
				return t
			}(),
			expected: "1.3.6.1 [=1]",
		},
		{
			name: "Номер 0",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("zero", 0)
				return t
			}(),
			expected: "1.3.6.1 [zero=0]",
		},
		{
			name: "Большой номер",
			table: func() *TableOID {
				t := NewTableOID(MustParseOID("1.3.6.1"))
				t.AddColumn("max", MaxOIDComponent)
				return t
			}(),
			expected: "1.3.6.1 [max=268435455]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.table.String()

			if result != tt.expected {
				t.Errorf("String() = %q, ожидалось %q", result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestTableOIDStringProperties(t *testing.T) {
	t.Run("Начинается с базы", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
		table.AddColumn("ifIndex", 1)

		str := table.String()

		if !strings.HasPrefix(str, "1.3.6.1.2.1.2.2.1") {
			t.Errorf("String() = %q, должно начинаться с базы", str)
		}
	})

	t.Run("Содержит все колонки", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"ifIndex": 1,
			"ifDescr": 2,
			"ifType":  3,
		})

		str := table.String()

		for name := range table.Columns {
			if !strings.Contains(str, name) {
				t.Errorf("String() не содержит колонку %q", name)
			}
		}
	})

	t.Run("Колонки отсортированы", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumns(map[string]uint32{
			"zebra": 1,
			"apple": 2,
			"mango": 3,
		})

		str := table.String()

		// Проверяем порядок
		applePos := strings.Index(str, "apple")
		mangoPos := strings.Index(str, "mango")
		zebraPos := strings.Index(str, "zebra")

		if applePos > mangoPos || mangoPos > zebraPos {
			t.Errorf("Колонки не отсортированы: %q", str)
		}
	})

	t.Run("Не изменяет таблицу", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumn("test", 1)

		before := table.Columns["test"]
		_ = table.String()
		after := table.Columns["test"]

		if before != after {
			t.Error("String() не должен изменять таблицу")
		}
	})
}

// Тест с round trip
func TestTableOIDStringRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Получаем строку
	str := table.String()

	// Проверяем, что строка не пустая
	if str == "" {
		t.Fatal("String() пустая")
	}

	// Проверяем, что содержит базу
	if !strings.Contains(str, "1.3.6.1.2.1.2.2.1") {
		t.Error("String() не содержит базу")
	}

	// Проверяем, что содержит колонки
	if !strings.Contains(str, "ifIndex=1") {
		t.Error("String() не содержит ifIndex=1")
	}
	if !strings.Contains(str, "ifDescr=2") {
		t.Error("String() не содержит ifDescr=2")
	}
}

// Тест с подтестами
func TestTableOIDStringCategories(t *testing.T) {
	t.Run("Пустая таблица", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))

		str := table.String()

		if str != "1.3.6.1 []" {
			t.Errorf("String() = %q, ожидалось '1.3.6.1 []'", str)
		}
	})

	t.Run("С колонками", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1"))
		table.AddColumn("test", 1)

		str := table.String()

		if str != "1.3.6.1 [test=1]" {
			t.Errorf("String() = %q, ожидалось '1.3.6.1 [test=1]'", str)
		}
	})
}

// Пример использования
func ExampleTableOID_String() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	fmt.Println(table.String())
	// Output: 1.3.6.1.2.1.2.2.1 [ifDescr=2, ifIndex=1]
}

// Пример с пустой таблицей
func ExampleTableOID_String_empty() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	fmt.Println(table.String())
	// Output: 1.3.6.1 []
}

// Бенчмарк
func BenchmarkTableOIDString(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	b.ReportAllocs()
	for b.Loop() {
		_ = table.String()
	}
}

func TestTableOIDWalkRoot(t *testing.T) {
	tests := []struct {
		name     string
		base     OID
		expected string
	}{
		{
			name:     "Стандартная база",
			base:     MustParseOID("1.3.6.1.2.1.2.2.1"),
			expected: "1.3.6.1.2.1.2.2.1",
		},
		{
			name:     "Короткая база",
			base:     MustParseOID("1.3.6.1"),
			expected: "1.3.6.1",
		},
		{
			name:     "База с первым 2",
			base:     MustParseOID("2.100.3"),
			expected: "2.100.3",
		},
		{
			name:     "База с первым 0",
			base:     MustParseOID("0.39.1"),
			expected: "0.39.1",
		},
		{
			name:     "Пустая база",
			base:     OID{},
			expected: "",
		},
		{
			name:     "Nil база",
			base:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTableOID(tt.base)

			result := table.WalkRoot()

			if result.String() != tt.expected {
				t.Errorf("WalkRoot() = %q, ожидалось %q", result.String(), tt.expected)
			}

			// Проверяем длину
			if len(result) != len(tt.base) {
				t.Errorf("len(WalkRoot()) = %d, ожидалось %d", len(result), len(tt.base))
			}
		})
	}
}

// Тест с проверкой копирования
func TestTableOIDWalkRootCopy(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	t.Run("Возвращает копию, а не ссылку", func(t *testing.T) {
		// Получаем WalkRoot
		root := table.WalkRoot()

		// Сохраняем оригинальное значение
		originalFirst := table.Base[0]

		// Изменяем результат
		root[0] = 99

		// Таблица не должна измениться
		if table.Base[0] != originalFirst {
			t.Error("WalkRoot должен возвращать копию")
		}
	})

	t.Run("Каждый вызов возвращает новую копию", func(t *testing.T) {
		root1 := table.WalkRoot()
		root2 := table.WalkRoot()

		// Изменяем первый
		root1[0] = 99

		// Второй не должен измениться
		if root2[0] != 1 {
			t.Error("Каждый вызов WalkRoot должен возвращать новую копию")
		}
	})

	t.Run("Изменение результата не влияет на таблицу", func(t *testing.T) {
		root := table.WalkRoot()

		// Изменяем все элементы
		for i := range root {
			root[i] = 99
		}

		// Таблица не должна измениться
		if table.Base[0] != 1 {
			t.Error("Изменение WalkRoot не должно влиять на таблицу")
		}
	})
}

// Тест с проверкой свойств
func TestTableOIDWalkRootProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	t.Run("WalkRoot равен Base", func(t *testing.T) {
		root := table.WalkRoot()

		if !root.Equal(table.Base) {
			t.Error("WalkRoot должен быть равен Base")
		}
	})

	t.Run("WalkRoot можно использовать как OID", func(t *testing.T) {
		root := table.WalkRoot()

		// Проверяем, что это валидный OID
		if err := root.Validate(); err != nil {
			t.Errorf("WalkRoot: невалидный OID: %v", err)
		}

		// Можно использовать StartsWith
		if !root.StartsWith(MustParseOID("1.3.6.1")) {
			t.Error("WalkRoot должен начинаться с 1.3.6.1")
		}
	})
}

// Тест с round trip
func TestTableOIDWalkRootRoundTrip(t *testing.T) {
	bases := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 2, 2, 1},
		{2, 100, 3},
	}

	for _, base := range bases {
		t.Run(base.String(), func(t *testing.T) {
			table := NewTableOID(base)

			root := table.WalkRoot()

			// Проверяем, что root равен базе
			if !root.Equal(base) {
				t.Errorf("WalkRoot() = %v, ожидалось %v", root, base)
			}

			// Проверяем, что root можно использовать для walk
			if len(root) == 0 {
				t.Error("WalkRoot не должен быть пустым")
			}
		})
	}
}

// Тест с подтестами
func TestTableOIDWalkRootCategories(t *testing.T) {
	t.Run("С валидной базой", func(t *testing.T) {
		table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

		root := table.WalkRoot()

		if root.String() != "1.3.6.1.2.1.2.2.1" {
			t.Errorf("WalkRoot = %s", root)
		}
	})

	t.Run("С пустой базой", func(t *testing.T) {
		table := NewTableOID(OID{})

		root := table.WalkRoot()

		if len(root) != 0 {
			t.Errorf("len = %d, ожидалось 0", len(root))
		}
	})

	t.Run("С nil базой", func(t *testing.T) {
		table := NewTableOID(nil)

		root := table.WalkRoot()

		if len(root) != 0 {
			t.Errorf("len = %d, ожидалось 0", len(root))
		}
	})
}

// Пример использования
func ExampleTableOID_WalkRoot() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	root := table.WalkRoot()

	fmt.Println(root)
	// Output: 1.3.6.1.2.1.2.2.1
}

// Пример с копированием
func ExampleTableOID_WalkRoot_copy() {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	root := table.WalkRoot()
	root = root.Append(99)

	// Таблица не изменилась
	fmt.Println(table.Base)
	fmt.Println(root)
	// Output:
	// 1.3.6.1
	// 1.3.6.1.99
}

// Бенчмарк
func BenchmarkTableOIDWalkRoot(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))

	b.ReportAllocs()
	for b.Loop() {
		_ = table.WalkRoot()
	}
}

func TestTableOIDIsColumnOID(t *testing.T) {
	// Создаем таблицу с колонками
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
		"ifType":  3,
	})

	tests := []struct {
		name     string
		fullOID  OID
		expected bool
	}{
		{
			name:     "Колонка 1 с индексом",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.1.1"),
			expected: true,
		},
		{
			name:     "Колонка 2 с индексом",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			expected: true,
		},
		{
			name:     "Колонка 3 с индексом",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.3.5"),
			expected: true,
		},
		{
			name:     "Колонка без индекса",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.1"),
			expected: true,
		},
		{
			name:     "Колонка с индексом 0",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.1.0"),
			expected: true,
		},
		{
			name:     "Колонка с несколькими индексами",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.1.1.2.3"),
			expected: true,
		},
		{
			name:     "Несуществующая колонка 4",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.4.1"),
			expected: false,
		},
		{
			name:     "Несуществующая колонка 99",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1.99.1"),
			expected: false,
		},
		{
			name:     "Только база",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2.1"),
			expected: false,
		},
		{
			name:     "Не принадлежит таблице",
			fullOID:  MustParseOID("1.3.6.1.2.1.1.1.0"),
			expected: false,
		},
		{
			name:     "Совсем другой OID",
			fullOID:  MustParseOID("2.100.3.1.1"),
			expected: false,
		},
		{
			name:     "Пустой OID",
			fullOID:  OID{},
			expected: false,
		},
		{
			name:     "Nil OID",
			fullOID:  nil,
			expected: false,
		},
		{
			name:     "Короче базы",
			fullOID:  MustParseOID("1.3.6.1.2.1.2.2"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := table.IsColumnOID(tt.fullOID)

			if result != tt.expected {
				t.Errorf("IsColumnOID(%v) = %v, ожидалось %v",
					tt.fullOID, result, tt.expected)
			}
		})
	}
}

// Тест с пустой таблицей
func TestTableOIDIsColumnOIDEmpty(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1"))

	// Любой OID не должен быть колонкой
	tests := []OID{
		MustParseOID("1.3.6.1.1.1"),
		MustParseOID("1.3.6.1.2.1"),
		MustParseOID("1.3.6.1"),
	}

	for _, oid := range tests {
		if table.IsColumnOID(oid) {
			t.Errorf("IsColumnOID(%v) = true, ожидалось false (пустая таблица)", oid)
		}
	}
}

// Тест с проверкой свойств
func TestTableOIDIsColumnOIDProperties(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	t.Run("Все OID от GetColumnOID являются колонками", func(t *testing.T) {
		for name := range table.Columns {
			for _, index := range []uint32{0, 1, 2} {
				oid, err := table.GetColumnOID(name, index)
				if err != nil {
					t.Fatalf("GetColumnOID: %v", err)
				}

				if !table.IsColumnOID(oid) {
					t.Errorf("IsColumnOID(%v) = false, ожидалось true", oid)
				}
			}
		}
	})

	t.Run("OID с несуществующей колонкой не является колонкой", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1.2.1.2.2.1.99.1")

		if table.IsColumnOID(oid) {
			t.Error("Несуществующая колонка не должна быть колонкой")
		}
	})

	t.Run("Не изменяет входной OID", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		table.IsColumnOID(oid)

		if !oid.Equal(oidCopy) {
			t.Error("IsColumnOID не должен изменять входной OID")
		}
	})
}

// Тест с round trip
func TestTableOIDIsColumnOIDRoundTrip(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumns(map[string]uint32{
		"ifIndex": 1,
		"ifDescr": 2,
	})

	// Создаем OID колонок
	for name, columnNum := range table.Columns {
		oid, err := table.GetColumnOID(name, 1)
		if err != nil {
			t.Fatalf("GetColumnOID: %v", err)
		}

		// Проверяем IsColumnOID
		if !table.IsColumnOID(oid) {
			t.Errorf("IsColumnOID(%v) = false для колонки %s (%d)", oid, name, columnNum)
		}

		// Парсим обратно
		column, index, err := table.ParseRowOID(oid)
		if err != nil {
			t.Errorf("ParseRowOID: %v", err)
			continue
		}

		if column != columnNum {
			t.Errorf("column = %d, ожидалось %d", column, columnNum)
		}
		if index != 1 {
			t.Errorf("index = %d, ожидалось 1", index)
		}
	}
}

// Тест с подтестами
func TestTableOIDIsColumnOIDCategories(t *testing.T) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	t.Run("Существующие колонки", func(t *testing.T) {
		tests := []OID{
			MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			MustParseOID("1.3.6.1.2.1.2.2.1.2"),
			MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2.3"),
		}

		for _, oid := range tests {
			if !table.IsColumnOID(oid) {
				t.Errorf("IsColumnOID(%v) = false, ожидалось true", oid)
			}
		}
	})

	t.Run("Несуществующие колонки", func(t *testing.T) {
		tests := []OID{
			MustParseOID("1.3.6.1.2.1.2.2.1.1.1"),
			MustParseOID("1.3.6.1.2.1.2.2.1.3.1"),
			MustParseOID("1.3.6.1.2.1.2.2.1"),
			MustParseOID("1.3.6.1.2.1.1.1.0"),
		}

		for _, oid := range tests {
			if table.IsColumnOID(oid) {
				t.Errorf("IsColumnOID(%v) = true, ожидалось false", oid)
			}
		}
	})
}

// Пример использования
func ExampleTableOID_IsColumnOID() {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)

	// Существующая колонка
	oid1 := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")
	fmt.Println(table.IsColumnOID(oid1))

	// Несуществующая колонка
	oid2 := MustParseOID("1.3.6.1.2.1.2.2.1.99.1")
	fmt.Println(table.IsColumnOID(oid2))

	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkTableOIDIsColumnOID(b *testing.B) {
	table := NewTableOID(MustParseOID("1.3.6.1.2.1.2.2.1"))
	table.AddColumn("ifDescr", 2)
	oid := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	b.ReportAllocs()
	for b.Loop() {
		_ = table.IsColumnOID(oid)
	}
}
