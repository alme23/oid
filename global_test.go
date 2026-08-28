// global_test.go
package oid

import (
	"testing"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestGlobalRegister(t *testing.T) {
	// Очищаем перед тестом
	ResetRegistry()

	oid := MustParseOID("1.3.6.1.4.1")

	if err := Register("enterprise", oid); err != nil {
		t.Fatalf("Ошибка регистрации: %v", err)
	}

	if Size() != 1 {
		t.Errorf("Size = %d, ожидалось 1", Size())
	}

	if !Contains("enterprise") {
		t.Error("Contains: имя не найдено")
	}
}

func TestGlobalLookup(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1.4.1")
	Register("enterprise", oid)

	// LookupByName
	result, exists := LookupByName("enterprise")
	if !exists {
		t.Error("LookupByName: OID не найден")
	}
	if !result.Equal(oid) {
		t.Errorf("LookupByName = %v, ожидалось %v", result, oid)
	}

	// LookupByOID
	name, exists := LookupByOID(oid)
	if !exists {
		t.Error("LookupByOID: имя не найдено")
	}
	if name != "enterprise" {
		t.Errorf("LookupByOID = %q, ожидалось 'enterprise'", name)
	}
}

func TestGlobalRemove(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1.4.1")
	Register("enterprise", oid)

	if !Remove("enterprise") {
		t.Error("Remove: ожидалось true")
	}

	if Size() != 0 {
		t.Errorf("Size после Remove = %d, ожидалось 0", Size())
	}

	if Contains("enterprise") {
		t.Error("Contains: имя должно быть удалено")
	}
}

func TestGlobalClear(t *testing.T) {
	ResetRegistry()

	Register("oid1", MustParseOID("1.3.6.1.4.1"))
	Register("oid2", MustParseOID("2.100.3"))

	if Size() != 2 {
		t.Errorf("Size до Clear = %d, ожидалось 2", Size())
	}

	Clear()

	if Size() != 0 {
		t.Errorf("Size после Clear = %d, ожидалось 0", Size())
	}
}

func TestGlobalDiff(t *testing.T) {
	ResetRegistry()

	// Снимок пустого реестра
	snapshot := Snapshot()

	// Добавляем OID
	oid1 := MustParseOID("1.3.6.1.4.1")
	Register("enterprise", oid1)

	added, removed, changed := Diff(snapshot)
	if len(added) != 1 {
		t.Errorf("Добавлено = %d, ожидалось 1", len(added))
	}
	if len(removed) != 0 {
		t.Errorf("Удалено = %d, ожидалось 0", len(removed))
	}
	if len(changed) != 0 {
		t.Errorf("Изменено = %d, ожидалось 0", len(changed))
	}
}

func TestGlobalMustRegister(t *testing.T) {
	ResetRegistry()

	// Успешная регистрация
	MustRegister("enterprise", MustParseOID("1.3.6.1.4.1"))

	// Проверка паники при дубликате
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegister должен паниковать при дубликате")
		}
	}()

	MustRegister("enterprise2", MustParseOID("1.3.6.1.4.1"))
}

func TestGlobalEdgeCases(t *testing.T) {
	ResetRegistry()

	// Регистрация
	if err := Register("test", MustParseOID("1.3.6.1")); err != nil {
		t.Errorf("Register: %v", err)
	}

	// Дубликат
	if err := Register("test2", MustParseOID("1.3.6.1")); err == nil {
		t.Error("Register: ожидалась ошибка дубликата")
	}

	// Snapshot
	snapshot := Snapshot()
	if len(snapshot) != 1 {
		t.Error("Snapshot: неверный размер")
	}

	// Diff
	added, removed, changed := Diff(snapshot)
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Error("Diff: должна быть пустой")
	}

	// Clear
	Clear()
	if Size() != 0 {
		t.Error("Clear: размер должен быть 0")
	}
}
