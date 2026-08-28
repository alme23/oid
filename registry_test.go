// oid/registry_test.go
package oid

import (
	"fmt"
	"sync"
	"testing"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry() вернул nil")
	}

	if reg.names == nil {
		t.Error("NewRegistry(): поле names не инициализировано")
	}

	if reg.oids == nil {
		t.Error("NewRegistry(): поле oids не инициализировано")
	}
}

func TestRegistryRegister(t *testing.T) {
	tests := []struct {
		name    string
		oidName string
		oid     OID
		wantErr bool
	}{
		{
			name:    "Корректная регистрация",
			oidName: "test",
			oid:     OID{1, 3, 6, 1, 4, 1},
			wantErr: false,
		},
		{
			name:    "Некорректный OID",
			oidName: "invalid",
			oid:     OID{3, 1},
			wantErr: true,
		},
		{
			name:    "Пустое имя",
			oidName: "",
			oid:     OID{1, 3, 6},
			wantErr: false, // Пустые имена допустимы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			err := reg.Register(tt.oidName, tt.oid)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Register(%q, %v): ожидалась ошибка", tt.oidName, tt.oid)
				}
			} else {
				if err != nil {
					t.Errorf("Register(%q, %v): неожиданная ошибка: %v", tt.oidName, tt.oid, err)
				}

				// Проверяем, что OID зарегистрирован
				registeredOID, exists := reg.LookupByName(tt.oidName)
				if !exists {
					t.Errorf("Register(%q, %v): OID не найден после регистрации", tt.oidName, tt.oid)
				}
				if !registeredOID.Equal(tt.oid) {
					t.Errorf("Register(%q, %v): зарегистрирован %v, ожидалось %v",
						tt.oidName, tt.oid, registeredOID, tt.oid)
				}
			}
		})
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	reg := NewRegistry()
	oid := OID{1, 3, 6, 1, 4, 1}

	// Первая регистрация
	err := reg.Register("first", oid)
	if err != nil {
		t.Fatalf("Первая регистрация не должна дать ошибку: %v", err)
	}

	// Регистрация того же OID под другим именем
	err = reg.Register("second", oid)
	if err == nil {
		t.Error("Регистрация того же OID под другим именем должна дать ошибку")
	}

	// Регистрация того же имени с другим OID
	err = reg.Register("first", OID{2, 100, 3})
	if err != nil {
		t.Errorf("Перерегистрация имени с другим OID должна быть разрешена: %v", err)
	}
}

func TestRegistryLookupByName(t *testing.T) {
	reg := NewRegistry()
	oid := OID{1, 3, 6, 1, 4, 1}
	reg.Register("enterprise", oid)

	// Существующее имя
	result, exists := reg.LookupByName("enterprise")
	if !exists {
		t.Error("LookupByName: ожидалось, что OID существует")
	}
	if !result.Equal(oid) {
		t.Errorf("LookupByName = %v, ожидалось %v", result, oid)
	}

	// Несуществующее имя
	result, exists = reg.LookupByName("nonexistent")
	if exists {
		t.Error("LookupByName: несуществующее имя не должно быть найдено")
	}
}

func TestRegistryLookupByOID(t *testing.T) {
	reg := NewRegistry()
	oid := OID{1, 3, 6, 1, 4, 1}
	reg.Register("enterprise", oid)

	// Существующий OID
	name, exists := reg.LookupByOID(oid)
	if !exists {
		t.Error("LookupByOID: ожидалось, что имя существует")
	}
	if name != "enterprise" {
		t.Errorf("LookupByOID = %q, ожидалось %q", name, "enterprise")
	}

	// Несуществующий OID
	name, exists = reg.LookupByOID(OID{2, 100, 3})
	if exists {
		t.Error("LookupByOID: несуществующий OID не должен быть найден")
	}
}

func TestRegistryRemove(t *testing.T) {
	reg := NewRegistry()
	oid := OID{1, 3, 6, 1, 4, 1}
	reg.Register("enterprise", oid)

	// Удаление существующего имени
	if !reg.Remove("enterprise") {
		t.Error("Remove: ожидалось true для существующего имени")
	}

	// Проверка, что OID удален
	_, exists := reg.LookupByName("enterprise")
	if exists {
		t.Error("Remove: OID должен быть удален")
	}

	// Удаление несуществующего имени
	if reg.Remove("nonexistent") {
		t.Error("Remove: ожидалось false для несуществующего имени")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	// Пустой список
	list := reg.List()
	if len(list) != 0 {
		t.Errorf("List: ожидался пустой список, получено %d элементов", len(list))
	}

	// Добавляем элементы
	oid1 := OID{1, 3, 6, 1, 4, 1}
	oid2 := OID{2, 100, 3}
	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Проверяем список
	list = reg.List()
	if len(list) != 2 {
		t.Errorf("List: ожидалось 2 элемента, получено %d", len(list))
	}

	if !list["first"].Equal(oid1) {
		t.Errorf("List: элемент 'first' = %v, ожидалось %v", list["first"], oid1)
	}

	if !list["second"].Equal(oid2) {
		t.Errorf("List: элемент 'second' = %v, ожидалось %v", list["second"], oid2)
	}
}

func TestRegistryConcurrency(t *testing.T) {
	reg := NewRegistry()

	// Тестируем конкурентный доступ
	done := make(chan bool)

	// Горутина для записи
	go func() {
		for i := 0; i < 100; i++ {
			name := fmt.Sprintf("oid-%d", i)
			reg.Register(name, OID{1, 3, uint32(i)})
		}
		done <- true
	}()

	// Горутина для чтения
	go func() {
		for i := 0; i < 100; i++ {
			reg.LookupByName("nonexistent")
			reg.List()
		}
		done <- true
	}()

	// Ждем завершения
	<-done
	<-done

	// Проверяем, что все зарегистрировано
	list := reg.List()
	if len(list) != 100 {
		t.Errorf("Ожидалось 100 зарегистрированных OID, получено %d", len(list))
	}
}

func TestBatchRegisterSuccess(t *testing.T) {
	reg := NewRegistry()

	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
		"private":    MustParseOID("1.3.6.1.4.1.99999"),
		"iso":        MustParseOID("1.3.6.1"),
	}

	if err := reg.BatchRegister(entries); err != nil {
		t.Fatalf("Ошибка пакетной регистрации: %v", err)
	}

	if reg.Size() != 3 {
		t.Errorf("Size = %d, ожидалось 3", reg.Size())
	}

	// Проверяем, что все зарегистрированы
	for name, oid := range entries {
		stored, exists := reg.LookupByName(name)
		if !exists {
			t.Errorf("OID '%s' не найден", name)
			continue
		}
		if !stored.Equal(oid) {
			t.Errorf("OID '%s' = %v, ожидалось %v", name, stored, oid)
		}
	}
}

func TestBatchRegisterDuplicateInside(t *testing.T) {
	reg := NewRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1.4.1"),
		"second": MustParseOID("1.3.6.1.4.1"), // Тот же OID
	}

	err := reg.BatchRegister(entries)
	if err == nil {
		t.Error("Ожидалась ошибка дублирования OID внутри пакета")
	}

	if reg.Size() != 0 {
		t.Errorf("Size = %d, ожидалось 0 (атомарность нарушена)", reg.Size())
	}
}

func TestBatchRegisterConflictWithExisting(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем существующий OID
	existingOID := MustParseOID("1.3.6.1.4.1")
	reg.Register("existing", existingOID)

	// Пытаемся зарегистрировать тот же OID под другим именем
	entries := map[string]OID{
		"new":      MustParseOID("2.100.3"),
		"conflict": existingOID,
	}

	err := reg.BatchRegister(entries)
	if err == nil {
		t.Error("Ожидалась ошибка конфликта с существующим OID")
	}

	// Проверяем атомарность: ничего не должно быть зарегистрировано
	if reg.Size() != 1 {
		t.Errorf("Size = %d, ожидалось 1 (атомарность нарушена)", reg.Size())
	}

	if _, exists := reg.LookupByName("new"); exists {
		t.Error("OID 'new' не должен быть зарегистрирован")
	}
}

func TestBatchRegisterEmpty(t *testing.T) {
	reg := NewRegistry()

	if err := reg.BatchRegister(map[string]OID{}); err != nil {
		t.Errorf("Пустой пакет не должен давать ошибку: %v", err)
	}

	if reg.Size() != 0 {
		t.Errorf("Size = %d, ожидалось 0", reg.Size())
	}
}

func TestGlobalBatchRegister(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
		"private":    MustParseOID("1.3.6.1.4.1.99999"),
	}

	if err := BatchRegister(entries); err != nil {
		t.Fatalf("Ошибка глобальной пакетной регистрации: %v", err)
	}

	if Size() != 2 {
		t.Errorf("Size = %d, ожидалось 2", Size())
	}
}

func TestMustBatchRegister(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
	}

	// Успешная регистрация
	MustBatchRegister(entries)

	// Проверка паники при ошибке
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBatchRegister должен паниковать при ошибке")
		}
	}()

	// Пытаемся зарегистрировать тот же OID
	MustBatchRegister(entries)
}

func TestBatchRegisterDuplicateOID(t *testing.T) {
	reg := NewRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1.4.1"),
		"second": MustParseOID("1.3.6.1.4.1"), // Тот же OID
	}

	err := reg.BatchRegister(entries)
	if err == nil {
		t.Error("Ожидалась ошибка дублирования OID")
	}

	// Проверяем атомарность
	if reg.Size() != 0 {
		t.Errorf("Size = %d, ожидалось 0", reg.Size())
	}
}

func TestBatchRegisterDuplicateName(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем существующее имя
	reg.Register("test", MustParseOID("1.3.6.1.4.1"))

	// Пытаемся зарегистрировать то же имя с другим OID
	entries := map[string]OID{
		"test":     MustParseOID("2.100.3"),
		"new_name": MustParseOID("1.3.6.1.4.2"),
	}

	err := reg.BatchRegister(entries)
	if err == nil {
		t.Error("Ожидалась ошибка конфликта имени")
	}

	// Проверяем атомарность
	if reg.Size() != 1 {
		t.Errorf("Size = %d, ожидалось 1", reg.Size())
	}

	// Проверяем, что новый OID не зарегистрирован
	if _, exists := reg.LookupByName("new_name"); exists {
		t.Error("new_name не должен быть зарегистрирован")
	}
}

func TestBatchRegisterMixedConflicts(t *testing.T) {
	reg := NewRegistry()

	// Существующие записи
	reg.Register("existing_name", MustParseOID("1.3.6.1.4.1"))

	// Пакет с конфликтами
	entries := map[string]OID{
		"valid_name":    MustParseOID("2.100.3"),
		"existing_name": MustParseOID("1.3.6.1.4.2"), // Конфликт имени
		"conflict_oid":  MustParseOID("1.3.6.1.4.1"), // Конфликт OID
	}

	err := reg.BatchRegister(entries)
	if err == nil {
		t.Error("Ожидалась ошибка конфликта")
	}

	// Проверяем атомарность - ничего не должно быть добавлено
	if reg.Size() != 1 {
		t.Errorf("Size = %d, ожидалось 1", reg.Size())
	}

	if _, exists := reg.LookupByName("valid_name"); exists {
		t.Error("valid_name не должен быть зарегистрирован")
	}
}

// Профилирование аллокаций
func TestRegistryAllocations(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	t.Log("=== Профиль аллокаций ===")

	allocs := testing.AllocsPerRun(1000, func() {
		reg.Register("new", oid)
	})
	t.Logf("Register: %.1f allocs/op", allocs)

	allocs = testing.AllocsPerRun(1000, func() {
		_, _ = reg.LookupByName("test")
	})
	t.Logf("LookupByName: %.1f allocs/op", allocs)

	allocs = testing.AllocsPerRun(1000, func() {
		_, _ = reg.LookupByOID(oid)
	})
	t.Logf("LookupByOID: %.1f allocs/op", allocs)

	allocs = testing.AllocsPerRun(1000, func() {
		_ = reg.Contains("test")
	})
	t.Logf("Contains: %.1f allocs/op", allocs)

	allocs = testing.AllocsPerRun(1000, func() {
		_ = reg.List()
	})
	t.Logf("List: %.1f allocs/op", allocs)
}

// ============================================
// БЕНЧМАРКИ
// ============================================

func BenchmarkRegistryLookupByName(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByName("test")
	}
}

func BenchmarkRegistryLookupByOID(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByOID(oid)
	}
}

func BenchmarkRegistryRemoveOptimized(b *testing.B) {
	// Подготавливаем реестр с записями
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Регистрируем и удаляем одну и ту же запись
		name := "test"
		reg.Register(name, oid)
		reg.Remove(name)
	}
}

func BenchmarkRegistryListOptimized(b *testing.B) {
	reg := NewRegistry()

	// Подготавливаем 10 записей
	for i := 0; i < 10; i++ {
		reg.Register(fmt.Sprintf("oid-%d", i), MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.List()
	}
}

// Бенчмарк с переиспользованием реестра
func BenchmarkRegistryReuseRegister(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	counter := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		name := fmt.Sprintf("oid-%d", counter%1000) // Переиспользуем имена
		counter++
		reg.Register(name, oid)
	}
}

// ============ РАЗМЕРЫ РЕЕСТРА ============

func BenchmarkRegistrySmall(b *testing.B) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Empty", 0},
		{"Tiny_1", 1},
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			reg := NewRegistry()
			for i := 0; i < tc.count; i++ {
				reg.Register(fmt.Sprintf("oid-%d", i),
					MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, _ = reg.LookupByName("oid-0")
			}
		})
	}
}

// ============ КОНКУРЕНТНЫЕ ОПЕРАЦИИ ============

// Параллельные бенчмарки всегда используют pb.Next()
func BenchmarkRegistryParallelRead(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = reg.LookupByName("test")
		}
	})
}

func BenchmarkRegistryParallelWrite(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			name := fmt.Sprintf("oid-%d", counter%1000)
			counter++
			_ = reg.Register(name, oid)
		}
	})
}

func BenchmarkRegistryParallelMixed(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if counter%2 == 0 {
				_, _ = reg.LookupByName("test")
			} else {
				name := fmt.Sprintf("oid-%d", counter%100)
				_ = reg.Register(name, oid)
			}
			counter++
		}
	})
}

func BenchmarkRegistryLookupByNamePreallocatedParallel(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	names := make([]string, 100)
	for i := range names {
		names[i] = fmt.Sprintf("oid-%d", i)
		reg.Register(names[i], oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			_, _ = reg.LookupByName(names[counter%100])
			counter++
		}
	})
}

func BenchmarkRegistryLookupByOIDPreallocatedParallel(b *testing.B) {
	reg := NewRegistry()

	oids := make([]OID, 100)
	for i := range oids {
		oids[i] = MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oids[i])
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			_, _ = reg.LookupByOID(oids[counter%100])
			counter++
		}
	})
}

// ============ ПАКЕТНЫЕ ОПЕРАЦИИ ============

func BenchmarkRegistryBatchRegister(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()
		entries := map[string]OID{
			"enterprise": MustParseOID("1.3.6.1.4.1"),
			"private":    MustParseOID("1.3.6.1.4.1.99999"),
			"iso":        MustParseOID("1.3.6.1"),
			"org":        MustParseOID("1.3"),
			"dod":        MustParseOID("1.3.6"),
		}

		if err := reg.BatchRegister(entries); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegistryBatchLookup(b *testing.B) {
	reg := NewRegistry()
	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
		"private":    MustParseOID("1.3.6.1.4.1.99999"),
		"iso":        MustParseOID("1.3.6.1"),
		"org":        MustParseOID("1.3"),
		"dod":        MustParseOID("1.3.6"),
	}
	reg.BatchRegister(entries)

	names := []string{"enterprise", "private", "iso", "org", "dod"}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, name := range names {
			_, _ = reg.LookupByName(name)
		}
	}
}

// ============ ГЛОБАЛЬНЫЙ РЕЕСТР ============

func BenchmarkGlobalRegister(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry() // Очищаем для повторной регистрации
		if err := Register("test", oid); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGlobalLookup(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = LookupByName("test")
	}
}

func BenchmarkGlobalBatchRegister(b *testing.B) {
	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
		"private":    MustParseOID("1.3.6.1.4.1.99999"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		if err := BatchRegister(entries); err != nil {
			b.Fatal(err)
		}
	}
}

// ============ СПЕЦИФИЧНЫЕ СЦЕНАРИИ ============

func BenchmarkRegistryRegisterUnique(b *testing.B) {
	reg := NewRegistry()

	counter := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		oid := OID{1, 3, uint32(counter + 1)}
		name := fmt.Sprintf("oid-%d", counter)
		counter++
		if err := reg.Register(name, oid); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegistryLookupMiss(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByName("nonexistent")
	}
}

func BenchmarkRegistryContains(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.Contains("test")
	}
}

func BenchmarkRegistryClearOptimized(b *testing.B) {
	// Подготавливаем OID заранее
	oids := make([]OID, 10)
	names := make([]string, 10)
	for i := 0; i < 10; i++ {
		oids[i] = MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1))
		names[i] = fmt.Sprintf("oid-%d", i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()

		// Быстрая регистрация
		for j := 0; j < 10; j++ {
			reg.Register(names[j], oids[j])
		}

		// Очистка
		reg.Clear()
	}
}

// ============ СРАВНЕНИЕ С MAP ============

func BenchmarkRegistryVsMap(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.Run("Registry", func(b *testing.B) {
		reg := NewRegistry()
		reg.Register("test", oid)

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByName("test")
		}
	})

	b.Run("Map", func(b *testing.B) {
		m := make(map[string]OID)
		m["test"] = oid

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = m["test"]
		}
	})
}

func BenchmarkRegistryVsMapParallel(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.Run("Registry", func(b *testing.B) {
		reg := NewRegistry()
		reg.Register("test", oid)

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = reg.LookupByName("test")
			}
		})
	})

	b.Run("Map_with_RWMutex", func(b *testing.B) {
		var mu sync.RWMutex
		m := make(map[string]OID)
		m["test"] = oid

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.RLock()
				_ = m["test"]
				mu.RUnlock()
			}
		})
	})
}

// ============ СТРЕСС-ТЕСТЫ ============

func BenchmarkRegistryStressRead(b *testing.B) {
	reg := NewRegistry()
	// Добавляем 1000 записей
	for i := 0; i < 1000; i++ {
		reg.Register(fmt.Sprintf("oid-%d", i), MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			name := fmt.Sprintf("oid-%d", counter%1000)
			_, _ = reg.LookupByName(name)
			counter++
		}
	})
}

func BenchmarkRegistryStressMixed(b *testing.B) {
	reg := NewRegistry()
	// Начальные записи
	for i := 0; i < 100; i++ {
		reg.Register(fmt.Sprintf("initial-%d", i), MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			op := counter % 10
			switch op {
			case 0, 1, 2, 3, 4: // 50% чтение
				name := fmt.Sprintf("initial-%d", counter%100)
				_, _ = reg.LookupByName(name)
			case 5, 6, 7: // 30% запись
				name := fmt.Sprintf("new-%d", counter%1000)
				_ = reg.Register(name, MustParseOID(fmt.Sprintf("1.3.6.1.%d", counter%1000+1)))
			case 8: // 10% удаление
				name := fmt.Sprintf("new-%d", (counter-1)%1000)
				reg.Remove(name)
			case 9: // 10% листинг
				_ = reg.List()
			}
			counter++
		}
	})
}

// Бенчмарк с предварительно вычисленными строками
func BenchmarkRegistryLookupByNamePreallocated(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	names := make([]string, 100)
	for i := range names {
		names[i] = fmt.Sprintf("oid-%d", i)
		reg.Register(names[i], oid)
	}

	counter := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByName(names[counter%100])
		counter++
	}
}

// Бенчмарк с предварительно вычисленными OID
func BenchmarkRegistryLookupByOIDPreallocated(b *testing.B) {
	reg := NewRegistry()

	oids := make([]OID, 100)
	for i := range oids {
		oids[i] = MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oids[i])
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByOID(oids[0]) // Всегда первый OID
	}
}

// Бенчмарк BatchRegister с предварительной подготовкой
func BenchmarkRegistryBatchRegisterPrepared(b *testing.B) {
	// Предварительно подготавливаем entries
	entries := map[string]OID{
		"enterprise": MustParseOID("1.3.6.1.4.1"),
		"private":    MustParseOID("1.3.6.1.4.1.99999"),
		"iso":        MustParseOID("1.3.6.1"),
		"org":        MustParseOID("1.3"),
		"dod":        MustParseOID("1.3.6"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()
		if err := reg.BatchRegister(entries); err != nil {
			b.Fatal(err)
		}
	}
}

// Бенчмарк для измерения чистого времени операций
func BenchmarkRegistryOperations(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.Run("Register", func(b *testing.B) {
		counter := 0
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			reg.Register(fmt.Sprintf("oid-%d", counter), oid)
			counter++
		}
	})

	b.Run("Lookup_Hit", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByName("test")
		}
	})

	b.Run("Lookup_Miss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByName("nonexistent")
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = reg.Contains("test")
		}
	})
}

// Сравнение с map для понимания overhead
func BenchmarkRegistryOverhead(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")
	key := oid.String()

	b.Run("Registry_Lookup", func(b *testing.B) {
		reg := NewRegistry()
		reg.Register("test", oid)

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByName("test")
		}
	})

	b.Run("Map_Lookup", func(b *testing.B) {
		m := make(map[string]OID)
		m["test"] = oid

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = m["test"]
		}
	})

	b.Run("Map_String_Lookup", func(b *testing.B) {
		m := make(map[string]string)
		m["test"] = key

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = m["test"]
		}
	})
}
