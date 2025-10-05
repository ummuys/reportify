package service

import (
	"fmt"
	"testing"
)

func TestReportService(t *testing.T) {
	mockService := newMockReportService(nil, nil, nil)
	mockService.CreateReport(t.Context(), "", nil)
	mockService.GetSchemas(t.Context())
	mockService.GetColumns(t.Context(), "", "")
	mockService.GetTables(t.Context(), "")
	fmt.Println("Сделаю когда-нибудь тесты")
}

//Правильно ли я понимаю moсk тестиование: я написал интерфейсы (service, handlers, db, cache, convert). Service дергает db, convert и cache, поэтому его нужно тестировать. Для него нужно сделать mock структуру, а для db/convert/cache можно не мокать, просто подключать их в подготовленную базу для тестирования
