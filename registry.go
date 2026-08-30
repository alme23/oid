package oid

import (
	"fmt"
	"maps"
	"sync"
)

// Registry хранит именованные OID. Полностью потокобезопасен.
type Registry struct {
	mu    sync.RWMutex
	names map[string]OID
	oids  map[string]string
}

// NewRegistry создает новый реестр OID
func NewRegistry() *Registry {
	return &Registry{
		names: make(map[string]OID),
		oids:  make(map[string]string),
	}
}

// Register регистрирует OID под указанным именем
func (r *Registry) Register(name string, oid OID) (err error) {
	if err := oid.Validate(); err != nil {
		return err
	}

	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)
	key := oidCopy.String()

	r.mu.Lock()
	defer r.mu.Unlock()

	if existingName, exists := r.oids[key]; exists && existingName != name {
		return fmt.Errorf("%w: %s как '%s'", ErrOIDAlreadyRegistered, key, existingName)
	}

	r.names[name] = oidCopy
	r.oids[key] = name
	return nil
}

// LookupByName возвращает копию OID по имени
func (r *Registry) LookupByName(name string) (oid OID, exists bool) {
	r.mu.RLock()
	oid, exists = r.names[name]
	r.mu.RUnlock()

	if !exists {
		return nil, false
	}

	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)
	return oidCopy, true
}

// LookupByNameNoCopy возвращает прямую ссылку на внутренний срез OID
func (r *Registry) LookupByNameNoCopy(name string) (oid OID, exists bool) {
	r.mu.RLock()
	oid, exists = r.names[name]
	r.mu.RUnlock()
	return oid, exists
}

// LookupByOID возвращает имя по OID
func (r *Registry) LookupByOID(oid OID) (name string, exists bool) {
	key := oid.String()

	r.mu.RLock()
	name, exists = r.oids[key]
	r.mu.RUnlock()

	return name, exists
}

// Remove удаляет OID из реестра
func (r *Registry) Remove(name string) (removed bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oid, exists := r.names[name]
	if !exists {
		return false
	}

	key := oid.String()
	delete(r.names, name)
	delete(r.oids, key)
	return true
}

// List возвращает изолированную глубокую копию всех зарегистрированных OID
func (r *Registry) List() (result map[string]OID) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result = make(map[string]OID, len(r.names))

	// Поверхностное копирование через maps.Copy
	maps.Copy(result, r.names)

	// Глубокое копирование OID
	for name, oid := range result {
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)
		result[name] = oidCopy
	}

	return result
}

// ListNoCopy возвращает новую карту со ссылками на внутренние срезы OID
func (r *Registry) ListNoCopy() (result map[string]OID) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result = make(map[string]OID, len(r.names))

	// Копирование через maps.Copy
	maps.Copy(result, r.names)

	return result
}

// Size возвращает количество зарегистрированных OID
func (r *Registry) Size() (size int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.names)
}

// Contains проверяет наличие имени в реестре
func (r *Registry) Contains(name string) (exists bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists = r.names[name]
	return exists
}

// Clear удаляет все записи из реестра
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.names = make(map[string]OID)
	r.oids = make(map[string]string)
}

// Names возвращает срез всех зарегистрированных имен
func (r *Registry) Names() (names []string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names = make([]string, 0, len(r.names))
	for name := range r.names {
		names = append(names, name)
	}
	return names
}

// OIDs возвращает глубокую копию всех OID в реестре
func (r *Registry) OIDs() (oids []OID) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	oids = make([]OID, 0, len(r.names))
	for _, oid := range r.names {
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)
		oids = append(oids, oidCopy)
	}
	return oids
}

// OIDsNoCopy возвращает срез ссылок на внутренние OID
func (r *Registry) OIDsNoCopy() (oids []OID) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	oids = make([]OID, 0, len(r.names))
	for _, oid := range r.names {
		oids = append(oids, oid)
	}
	return oids
}

// BatchRegister регистрирует несколько OID атомарно
func (r *Registry) BatchRegister(entries map[string]OID) (err error) {
	if len(entries) == 0 {
		return nil
	}

	type preparedEntry struct {
		name    string
		oidCopy OID
		key     string
	}

	prepared := make([]preparedEntry, 0, len(entries))
	for name, oid := range entries {
		if err := oid.Validate(); err != nil {
			return err
		}

		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)
		key := oidCopy.String()

		prepared = append(prepared, preparedEntry{
			name:    name,
			oidCopy: oidCopy,
			key:     key,
		})
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	seenNames := make(map[string]struct{}, len(prepared))
	seenKeys := make(map[string]struct{}, len(prepared))

	for _, entry := range prepared {
		if _, duplicate := seenNames[entry.name]; duplicate {
			return fmt.Errorf("%w: '%s'", ErrDuplicateNameInBatch, entry.name)
		}
		if _, duplicate := seenKeys[entry.key]; duplicate {
			return fmt.Errorf("%w: '%s'", ErrDuplicateOIDInBatch, entry.oidCopy)
		}

		if _, exists := r.names[entry.name]; exists {
			return fmt.Errorf("%w: '%s'", ErrNameAlreadyExists, entry.name)
		}
		if existingName, exists := r.oids[entry.key]; exists {
			return fmt.Errorf("%w: %s как '%s'", ErrOIDAlreadyRegistered, entry.oidCopy, existingName)
		}

		seenNames[entry.name] = struct{}{}
		seenKeys[entry.key] = struct{}{}
	}

	for _, entry := range prepared {
		r.names[entry.name] = entry.oidCopy
		r.oids[entry.key] = entry.name
	}

	return nil
}
