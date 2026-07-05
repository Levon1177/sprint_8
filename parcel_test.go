package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// Инициализация БД для тестов (обычно уже есть в каркасе)
func InitTestStore(t *testing.T) ParcelStore {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close()
		require.NoError(t, err)
	})

	return NewParcelStore(db)
}

// TestAddGetByClient проверяет добавление и получение посылки
func TestAddGetByClient(t *testing.T) {
	store := InitTestStore(t)

	// Данные тестовой посылки
	parcel := Parcel{
		Client:    1001,
		Status:    ParcelStatusRegistered,
		Address:   "Тестовый адрес",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 1. Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	parcel.Number = id

	// 2. Получаем добавленную посылку по идентификатору
	stored, err := store.Get(id)
	require.NoError(t, err)

	// 3. Проверяем, что данные совпадают
	assert.Equal(t, parcel.Client, stored.Client)
	assert.Equal(t, parcel.Status, stored.Status)
	assert.Equal(t, parcel.Address, stored.Address)
	// Даты можно сравнить строками или через встроенные форматы
	assert.Equal(t, parcel.CreatedAt, stored.CreatedAt)
}

// TestStateChange проверяет цепочку изменений статуса
func TestStateChange(t *testing.T) {
	store := InitTestStore(t)

	parcel := Parcel{
		Client:    1002,
		Status:    ParcelStatusRegistered,
		Address:   "Адрес для статусов",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Меняем статус на sent
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Проверяем, что статус изменился в БД
	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, p.Status)
}

// TestSetAddress проверяет изменение адреса
func TestSetAddress(t *testing.T) {
	store := InitTestStore(t)

	parcel := Parcel{
		Client:    1003,
		Status:    ParcelStatusRegistered,
		Address:   "Старый адрес",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Меняем адрес
	newAddress := "Новый адрес"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Проверяем изменения
	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, p.Address)
}

// TestDelete проверяет удаление посылки
func TestDelete(t *testing.T) {
	store := InitTestStore(t)

	parcel := Parcel{
		Client:    1004,
		Status:    ParcelStatusRegistered,
		Address:   "Адрес под удаление",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Удаляем
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверяем, что при попытке получить выдает ошибку sql.ErrNoRows
	_, err = store.Get(id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
