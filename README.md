# OID Package for Go

Высокопроизводительный пакет для работы с Object Identifiers (OID) в Go.

## Особенности

- 🚀 **Высокая производительность**: оптимизированные алгоритмы ASN.1 кодирования
- 💾 **Низкое потребление памяти**: минимальное количество аллокаций
- 🔒 **Потокобезопасность**: все операции с реестром защищены мьютексами
- 📦 **Готовые интеграции**: database/sql, JSON, ASN.1
- 🎯 **Полное соответствие стандартам**: RFC 3061, RFC 4517

## Установка

```bash
go get github.com/alme23/oid
```

## Быстрый старт

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание OID
    enterpriseOID := oid.MustParseOID("1.3.6.1.4.1")
    
    // Строковое представление
    fmt.Println(enterpriseOID.String()) // 1.3.6.1.4.1
    
    // Сравнение
    oid1 := oid.MustParseOID("1.3.6.1.4.1")
    oid2 := oid.MustParseOID("1.3.6.1.4.1")
    fmt.Println(oid1.Equal(oid2)) // true
    
    // Проверка префикса
    childOID := oid.MustParseOID("1.3.6.1.4.1.99999")
    fmt.Println(childOID.StartsWith(enterpriseOID)) // true
}
```

## API

### Создание OID

```go
// Парсинг из строки
oid, err := oid.ParseOID("1.3.6.1.4.1")

// Парсинг с паникой при ошибке
oid := oid.MustParseOID("1.3.6.1.4.1")

// Создание из компонентов
oid := oid.OID{1, 3, 6, 1, 4, 1}
```

### Операции с OID

```go
// Строковое представление
str := oid.String()

// Добавление компонентов
newOID := oid.Append(99999, 1, 1)

// Получение родителя
parent, err := oid.Parent()

// Получение последнего компонента
last, err := oid.Last()

// Сравнение
equal := oid1.Equal(oid2)

// Проверка префикса
startsWith := oid.StartsWith(prefix)
```

### Сериализация

```go
// ASN.1 DER
data, err := oid.MarshalBinary()
err = oid.UnmarshalBinary(data)

// ASN.1 BER
data, err := oid.MarshalBER()
err = oid.UnmarshalBER(data)

// JSON
jsonData, err := oid.MarshalJSON()
err = oid.UnmarshalJSON(jsonData)
```

### Реестр OID

```go
// Создание реестра
registry := oid.NewRegistry()

// Регистрация
err := registry.Register("enterprise", oid)

// Поиск с копированием (безопасно)
oid, exists := registry.LookupByName("enterprise")

// Поиск без копирования (быстро)
oidNoCopy, exists := registry.LookupByNameNoCopy("enterprise")

// Поиск по OID
name, exists := registry.LookupByOID(oid)

// Удаление
removed := registry.Remove("enterprise")

// Список всех записей
allOIDs := registry.List()
```

### Глобальный реестр

```go
// Регистрация в глобальном реестре
oid.MustRegister("enterprise", oid)

// Поиск
oid, exists := oid.LookupByName("enterprise")

// Пакетная регистрация
entries := map[string]oid.OID{
    "enterprise": oid.MustParseOID("1.3.6.1.4.1"),
    "private":    oid.MustParseOID("1.3.6.1.4.1.99999"),
}
err := oid.BatchRegister(entries)
```

### Работа с базой данных

```go
// Использование в SQL запросах
var oid oid.OID
err := db.QueryRow("SELECT oid FROM table WHERE id = $1", 1).Scan(&oid)

// NULL-значения
var nullableOID oid.NullOID
err := db.QueryRow("SELECT oid FROM table WHERE id = $1", 1).Scan(&nullableOID)
if nullableOID.Valid {
    fmt.Println(nullableOID.OID)
}

// Массивы OID (PostgreSQL)
var oids oid.Array
err := db.QueryRow("SELECT oids FROM table WHERE id = $1", 1).Scan(&oids)
```

### SNMP Registry

```go
// Создание SNMP реестра
reg := oid.NewSNMPRegistry()

// Регистрация скалярных OID
reg.RegisterScalar("sysDescr", oid.MustScalarOID("1.3.6.1.2.1.1.1.0"))
reg.RegisterScalar("sysUpTime", oid.MustScalarOID("1.3.6.1.2.1.1.3.0"))

// Регистрация колумнарных OID
base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
reg.RegisterColumn("ifDescr", oid.NewColumnarOID(base, 2))

// Получение с индексами
ifDescrOID, exists := reg.GetColumnWithIndexes("ifDescr", 1)
```

### TableOID

```go
// Создание таблицы
table := oid.NewTableOID(oid.MustParseOID("1.3.6.1.2.1.2.2.1"))

// Добавление колонок
table.AddColumn("ifIndex", 1)
table.AddColumn("ifDescr", 2)
table.AddColumn("ifType", 3)

// Получение OID колонки
oid, err := table.GetColumnOID("ifDescr", 1)

// Парсинг OID строки
column, index, err := table.ParseRowOID(oid)
```

### ScalarOID

```go
// Создание скалярного OID
scalar := oid.MustScalarOID("1.3.6.1.2.1.1.1.0")

// Проверка
if scalar.IsScalar() {
    fmt.Println("Это скалярный OID")
}

// Получение базы (без .0)
base := scalar.Base()
```

### ColumnarOID

```go
// Создание колумнарного OID
base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
col := oid.NewColumnarOID(base, 2, 1)

// Полный OID
fullOID := col.FullOID()

// Строковое представление
str := col.String()

// Индексы
if col.HasIndexes() {
    fmt.Println(col.IndexString())
}
```

## Производительность

Результаты бенчмарков на Intel Core i5-7260U:

|Операция|Время (ns/op)|Память (B/op)|Аллокации|
|--|--|--|--|
|ParseOID|65-169|16-48|1|
|String|44-133|8-29|1-2|
|MarshalBinary|56-58|16|1|
|MarshalBER|40-74|5-16|1|
|UnmarshalBinary|92-96|72|2|
|UnmarshalBER|73-189|40-184|2-4|
|NoCopy Lookup|24|0|0|
|Contains|25|0|0|
|Size|16|0|0|

### Сравнение со стандартной библиотекой
|Операция|oid|Стандартная|Ускорение|
|--|--|--|--|
|MarshalBER|40-74 ns|261-315 ns|5-6x|
|UnmarshalBER|73-189 ns|137-177 ns|1.4-1.9x|

## Примеры использования пакета OID

### Базовые операции с OID

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание OID
    oid1 := oid.MustParseOID("1.3.6.1.4.1")
    oid2 := oid.MustParseOID("1.3.6.1.4.1.99999")
    
    // Строковое представление
    fmt.Println(oid1.String()) // 1.3.6.1.4.1
    
    // Сравнение
    fmt.Println(oid1.Equal(oid2)) // false
    
    // Проверка префикса
    fmt.Println(oid2.StartsWith(oid1)) // true
    
    // Добавление компонентов
    extended := oid1.Append(99999, 1, 1)
    fmt.Println(extended) // 1.3.6.1.4.1.99999.1.1
    
    // Получение родителя
    parent, _ := extended.Parent()
    fmt.Println(parent) // 1.3.6.1.4.1.99999.1
    
    // Получение последнего компонента
    last, _ := extended.Last()
    fmt.Println(last) // 1
    
    // Валидация
    fmt.Println(oid1.Validate()) // nil
    fmt.Println(oid.OID{3, 1}.Validate()) // ошибка
}
```
### Сериализация

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    oid := oid.MustParseOID("1.3.6.1.4.1")
    
    // ASN.1 DER
    derData, _ := oid.MarshalBinary()
    fmt.Printf("DER: %x\n", derData)
    
    var decodedDER oid.OID
    decodedDER.UnmarshalBinary(derData)
    
    // ASN.1 BER
    berData, _ := oid.MarshalBER()
    fmt.Printf("BER: %x\n", berData)
    
    var decodedBER oid.OID
    decodedBER.UnmarshalBER(berData)
    
    // JSON
    jsonData, _ := json.Marshal(oid)
    fmt.Printf("JSON: %s\n", jsonData)
    
    var decodedJSON oid.OID
    json.Unmarshal(jsonData, &decodedJSON)
}
```

### Registry

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание реестра
    reg := oid.NewRegistry()
    
    // Регистрация
    reg.Register("enterprise", oid.MustParseOID("1.3.6.1.4.1"))
    reg.Register("private", oid.MustParseOID("1.3.6.1.4.1.99999"))
    
    // Поиск с копированием
    if o, exists := reg.LookupByName("enterprise"); exists {
        fmt.Println(o) // 1.3.6.1.4.1
    }
    
    // Поиск без копирования (быстро)
    if o, exists := reg.LookupByNameNoCopy("enterprise"); exists {
        fmt.Println(o) // 1.3.6.1.4.1
    }
    
    // Поиск по OID
    if name, exists := reg.LookupByOID(oid.MustParseOID("1.3.6.1.4.1")); exists {
        fmt.Println(name) // enterprise
    }
    
    // Проверка наличия
    fmt.Println(reg.Contains("enterprise")) // true
    
    // Размер
    fmt.Println(reg.Size()) // 2
    
    // Список всех
    list := reg.List()
    for name, o := range list {
        fmt.Printf("%s: %s\n", name, o)
    }
    
    // Удаление
    reg.Remove("enterprise")
    
    // Очистка
    reg.Clear()
}
```

### Глобальный реестр

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Регистрация
    oid.MustRegister("enterprise", oid.MustParseOID("1.3.6.1.4.1"))
    
    // Поиск
    if o, exists := oid.LookupByName("enterprise"); exists {
        fmt.Println(o)
    }
    
    // Пакетная регистрация
    entries := map[string]oid.OID{
        "first":  oid.MustParseOID("1.3.6.1"),
        "second": oid.MustParseOID("2.100.3"),
    }
    oid.BatchRegister(entries)
    
    // Проверка
    fmt.Println(oid.Size()) // 3
    
    // Снимок
    snapshot := oid.Snapshot()
    
    // Добавляем ещё
    oid.MustRegister("third", oid.MustParseOID("0.39.1"))
    
    // Diff
    added, removed, changed := oid.Diff(snapshot)
    fmt.Printf("Added: %d, Removed: %d, Changed: %d\n", 
        len(added), len(removed), len(changed))
    
    // Очистка
    oid.Clear()
    
    // Сброс
    oid.ResetRegistry()
}
```

### Работа с базой данных

```go
package main

import (
    "database/sql"
    "fmt"
    "github.com/alme23/oid"
    _ "github.com/lib/pq"
)

func main() {
    db, _ := sql.Open("postgres", "postgres://user:pass@localhost/db")
    defer db.Close()
    
    // Создание таблицы
    db.Exec(`CREATE TABLE IF NOT EXISTS devices (
        id SERIAL PRIMARY KEY,
        device_oid TEXT
    )`)
    
    // Вставка
    deviceOID := oid.MustParseOID("1.3.6.1.4.1.99999.1.1")
    db.Exec("INSERT INTO devices (device_oid) VALUES ($1)", deviceOID)
    
    // Чтение
    var oid oid.OID
    db.QueryRow("SELECT device_oid FROM devices WHERE id = 1").Scan(&oid)
    fmt.Println(oid)
    
    // NULL значения
    var nullableOID oid.NullOID
    db.QueryRow("SELECT device_oid FROM devices WHERE id = 999").Scan(&nullableOID)
    if nullableOID.Valid {
        fmt.Println(nullableOID.OID)
    } else {
        fmt.Println("NULL")
    }
    
    // Массивы
    var oids oid.Array
    db.QueryRow("SELECT ARRAY['1.3.6.1','2.100.3']::text[]").Scan(&oids)
    fmt.Println(oids)
}
```

### SNMP Registry

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    reg := oid.NewSNMPRegistry()
    
    // Скалярные OID
    reg.RegisterScalar("sysDescr", oid.MustScalarOID("1.3.6.1.2.1.1.1.0"))
    reg.RegisterScalar("sysUpTime", oid.MustScalarOID("1.3.6.1.2.1.1.3.0"))
    
    // Колумнарные OID
    base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
    reg.RegisterColumn("ifDescr", oid.NewColumnarOID(base, 2))
    reg.RegisterColumn("ifType", oid.NewColumnarOID(base, 3))
    
    // Получение скалярного
    if o, exists := reg.GetScalar("sysDescr"); exists {
        fmt.Println(o)
    }
    
    // Получение колумнарного с индексом
    if o, exists := reg.GetColumnWithIndexes("ifDescr", 1); exists {
        fmt.Println(o) // 1.3.6.1.2.1.2.2.1.2.1
    }
}
```

### TableOID

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание таблицы
    table := oid.NewTableOID(oid.MustParseOID("1.3.6.1.2.1.2.2.1"))
    
    // Добавление колонок
    table.AddColumn("ifIndex", 1)
    table.AddColumn("ifDescr", 2)
    table.AddColumn("ifType", 3)
    
    // Получение OID колонки
    oid, _ := table.GetColumnOID("ifDescr", 1)
    fmt.Println(oid) // 1.3.6.1.2.1.2.2.1.2.1
    
    // Получение строки
    row, _ := table.GetRowOID(1)
    for name, o := range row {
        fmt.Printf("%s: %s\n", name, o)
    }
    
    // Парсинг
    column, index, _ := table.ParseRowOID(oid)
    fmt.Printf("Column: %d, Index: %d\n", column, index)
    
    // Проверка
    fmt.Println(table.IsColumnOID(oid)) // true
    
    // Имена колонок
    fmt.Println(table.GetColumnNames()) // [ifDescr ifIndex ifType]
}
```

### ScalarOID

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание скалярного OID
    scalar := oid.MustScalarOID("1.3.6.1.2.1.1.1.0")
    
    // Проверка
    fmt.Println(scalar.IsScalar()) // true
    
    // Получение базы (без .0)
    base := scalar.Base()
    fmt.Println(base) // 1.3.6.1.2.1.1.1
    
    // Добавление компонентов
    extended := scalar.Append(1)
    fmt.Println(extended) // 1.3.6.1.2.1.1.1.0.1
    
    // Родитель
    parent, _ := scalar.Parent()
    fmt.Println(parent) // 1.3.6.1.2.1.1.1
}
```

### ColumnarOID

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
    
    // Создание колумнарного OID
    col := oid.NewColumnarOID(base, 2, 1)
    
    // Полный OID
    fullOID := col.FullOID()
    fmt.Println(fullOID) // 1.3.6.1.2.1.2.2.1.2.1
    
    // Строковое представление
    fmt.Println(col.String()) // 1.3.6.1.2.1.2.2.1.2.1
    
    // Индексы
    fmt.Println(col.HasIndexes()) // true
    fmt.Println(col.IndexString()) // 1
    
    // Добавление индекса
    col2 := col.AppendIndex(2)
    fmt.Println(col2.IndexString()) // 1.2
    
    // Родитель
    parent := col2.Parent()
    fmt.Println(parent.IndexString()) // 1
}
```

### NullOID

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание NullOID
    nullOID := oid.FromOID(oid.MustParseOID("1.3.6.1"))
    fmt.Println(nullOID.Valid) // true
    fmt.Println(nullOID.OID)   // 1.3.6.1
    
    // NULL
    nullOID2 := oid.NullOID{Valid: false}
    fmt.Println(nullOID2.Valid) // false
    
    // Строковое представление
    fmt.Println(nullOID.String())  // 1.3.6.1
    fmt.Println(nullOID2.String()) // ""
    
    // JSON
    data, _ := nullOID.MarshalJSON()
    fmt.Println(string(data)) // "1.3.6.1"
    
    data2, _ := nullOID2.MarshalJSON()
    fmt.Println(string(data2)) // null
}
```

### Array

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Создание массива
    array := oid.Array{
        oid.MustParseOID("1.3.6.1"),
        oid.MustParseOID("2.100.3"),
    }
    
    // Строковое представление
    fmt.Println(array.String()) // [1.3.6.1, 2.100.3]
    
    // PostgreSQL формат
    value, _ := array.Value()
    fmt.Println(value) // {1.3.6.1,2.100.3}
    
    // Проверка
    fmt.Println(array.Contains(oid.MustParseOID("1.3.6.1"))) // true
    
    // Добавление
    extended := array.Append(oid.MustParseOID("0.39.1"))
    fmt.Println(extended) // [1.3.6.1, 2.100.3, 0.39.1]
    
    // JSON
    jsonData, _ := array.MarshalJSON()
    fmt.Println(string(jsonData)) // ["1.3.6.1","2.100.3"]
}
```

### Комплексный пример с SNMP

```go
package main

import (
    "fmt"
    "github.com/alme23/oid"
)

func main() {
    // Инициализация SNMP реестра
    reg := oid.NewSNMPRegistry()
    
    // Регистрация стандартных OID
    initSNMPRegistry(reg)
    
    // Получение OID для запросов
    sysDescr, _ := reg.GetScalar("sysDescr")
    sysUpTime, _ := reg.GetScalar("sysUpTime")
    
    fmt.Printf("GET %s\n", sysDescr)
    fmt.Printf("GET %s\n", sysUpTime)
    
    // Walk по таблице интерфейсов
    ifDescrOID, _ := reg.GetColumnWithIndexes("ifDescr", 1)
    fmt.Printf("WALK %s\n", ifDescrOID)
    
    // Проверка принадлежности
    base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
    parsed, _ := oid.ParseColumnarOID(base, ifDescrOID)
    fmt.Printf("Column: %d, Indexes: %v\n", parsed.Column, parsed.Indexes)
}

func initSNMPRegistry(reg *oid.SNMPRegistry) {
    // Системные OID
    reg.RegisterScalar("sysDescr", oid.MustScalarOID("1.3.6.1.2.1.1.1.0"))
    reg.RegisterScalar("sysUpTime", oid.MustScalarOID("1.3.6.1.2.1.1.3.0"))
    reg.RegisterScalar("sysName", oid.MustScalarOID("1.3.6.1.2.1.1.5.0"))
    
    // Интерфейсы
    base := oid.MustParseOID("1.3.6.1.2.1.2.2.1")
    reg.RegisterColumn("ifIndex", oid.NewColumnarOID(base, 1))
    reg.RegisterColumn("ifDescr", oid.NewColumnarOID(base, 2))
    reg.RegisterColumn("ifType", oid.NewColumnarOID(base, 3))
}
```
