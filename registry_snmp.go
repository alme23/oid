package oid

// SNMPRegistry хранит скалярные и колумнарные OID
type SNMPRegistry struct {
	scalars map[string]*ScalarOID
	columns map[string]*ColumnarOID
}

// NewSNMPRegistry создает реестр SNMP OID
func NewSNMPRegistry() *SNMPRegistry {
	return &SNMPRegistry{
		scalars: make(map[string]*ScalarOID),
		columns: make(map[string]*ColumnarOID),
	}
}

// RegisterScalar регистрирует скалярный OID
func (r *SNMPRegistry) RegisterScalar(name string, oid ScalarOID) {
	// Создаем копию и храним указатель
	oidCopy := make(ScalarOID, len(oid))
	copy(oidCopy, oid)
	r.scalars[name] = &oidCopy
}

// RegisterScalarNoCopy регистрирует скалярный OID без копирования (хранит ссылку)
func (r *SNMPRegistry) RegisterScalarNoCopy(name string, oid ScalarOID) {
	r.scalars[name] = &oid
}

// RegisterColumn регистрирует колумнарный OID
func (r *SNMPRegistry) RegisterColumn(name string, oid ColumnarOID) {
	// Создаем глубокую копию и храним указатель
	colCopy := ColumnarOID{
		Base:    make(OID, len(oid.Base)),
		Column:  oid.Column,
		Indexes: make([]uint32, len(oid.Indexes)),
	}
	copy(colCopy.Base, oid.Base)
	copy(colCopy.Indexes, oid.Indexes)
	r.columns[name] = &colCopy
}

// RegisterColumnNoCopy регистрирует указатель на колонку
func (r *SNMPRegistry) RegisterColumnNoCopy(name string, oid *ColumnarOID) {
	r.columns[name] = oid
}

// GetScalar возвращает глубокую копию скалярного OID
func (r *SNMPRegistry) GetScalar(name string) (ScalarOID, bool) {
	oid, exists := r.scalars[name]
	if !exists || oid == nil {
		return nil, exists
	}

	// Для пустого OID возвращаем nil
	if len(*oid) == 0 {
		return nil, true
	}

	// Глубокая копия
	oidCopy := make(ScalarOID, len(*oid))
	copy(oidCopy, *oid)
	return oidCopy, true
}

// GetScalarNoCopy возвращает указатель на OID
func (r *SNMPRegistry) GetScalarNoCopy(name string) (*ScalarOID, bool) {
	oid, exists := r.scalars[name]
	return oid, exists
}

// GetColumn возвращает глубокую копию колумнарного OID
func (r *SNMPRegistry) GetColumn(name string) (ColumnarOID, bool) {
	col, exists := r.columns[name]
	if !exists || col == nil {
		return ColumnarOID{}, exists
	}

	// Глубокая копия
	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)
	return colCopy, true
}

// GetColumnNoCopy возвращает указатель на колонку
func (r *SNMPRegistry) GetColumnNoCopy(name string) (*ColumnarOID, bool) {
	col, exists := r.columns[name]
	return col, exists
}

// GetColumnWithIndexes возвращает OID с индексами
func (r *SNMPRegistry) GetColumnWithIndexes(name string, indexes ...uint32) (OID, bool) {
	col, exists := r.columns[name]
	if !exists || col == nil {
		return nil, false
	}
	return col.WithIndexes(indexes...).FullOID(), true
}
