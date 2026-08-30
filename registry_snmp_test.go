package oid

import (
	"fmt"
	"testing"
)

func TestSNMPRegistryStructure(t *testing.T) {
	reg := NewSNMPRegistry()

	t.Run("Поля инициализированы", func(t *testing.T) {
		if reg.scalars == nil {
			t.Error("scalars должен быть инициализирован")
		}

		if reg.columns == nil {
			t.Error("columns должен быть инициализирован")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		if len(reg.scalars) != 0 {
			t.Errorf("len(scalars) = %d, want 0", len(reg.scalars))
		}

		if len(reg.columns) != 0 {
			t.Errorf("len(columns) = %d, want 0", len(reg.columns))
		}
	})
}

func TestSNMPRegistryMapTypes(t *testing.T) {
	reg := NewSNMPRegistry()

	// Проверяем типы map
	var _ map[string]*ScalarOID = reg.scalars
	var _ map[string]*ColumnarOID = reg.columns
}

func TestSNMPRegistryIndependence(t *testing.T) {
	reg1 := NewSNMPRegistry()
	reg2 := NewSNMPRegistry()

	if reg1 == reg2 {
		t.Error("NewSNMPRegistry должен создавать разные экземпляры")
	}

	// Добавляем в первый
	oid := MustScalarOID("1.3.6.1.0")
	reg1.RegisterScalar("test", oid)

	// Второй не должен измениться
	if len(reg2.scalars) != 0 {
		t.Error("Экземпляры не должны влиять друг на друга")
	}
}

func TestSNMPRegistryPointerStorage(t *testing.T) {
	reg := NewSNMPRegistry()

	oid := MustScalarOID("1.3.6.1.0")
	reg.RegisterScalar("test", oid)

	// Проверяем, что хранится указатель
	got, exists := reg.scalars["test"]
	if !exists {
		t.Fatal("test not found")
	}

	if got == nil {
		t.Error("Должен храниться указатель")
	}

	if !(*got).Equal(oid) {
		t.Errorf("got %v, want %v", *got, oid)
	}
}

func TestSNMPRegistryColumnStorage(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1)
	reg.RegisterColumn("ifDescr", col)

	got, exists := reg.columns["ifDescr"]
	if !exists {
		t.Fatal("ifDescr not found")
	}

	if got == nil {
		t.Error("Должен храниться указатель")
	}

	if !got.Equal(col) {
		t.Errorf("got %v, want %v", *got, col)
	}
}

// Пример использования
func ExampleSNMPRegistry() {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("sysDescr", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	if oid, exists := reg.GetScalar("sysDescr"); exists {
		fmt.Println(oid)
	}
	// Output: 1.3.6.1.2.1.1.1.0
}

// Бенчмарк
func BenchmarkSNMPRegistryCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = NewSNMPRegistry()
	}
}

func TestNewSNMPRegistry(t *testing.T) {
	t.Run("Инициализация", func(t *testing.T) {
		reg := NewSNMPRegistry()

		if reg == nil {
			t.Fatal("NewSNMPRegistry: nil результат")
		}

		// Проверяем инициализацию map
		if reg.scalars == nil {
			t.Error("scalars должен быть инициализирован")
		}

		if reg.columns == nil {
			t.Error("columns должен быть инициализирован")
		}

		// Проверяем, что map пустые
		if len(reg.scalars) != 0 {
			t.Errorf("len(scalars) = %d, ожидалось 0", len(reg.scalars))
		}

		if len(reg.columns) != 0 {
			t.Errorf("len(columns) = %d, ожидалось 0", len(reg.columns))
		}
	})

	t.Run("Независимые экземпляры", func(t *testing.T) {
		reg1 := NewSNMPRegistry()
		reg2 := NewSNMPRegistry()

		if reg1 == reg2 {
			t.Error("NewSNMPRegistry должен создавать разные экземпляры")
		}

		// Добавляем в первый
		oid := MustScalarOID("1.3.6.1.0")
		reg1.RegisterScalar("test", oid)

		// Второй не должен измениться
		if len(reg2.scalars) != 0 {
			t.Error("Экземпляры не должны влиять друг на друга")
		}
	})

	t.Run("Типы map", func(t *testing.T) {
		reg := NewSNMPRegistry()

		// Проверяем, что map хранит указатели
		var _ map[string]*ScalarOID = reg.scalars
		var _ map[string]*ColumnarOID = reg.columns
	})
}

// Пример использования
func ExampleNewSNMPRegistry() {
	reg := NewSNMPRegistry()

	// Регистрируем
	reg.RegisterScalar("sysDescr", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	// Безопасное получение (копия)
	if oid, exists := reg.GetScalar("sysDescr"); exists {
		fmt.Println(oid)
	}

	// Быстрое получение (указатель)
	if oidPtr, exists := reg.GetScalarNoCopy("sysDescr"); exists && oidPtr != nil {
		fmt.Println(*oidPtr)
	}
	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1.0
}

// Бенчмарки
func BenchmarkNewSNMPRegistry(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = NewSNMPRegistry()
	}
}

func TestRegisterScalar(t *testing.T) {
	reg := NewSNMPRegistry()

	tests := []struct {
		name        string
		oidName     string
		oid         ScalarOID
		expectedLen int
	}{
		{
			name:        "Первая регистрация",
			oidName:     "sysDescr",
			oid:         MustScalarOID("1.3.6.1.2.1.1.1.0"),
			expectedLen: 1,
		},
		{
			name:        "Вторая регистрация",
			oidName:     "sysUpTime",
			oid:         MustScalarOID("1.3.6.1.2.1.1.3.0"),
			expectedLen: 2,
		},
		{
			name:        "Третья регистрация",
			oidName:     "sysName",
			oid:         MustScalarOID("1.3.6.1.2.1.1.5.0"),
			expectedLen: 3,
		},
		{
			name:        "Пустое имя",
			oidName:     "",
			oid:         MustScalarOID("1.3.6.1.0"),
			expectedLen: 4,
		},
		{
			name:        "Дубликат (перезапись)",
			oidName:     "sysDescr",
			oid:         MustScalarOID("1.3.6.1.2.1.1.5.0"),
			expectedLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg.RegisterScalar(tt.oidName, tt.oid)

			if len(reg.scalars) != tt.expectedLen {
				t.Errorf("len = %d, want %d", len(reg.scalars), tt.expectedLen)
			}

			got, exists := reg.scalars[tt.oidName]
			if !exists {
				t.Fatalf("OID '%s' not found", tt.oidName)
			}
			if got == nil {
				t.Fatal("pointer is nil")
			}
			if !(*got).Equal(tt.oid) {
				t.Errorf("got %v, want %v", *got, tt.oid)
			}
		})
	}
}

func TestRegisterScalarCopy(t *testing.T) {
	reg := NewSNMPRegistry()

	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	// Изменяем оригинал
	oid[0] = 99

	// Реестр не должен измениться
	if (*reg.scalars["sysDescr"])[0] != 1 {
		t.Error("RegisterScalar должен создать копию")
	}
}

func TestRegisterScalarOverwrite(t *testing.T) {
	reg := NewSNMPRegistry()

	oid1 := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid2 := MustScalarOID("1.3.6.1.2.1.1.5.0")

	reg.RegisterScalar("test", oid1)
	reg.RegisterScalar("test", oid2)

	if len(reg.scalars) != 1 {
		t.Errorf("len = %d, want 1", len(reg.scalars))
	}

	if !(*reg.scalars["test"]).Equal(oid2) {
		t.Error("Должен сохраниться второй OID")
	}
}

func TestRegisterScalarNil(t *testing.T) {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("nil", nil)

	got, exists := reg.scalars["nil"]
	if !exists {
		t.Fatal("nil not found")
	}
	if got == nil {
		t.Fatal("pointer is nil")
	}
	if len(*got) != 0 {
		t.Errorf("len = %d, want 0", len(*got))
	}
}

func TestRegisterScalarEmpty(t *testing.T) {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("empty", ScalarOID{})

	got, exists := reg.scalars["empty"]
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("pointer is nil")
	}
	if len(*got) != 0 {
		t.Errorf("len = %d, want 0", len(*got))
	}
}

func TestRegisterScalarChangeThroughPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("test", oid)

	// Получаем указатель
	got := reg.scalars["test"]

	// Изменяем через указатель
	(*got)[0] = 99

	// Реестр изменился
	if (*reg.scalars["test"])[0] != 99 {
		t.Error("Изменение через указатель должно влиять")
	}

	// Восстанавливаем
	(*got)[0] = 1
}

// Пример использования
func ExampleSNMPRegistry_RegisterScalar() {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("sysDescr", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	if oid, exists := reg.GetScalar("sysDescr"); exists {
		fmt.Println(oid)
	}
	// Output: 1.3.6.1.2.1.1.1.0
}

// Бенчмарк
func BenchmarkRegisterScalar(b *testing.B) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg.RegisterScalar("test", oid)
	}
}

func TestRegisterScalarNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()

	tests := []struct {
		name        string
		oidName     string
		oid         ScalarOID
		expectedLen int
	}{
		{
			name:        "Первая регистрация",
			oidName:     "sysDescr",
			oid:         MustScalarOID("1.3.6.1.2.1.1.1.0"),
			expectedLen: 1,
		},
		{
			name:        "Вторая регистрация",
			oidName:     "sysUpTime",
			oid:         MustScalarOID("1.3.6.1.2.1.1.3.0"),
			expectedLen: 2,
		},
		{
			name:        "Дубликат (перезапись)",
			oidName:     "sysDescr",
			oid:         MustScalarOID("1.3.6.1.2.1.1.5.0"),
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg.RegisterScalarNoCopy(tt.oidName, tt.oid)

			if len(reg.scalars) != tt.expectedLen {
				t.Errorf("len = %d, want %d", len(reg.scalars), tt.expectedLen)
			}

			got, exists := reg.scalars[tt.oidName]
			if !exists {
				t.Fatalf("OID '%s' not found", tt.oidName)
			}
			if got == nil {
				t.Fatal("pointer is nil")
			}
			if !(*got).Equal(tt.oid) {
				t.Errorf("got %v, want %v", *got, tt.oid)
			}
		})
	}
}

func TestRegisterScalarNoCopyStoresReference(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	reg.RegisterScalarNoCopy("sysDescr", oid)

	// Изменяем оригинал
	oid[0] = 99

	// Реестр должен измениться (хранит ссылку)
	if (*reg.scalars["sysDescr"])[0] != 99 {
		t.Error("RegisterScalarNoCopy должен хранить ссылку")
	}

	// Восстанавливаем
	oid[0] = 1
}

func TestRegisterScalarNoCopyChangeThroughPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalarNoCopy("sysDescr", oid)

	// Получаем указатель
	got := reg.scalars["sysDescr"]

	// Изменяем через указатель
	(*got)[0] = 99

	// Оригинал тоже изменился (общий массив)
	if oid[0] != 99 {
		t.Error("Изменение через указатель должно влиять на оригинал")
	}

	// Восстанавливаем
	(*got)[0] = 1
}

func TestRegisterScalarNoCopyNil(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterScalarNoCopy("nil", nil)

	got, exists := reg.scalars["nil"]
	if !exists {
		t.Fatal("nil not found")
	}
	if got == nil {
		t.Fatal("pointer is nil")
	}
	if len(*got) != 0 {
		t.Errorf("len = %d, want 0", len(*got))
	}
}

func TestRegisterScalarNoCopyOverwrite(t *testing.T) {
	reg := NewSNMPRegistry()

	oid1 := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid2 := MustScalarOID("1.3.6.1.2.1.1.5.0")

	reg.RegisterScalarNoCopy("test", oid1)
	reg.RegisterScalarNoCopy("test", oid2)

	if len(reg.scalars) != 1 {
		t.Errorf("len = %d, want 1", len(reg.scalars))
	}

	if !(*reg.scalars["test"]).Equal(oid2) {
		t.Error("Должен сохраниться второй OID")
	}
}

func TestRegisterScalarVsNoCopy(t *testing.T) {
	reg1 := NewSNMPRegistry()
	reg2 := NewSNMPRegistry()

	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	// RegisterScalar - копия
	reg1.RegisterScalar("test", oid)
	oid[0] = 99
	if (*reg1.scalars["test"])[0] != 1 {
		t.Error("RegisterScalar должен создать копию")
	}

	// Восстанавливаем
	oid[0] = 1

	// RegisterScalarNoCopy - ссылка
	reg2.RegisterScalarNoCopy("test", oid)
	oid[0] = 99
	if (*reg2.scalars["test"])[0] != 99 {
		t.Error("RegisterScalarNoCopy должен хранить ссылку")
	}

	// Восстанавливаем
	oid[0] = 1
}

// Пример использования
func ExampleSNMPRegistry_RegisterScalarNoCopy() {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	// Регистрируем без копирования
	reg.RegisterScalarNoCopy("sysDescr", oid)

	// Изменяем оригинал
	oid[0] = 99

	// Реестр видит изменение
	if ptr, exists := reg.GetScalarNoCopy("sysDescr"); exists {
		fmt.Println((*ptr)[0])
	}
	// Output: 99
}

// Бенчмарк
func BenchmarkRegisterScalarNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg.RegisterScalarNoCopy("test", oid)
	}
}

// Сравнение RegisterScalar vs RegisterScalarNoCopy
func BenchmarkRegisterScalarVsNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.Run("RegisterScalar", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			reg.RegisterScalar("test", oid)
		}
	})

	b.Run("RegisterScalarNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			reg.RegisterScalarNoCopy("test", oid)
		}
	})
}

func TestRegisterColumn(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name        string
		colName     string
		col         ColumnarOID
		expectedLen int
	}{
		{
			name:        "Первая колонка",
			colName:     "ifIndex",
			col:         NewColumnarOID(base, 1),
			expectedLen: 1,
		},
		{
			name:        "Вторая колонка",
			colName:     "ifDescr",
			col:         NewColumnarOID(base, 2),
			expectedLen: 2,
		},
		{
			name:        "Третья колонка с индексами",
			colName:     "ifType",
			col:         NewColumnarOID(base, 3, 1, 2),
			expectedLen: 3,
		},
		{
			name:        "Дубликат (перезапись)",
			colName:     "ifIndex",
			col:         NewColumnarOID(base, 10),
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg.RegisterColumn(tt.colName, tt.col)

			if len(reg.columns) != tt.expectedLen {
				t.Errorf("len = %d, want %d", len(reg.columns), tt.expectedLen)
			}

			got, exists := reg.columns[tt.colName]
			if !exists {
				t.Fatalf("Column '%s' not found", tt.colName)
			}
			if got == nil {
				t.Fatal("pointer is nil")
			}
			if !got.Equal(tt.col) {
				t.Errorf("got %v, want %v", got, tt.col)
			}
		})
	}
}

func TestRegisterColumnDeepCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	// Изменяем оригинал
	col.Column = 99
	col.Base[0] = 99
	col.Indexes[0] = 99

	// Реестр не должен измениться
	if reg.columns["ifDescr"].Column != 2 {
		t.Error("Column should be copied")
	}
	if reg.columns["ifDescr"].Base[0] != 1 {
		t.Error("Base should be copied")
	}
	if reg.columns["ifDescr"].Indexes[0] != 1 {
		t.Error("Indexes should be copied")
	}
}

func TestRegisterColumnOverwrite(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col1 := NewColumnarOID(base, 2)
	col2 := NewColumnarOID(base, 3)

	reg.RegisterColumn("test", col1)
	reg.RegisterColumn("test", col2)

	if len(reg.columns) != 1 {
		t.Errorf("len = %d, want 1", len(reg.columns))
	}

	if reg.columns["test"].Column != 3 {
		t.Error("Должен сохраниться второй Column")
	}
}

func TestRegisterColumnEmpty(t *testing.T) {
	reg := NewSNMPRegistry()

	reg.RegisterColumn("empty", ColumnarOID{})

	got, exists := reg.columns["empty"]
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("pointer is nil")
	}
	if got.Column != 0 {
		t.Errorf("Column = %d, want 0", got.Column)
	}
	if len(got.Base) != 0 {
		t.Errorf("len(Base) = %d, want 0", len(got.Base))
	}
	if len(got.Indexes) != 0 {
		t.Errorf("len(Indexes) = %d, want 0", len(got.Indexes))
	}
}

func TestRegisterColumnChangeThroughPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("test", col)

	// Получаем указатель
	got := reg.columns["test"]

	// Изменяем через указатель
	got.Column = 99
	got.Base[0] = 99
	got.Indexes[0] = 99

	// Реестр изменился
	if reg.columns["test"].Column != 99 {
		t.Error("Column should change")
	}
	if reg.columns["test"].Base[0] != 99 {
		t.Error("Base should change")
	}
	if reg.columns["test"].Indexes[0] != 99 {
		t.Error("Indexes should change")
	}
}

// Пример использования
func ExampleSNMPRegistry_RegisterColumn() {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	if col, exists := reg.GetColumn("ifDescr"); exists {
		fmt.Println(col.String())
	}
	// Output: 1.3.6.1.2.1.2.2.1.2
}

// Бенчмарк
func BenchmarkRegisterColumn(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg.RegisterColumn("test", col)
	}
}

func TestRegisterColumnNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col1 := NewColumnarOID(base, 1)
	col2 := NewColumnarOID(base, 2, 1, 2)

	tests := []struct {
		name        string
		colName     string
		col         *ColumnarOID
		expectedLen int
	}{
		{
			name:        "Первая колонка",
			colName:     "ifIndex",
			col:         &col1,
			expectedLen: 1,
		},
		{
			name:        "Вторая колонка",
			colName:     "ifDescr",
			col:         &col2,
			expectedLen: 2,
		},
		{
			name:        "Дубликат (перезапись)",
			colName:     "ifIndex",
			col:         &col2,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg.RegisterColumnNoCopy(tt.colName, tt.col)

			if len(reg.columns) != tt.expectedLen {
				t.Errorf("len = %d, want %d", len(reg.columns), tt.expectedLen)
			}

			got, exists := reg.columns[tt.colName]
			if !exists {
				t.Fatalf("Column '%s' not found", tt.colName)
			}
			if got == nil {
				t.Fatal("pointer is nil")
			}
			if got != tt.col {
				t.Error("should store the same pointer")
			}
		})
	}
}

func TestRegisterColumnNoCopyStoresReference(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)

	reg.RegisterColumnNoCopy("ifDescr", &col)

	// Изменяем оригинал
	col.Column = 99
	col.Base[0] = 99
	col.Indexes[0] = 99

	// Реестр должен измениться (хранит ссылку)
	if reg.columns["ifDescr"].Column != 99 {
		t.Error("Column should change")
	}
	if reg.columns["ifDescr"].Base[0] != 99 {
		t.Error("Base should change")
	}
	if reg.columns["ifDescr"].Indexes[0] != 99 {
		t.Error("Indexes should change")
	}

	// Восстанавливаем
	col.Column = 2
	col.Base[0] = 1
	col.Indexes[0] = 1
}

func TestRegisterColumnNoCopyChangeThroughPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)

	reg.RegisterColumnNoCopy("ifDescr", &col)

	// Получаем указатель из реестра
	got := reg.columns["ifDescr"]

	// Изменяем через указатель
	got.Column = 99
	got.Base[0] = 99
	got.Indexes[0] = 99

	// Оригинал тоже изменился (тот же объект)
	if col.Column != 99 {
		t.Error("Column should change in original")
	}
	if col.Base[0] != 99 {
		t.Error("Base should change in original")
	}
	if col.Indexes[0] != 99 {
		t.Error("Indexes should change in original")
	}

	// Восстанавливаем
	got.Column = 2
	got.Base[0] = 1
	got.Indexes[0] = 1
}

func TestRegisterColumnNoCopyNil(t *testing.T) {
	reg := NewSNMPRegistry()

	reg.RegisterColumnNoCopy("nil", nil)

	got, exists := reg.columns["nil"]
	if !exists {
		t.Fatal("nil not found")
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestRegisterColumnNoCopyEmpty(t *testing.T) {
	reg := NewSNMPRegistry()
	empty := ColumnarOID{}

	reg.RegisterColumnNoCopy("empty", &empty)

	got, exists := reg.columns["empty"]
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("pointer is nil")
	}
	if got.Column != 0 {
		t.Errorf("Column = %d, want 0", got.Column)
	}
}

func TestRegisterColumnNoCopyOverwrite(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col1 := NewColumnarOID(base, 2)
	col2 := NewColumnarOID(base, 3)

	reg.RegisterColumnNoCopy("test", &col1)
	reg.RegisterColumnNoCopy("test", &col2)

	if len(reg.columns) != 1 {
		t.Errorf("len = %d, want 1", len(reg.columns))
	}

	if reg.columns["test"] != &col2 {
		t.Error("Должен сохраниться второй указатель")
	}
}

func TestRegisterColumnVsNoCopy(t *testing.T) {
	reg1 := NewSNMPRegistry()
	reg2 := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)

	// RegisterColumn - копия
	reg1.RegisterColumn("test", col)
	col.Column = 99
	if reg1.columns["test"].Column != 2 {
		t.Error("RegisterColumn должен создать копию")
	}

	// Восстанавливаем
	col.Column = 2

	// RegisterColumnNoCopy - ссылка
	reg2.RegisterColumnNoCopy("test", &col)
	col.Column = 99
	if reg2.columns["test"].Column != 99 {
		t.Error("RegisterColumnNoCopy должен хранить ссылку")
	}

	// Восстанавливаем
	col.Column = 2
}

// Пример использования
func ExampleSNMPRegistry_RegisterColumnNoCopy() {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	// Регистрируем указатель
	reg.RegisterColumnNoCopy("ifDescr", &col)

	// Изменяем оригинал
	col.Column = 99

	// Реестр видит изменение
	if ptr, exists := reg.GetColumnNoCopy("ifDescr"); exists {
		fmt.Println(ptr.Column)
	}
	// Output: 99
}

// Бенчмарк
func BenchmarkRegisterColumnNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg.RegisterColumnNoCopy("test", &col)
	}
}

// Сравнение RegisterColumn vs RegisterColumnNoCopy
func BenchmarkRegisterColumnVsNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	b.Run("RegisterColumn", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			reg.RegisterColumn("test", col)
		}
	})

	b.Run("RegisterColumnNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			reg.RegisterColumnNoCopy("test", &col)
		}
	})
}

func TestGetScalar(t *testing.T) {
	reg := NewSNMPRegistry()

	// Регистрируем OID
	sysDescr := MustScalarOID("1.3.6.1.2.1.1.1.0")
	sysUpTime := MustScalarOID("1.3.6.1.2.1.1.3.0")

	reg.RegisterScalar("sysDescr", sysDescr)
	reg.RegisterScalar("sysUpTime", sysUpTime)

	tests := []struct {
		name     string
		lookup   string
		expected ScalarOID
		exists   bool
	}{
		{
			name:     "Существующий sysDescr",
			lookup:   "sysDescr",
			expected: sysDescr,
			exists:   true,
		},
		{
			name:     "Существующий sysUpTime",
			lookup:   "sysUpTime",
			expected: sysUpTime,
			exists:   true,
		},
		{
			name:     "Несуществующий",
			lookup:   "nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.GetScalar(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if !got.Equal(tt.expected) {
					t.Errorf("got %v, want %v", got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestGetScalarDeepCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	// Получаем копию
	got, exists := reg.GetScalar("sysDescr")
	if !exists {
		t.Fatal("sysDescr not found")
	}

	// Изменяем копию
	got[0] = 99

	// Реестр не должен измениться
	if (*reg.scalars["sysDescr"])[0] != 1 {
		t.Error("GetScalar должен вернуть глубокую копию")
	}
}

func TestGetScalarNil(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("nil", nil)

	got, exists := reg.GetScalar("nil")
	if !exists {
		t.Fatal("nil not found")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestGetScalarEmpty(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("empty", ScalarOID{})

	got, exists := reg.GetScalar("empty")
	if !exists {
		t.Fatal("empty not found")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestGetScalarEachCallReturnsNewCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	got1, _ := reg.GetScalar("sysDescr")
	got2, _ := reg.GetScalar("sysDescr")

	// Изменяем первую копию
	got1[0] = 99

	// Вторая не должна измениться
	if got2[0] != 1 {
		t.Error("Каждый вызов должен возвращать новую копию")
	}
}

func TestGetScalarNotModifyRegistry(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	before := len(reg.scalars)
	reg.GetScalar("sysDescr")
	after := len(reg.scalars)

	if before != after {
		t.Error("GetScalar не должен изменять реестр")
	}
}

// Пример использования
func ExampleSNMPRegistry_GetScalar() {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("sysDescr", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	if oid, exists := reg.GetScalar("sysDescr"); exists {
		fmt.Println(oid)
	}

	_, exists := reg.GetScalar("nonexistent")
	fmt.Println(exists)
	// Output:
	// 1.3.6.1.2.1.1.1.0
	// false
}

// Бенчмарк
func BenchmarkGetScalar(b *testing.B) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("test", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.GetScalar("test")
	}
}

func TestGetScalarNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()

	// Регистрируем OID
	sysDescr := MustScalarOID("1.3.6.1.2.1.1.1.0")
	sysUpTime := MustScalarOID("1.3.6.1.2.1.1.3.0")

	reg.RegisterScalar("sysDescr", sysDescr)
	reg.RegisterScalar("sysUpTime", sysUpTime)

	tests := []struct {
		name     string
		lookup   string
		expected ScalarOID
		exists   bool
	}{
		{
			name:     "Существующий sysDescr",
			lookup:   "sysDescr",
			expected: sysDescr,
			exists:   true,
		},
		{
			name:     "Существующий sysUpTime",
			lookup:   "sysUpTime",
			expected: sysUpTime,
			exists:   true,
		},
		{
			name:     "Несуществующий",
			lookup:   "nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.GetScalarNoCopy(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if got == nil {
					t.Fatal("got nil pointer")
				}
				if !(*got).Equal(tt.expected) {
					t.Errorf("got %v, want %v", *got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestGetScalarNoCopyReturnsPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	got, exists := reg.GetScalarNoCopy("sysDescr")
	if !exists {
		t.Fatal("sysDescr not found")
	}
	if got == nil {
		t.Fatal("got nil pointer")
	}

	// Изменяем через указатель
	(*got)[0] = 99

	// Реестр должен измениться
	if (*reg.scalars["sysDescr"])[0] != 99 {
		t.Error("Изменение через указатель должно влиять на реестр")
	}

	// Восстанавливаем
	(*got)[0] = 1
}

func TestGetScalarNoCopySamePointer(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	got1, _ := reg.GetScalarNoCopy("sysDescr")
	got2, _ := reg.GetScalarNoCopy("sysDescr")

	// Оба должны указывать на один и тот же объект
	if got1 != got2 {
		t.Error("GetScalarNoCopy должен возвращать один и тот же указатель")
	}
}

func TestGetScalarNoCopyNil(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("nil", nil)

	got, exists := reg.GetScalarNoCopy("nil")
	if !exists {
		t.Fatal("nil not found")
	}
	if got == nil {
		t.Fatal("got nil pointer")
	}
	if len(*got) != 0 {
		t.Errorf("len = %d, want 0", len(*got))
	}
}

func TestGetScalarNoCopyEmpty(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("empty", ScalarOID{})

	got, exists := reg.GetScalarNoCopy("empty")
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("got nil pointer")
	}
	if len(*got) != 0 {
		t.Errorf("len = %d, want 0", len(*got))
	}
}

func TestGetScalarNoCopyNotModifyRegistry(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	before := len(reg.scalars)
	reg.GetScalarNoCopy("sysDescr")
	after := len(reg.scalars)

	if before != after {
		t.Error("GetScalarNoCopy не должен изменять реестр")
	}
}

func TestGetScalarVsGetScalarNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	oid := MustScalarOID("1.3.6.1.2.1.1.1.0")
	reg.RegisterScalar("sysDescr", oid)

	// GetScalar - копия
	copied, _ := reg.GetScalar("sysDescr")
	copied[0] = 99
	if (*reg.scalars["sysDescr"])[0] != 1 {
		t.Error("GetScalar должен вернуть копию")
	}

	// GetScalarNoCopy - указатель
	referenced, _ := reg.GetScalarNoCopy("sysDescr")
	(*referenced)[0] = 99
	if (*reg.scalars["sysDescr"])[0] != 99 {
		t.Error("GetScalarNoCopy должен вернуть указатель")
	}

	// Восстанавливаем
	(*referenced)[0] = 1
}

// Пример использования
func ExampleSNMPRegistry_GetScalarNoCopy() {
	reg := NewSNMPRegistry()

	reg.RegisterScalar("sysDescr", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	if oidPtr, exists := reg.GetScalarNoCopy("sysDescr"); exists && oidPtr != nil {
		fmt.Println(*oidPtr)
	}
	// Output: 1.3.6.1.2.1.1.1.0
}

// Бенчмарк
func BenchmarkGetScalarNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("test", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.GetScalarNoCopy("test")
	}
}

// Сравнение GetScalar vs GetScalarNoCopy
func BenchmarkGetScalarVsNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	reg.RegisterScalar("test", MustScalarOID("1.3.6.1.2.1.1.1.0"))

	b.Run("GetScalar", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.GetScalar("test")
		}
	})

	b.Run("GetScalarNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.GetScalarNoCopy("test")
		}
	})
}

func TestGetColumn(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Регистрируем колонки
	ifIndex := NewColumnarOID(base, 1)
	ifDescr := NewColumnarOID(base, 2, 1, 2)

	reg.RegisterColumn("ifIndex", ifIndex)
	reg.RegisterColumn("ifDescr", ifDescr)

	tests := []struct {
		name     string
		lookup   string
		expected ColumnarOID
		exists   bool
	}{
		{
			name:     "Существующая ifIndex",
			lookup:   "ifIndex",
			expected: ifIndex,
			exists:   true,
		},
		{
			name:     "Существующая ifDescr с индексами",
			lookup:   "ifDescr",
			expected: ifDescr,
			exists:   true,
		},
		{
			name:     "Несуществующая",
			lookup:   "nonexistent",
			expected: ColumnarOID{},
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: ColumnarOID{},
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.GetColumn(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if !got.Equal(tt.expected) {
					t.Errorf("got %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestGetColumnDeepCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	// Получаем копию
	got, exists := reg.GetColumn("ifDescr")
	if !exists {
		t.Fatal("ifDescr not found")
	}

	// Изменяем копию
	got.Column = 99
	got.Base[0] = 99
	got.Indexes[0] = 99

	// Реестр не должен измениться
	if reg.columns["ifDescr"].Column != 2 {
		t.Error("Column should be copied")
	}
	if reg.columns["ifDescr"].Base[0] != 1 {
		t.Error("Base should be copied")
	}
	if reg.columns["ifDescr"].Indexes[0] != 1 {
		t.Error("Indexes should be copied")
	}
}

func TestGetColumnEmpty(t *testing.T) {
	reg := NewSNMPRegistry()

	reg.RegisterColumn("empty", ColumnarOID{})

	got, exists := reg.GetColumn("empty")
	if !exists {
		t.Fatal("empty not found")
	}

	if got.Column != 0 {
		t.Errorf("Column = %d, want 0", got.Column)
	}
	if len(got.Base) != 0 {
		t.Errorf("len(Base) = %d, want 0", len(got.Base))
	}
	if len(got.Indexes) != 0 {
		t.Errorf("len(Indexes) = %d, want 0", len(got.Indexes))
	}
}

func TestGetColumnEachCallReturnsNewCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	got1, _ := reg.GetColumn("ifDescr")
	got2, _ := reg.GetColumn("ifDescr")

	// Изменяем первую копию
	got1.Column = 99
	got1.Base[0] = 99
	got1.Indexes[0] = 99

	// Вторая не должна измениться
	if got2.Column != 2 {
		t.Error("Column should be independent")
	}
	if got2.Base[0] != 1 {
		t.Error("Base should be independent")
	}
	if got2.Indexes[0] != 1 {
		t.Error("Indexes should be independent")
	}
}

func TestGetColumnNotModifyRegistry(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)
	reg.RegisterColumn("ifDescr", col)

	before := len(reg.columns)
	reg.GetColumn("ifDescr")
	after := len(reg.columns)

	if before != after {
		t.Error("GetColumn не должен изменять реестр")
	}
}

// Пример использования
func ExampleSNMPRegistry_GetColumn() {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	if col, exists := reg.GetColumn("ifDescr"); exists {
		fmt.Println(col.String())
	}

	_, exists := reg.GetColumn("nonexistent")
	fmt.Println(exists)
	// Output:
	// 1.3.6.1.2.1.2.2.1.2
	// false
}

// Бенчмарк
func BenchmarkGetColumn(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	reg.RegisterColumn("test", NewColumnarOID(base, 2))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.GetColumn("test")
	}
}

func TestGetColumnNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Регистрируем колонки
	ifIndex := NewColumnarOID(base, 1)
	ifDescr := NewColumnarOID(base, 2, 1, 2)

	reg.RegisterColumn("ifIndex", ifIndex)
	reg.RegisterColumn("ifDescr", ifDescr)

	tests := []struct {
		name     string
		lookup   string
		expected ColumnarOID
		exists   bool
	}{
		{
			name:     "Существующая ifIndex",
			lookup:   "ifIndex",
			expected: ifIndex,
			exists:   true,
		},
		{
			name:     "Существующая ifDescr",
			lookup:   "ifDescr",
			expected: ifDescr,
			exists:   true,
		},
		{
			name:     "Несуществующая",
			lookup:   "nonexistent",
			expected: ColumnarOID{},
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: ColumnarOID{},
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.GetColumnNoCopy(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if got == nil {
					t.Fatal("got nil pointer")
				}
				if !got.Equal(tt.expected) {
					t.Errorf("got %v, want %v", *got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestGetColumnNoCopyReturnsPointer(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	got, exists := reg.GetColumnNoCopy("ifDescr")
	if !exists {
		t.Fatal("ifDescr not found")
	}
	if got == nil {
		t.Fatal("got nil pointer")
	}

	// Изменяем через указатель
	got.Column = 99
	got.Base[0] = 99
	got.Indexes[0] = 99

	// Реестр должен измениться
	if reg.columns["ifDescr"].Column != 99 {
		t.Error("Column should change")
	}
	if reg.columns["ifDescr"].Base[0] != 99 {
		t.Error("Base should change")
	}
	if reg.columns["ifDescr"].Indexes[0] != 99 {
		t.Error("Indexes should change")
	}
}

func TestGetColumnNoCopySamePointer(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)
	reg.RegisterColumn("ifDescr", col)

	got1, _ := reg.GetColumnNoCopy("ifDescr")
	got2, _ := reg.GetColumnNoCopy("ifDescr")

	// Оба должны указывать на один и тот же объект
	if got1 != got2 {
		t.Error("GetColumnNoCopy должен возвращать один и тот же указатель")
	}
}

func TestGetColumnNoCopyEmpty(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterColumn("empty", ColumnarOID{})

	got, exists := reg.GetColumnNoCopy("empty")
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("got nil pointer")
	}
	if got.Column != 0 {
		t.Errorf("Column = %d, want 0", got.Column)
	}
	if len(got.Base) != 0 {
		t.Errorf("len(Base) = %d, want 0", len(got.Base))
	}
	if len(got.Indexes) != 0 {
		t.Errorf("len(Indexes) = %d, want 0", len(got.Indexes))
	}
}

func TestGetColumnNoCopyNotModifyRegistry(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)
	reg.RegisterColumn("ifDescr", col)

	before := len(reg.columns)
	reg.GetColumnNoCopy("ifDescr")
	after := len(reg.columns)

	if before != after {
		t.Error("GetColumnNoCopy не должен изменять реестр")
	}
}

func TestGetColumnVsGetColumnNoCopy(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	// GetColumn - копия
	copied, _ := reg.GetColumn("ifDescr")
	copied.Column = 99
	if reg.columns["ifDescr"].Column != 2 {
		t.Error("GetColumn должен вернуть копию")
	}

	// GetColumnNoCopy - указатель
	referenced, _ := reg.GetColumnNoCopy("ifDescr")
	referenced.Column = 99
	if reg.columns["ifDescr"].Column != 99 {
		t.Error("GetColumnNoCopy должен вернуть указатель")
	}

	// Восстанавливаем
	referenced.Column = 2
}

// Пример использования
func ExampleSNMPRegistry_GetColumnNoCopy() {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	if colPtr, exists := reg.GetColumnNoCopy("ifDescr"); exists && colPtr != nil {
		fmt.Println(colPtr.String())
	}
	// Output: 1.3.6.1.2.1.2.2.1.2
}

// Бенчмарк
func BenchmarkGetColumnNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	reg.RegisterColumn("test", NewColumnarOID(base, 2))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.GetColumnNoCopy("test")
	}
}

// Сравнение GetColumn vs GetColumnNoCopy
func BenchmarkGetColumnVsNoCopy(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	reg.RegisterColumn("test", NewColumnarOID(base, 2))

	b.Run("GetColumn", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.GetColumn("test")
		}
	})

	b.Run("GetColumnNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.GetColumnNoCopy("test")
		}
	})
}

func TestGetColumnWithIndexes(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Регистрируем колонки
	reg.RegisterColumn("ifIndex", NewColumnarOID(base, 1))
	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	tests := []struct {
		name     string
		colName  string
		indexes  []uint32
		expected string
		exists   bool
	}{
		{
			name:     "Один индекс",
			colName:  "ifDescr",
			indexes:  []uint32{1},
			expected: "1.3.6.1.2.1.2.2.1.2.1",
			exists:   true,
		},
		{
			name:     "Два индекса",
			colName:  "ifDescr",
			indexes:  []uint32{1, 2},
			expected: "1.3.6.1.2.1.2.2.1.2.1.2",
			exists:   true,
		},
		{
			name:     "Три индекса",
			colName:  "ifDescr",
			indexes:  []uint32{1, 2, 3},
			expected: "1.3.6.1.2.1.2.2.1.2.1.2.3",
			exists:   true,
		},
		{
			name:     "Без индексов",
			colName:  "ifDescr",
			indexes:  []uint32{},
			expected: "1.3.6.1.2.1.2.2.1.2",
			exists:   true,
		},
		{
			name:     "Nil индексы",
			colName:  "ifDescr",
			indexes:  nil,
			expected: "1.3.6.1.2.1.2.2.1.2",
			exists:   true,
		},
		{
			name:     "Индекс 0",
			colName:  "ifIndex",
			indexes:  []uint32{0},
			expected: "1.3.6.1.2.1.2.2.1.1.0",
			exists:   true,
		},
		{
			name:     "Несуществующая колонка",
			colName:  "nonexistent",
			indexes:  []uint32{1},
			expected: "",
			exists:   false,
		},
		{
			name:     "Пустое имя",
			colName:  "",
			indexes:  []uint32{1},
			expected: "",
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.GetColumnWithIndexes(tt.colName, tt.indexes...)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if got.String() != tt.expected {
					t.Errorf("got %s, want %s", got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestGetColumnWithIndexesReturnsNewOID(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	oid1, _ := reg.GetColumnWithIndexes("ifDescr", 1)
	oid2, _ := reg.GetColumnWithIndexes("ifDescr", 1)

	// Изменяем первый
	oid1[0] = 99

	// Второй не должен измениться
	if oid2[0] != 1 {
		t.Error("Каждый вызов должен возвращать новый OID")
	}
}

func TestGetColumnWithIndexesNotModifyRegistry(t *testing.T) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)
	reg.RegisterColumn("ifDescr", col)

	before := len(reg.columns)
	reg.GetColumnWithIndexes("ifDescr", 1)
	after := len(reg.columns)

	if before != after {
		t.Error("GetColumnWithIndexes не должен изменять реестр")
	}

	// Проверяем, что оригинальная колонка не изменилась
	if reg.columns["ifDescr"].Column != 2 {
		t.Error("Column should not change")
	}
	if reg.columns["ifDescr"].Base[0] != 1 {
		t.Error("Base should not change")
	}
	if len(reg.columns["ifDescr"].Indexes) != 2 {
		t.Error("Indexes should not change")
	}
}

func TestGetColumnWithIndexesEmptyColumn(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterColumn("empty", ColumnarOID{})

	got, exists := reg.GetColumnWithIndexes("empty", 1)
	if !exists {
		t.Fatal("empty not found")
	}
	if got == nil {
		t.Fatal("got nil")
	}

	// Для пустой колонки результат: Column(0) + Indexes(1)
	expected := "0.1"
	if got.String() != expected {
		t.Errorf("got %s, want %s", got, expected)
	}
}

func TestGetColumnWithIndexesNilColumn(t *testing.T) {
	reg := NewSNMPRegistry()
	reg.RegisterColumn("nil", ColumnarOID{Base: nil, Indexes: nil})

	got, exists := reg.GetColumnWithIndexes("nil", 1)
	if !exists {
		t.Fatal("nil not found")
	}
	if got == nil {
		t.Fatal("got nil")
	}
}

// Пример использования
func ExampleSNMPRegistry_GetColumnWithIndexes() {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	reg.RegisterColumn("ifDescr", NewColumnarOID(base, 2))

	// Один индекс
	oid1, _ := reg.GetColumnWithIndexes("ifDescr", 1)
	fmt.Println(oid1)

	// Два индекса
	oid2, _ := reg.GetColumnWithIndexes("ifDescr", 1, 2)
	fmt.Println(oid2)

	// Output:
	// 1.3.6.1.2.1.2.2.1.2.1
	// 1.3.6.1.2.1.2.2.1.2.1.2
}

// Бенчмарк
func BenchmarkGetColumnWithIndexes(b *testing.B) {
	reg := NewSNMPRegistry()
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	reg.RegisterColumn("test", NewColumnarOID(base, 2))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.GetColumnWithIndexes("test", 1)
	}
}
