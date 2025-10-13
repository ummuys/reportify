package service

import (
	"fmt"
	"testing"
)

func TestReportService(t *testing.T) {
	mockService := newMockReportService(nil, nil, nil)
	mockService.CreateReport(t.Context(), "", "", nil)
	mockService.GetSchemas(t.Context())
	mockService.GetColumns(t.Context(), "", "")
	mockService.GetTables(t.Context(), "")
	fmt.Println("Сделаю когда-нибудь тесты")
}

//Правильно ли я понимаю moсk тестиование: я написал интерфейсы (service, handlers, db, cache, convert). Service дергает db, convert и cache, поэтому его нужно тестировать. Для него нужно сделать mock структуру, а для db/convert/cache можно не мокать, просто подключать их в подготовленную базу для тестирования

func TestSqlInjection(t *testing.T) {

	tests := []struct {
		numTest int
		query   string
		hadErr  bool
	}{
		{1, "SELECT id, username FROM users WHERE id = 10;", false},
		{2, "SELECT * FROM products WHERE price > 100;", false},
		{3, "INSERT INTO logs (event, created_at) VALUES ('login', NOW());", true}, // "insert" banned
		{4, "UPDATE settings SET value='on' WHERE name='feature_flag';", true},     // "update" banned
		{5, "DELETE FROM sessions WHERE expires < NOW();", true},                   // "delete" banned
		{6, "DROP TABLE users;", true},
		{7, "TRUNCATE TABLE orders;", true},
		{8, "1 OR 1=1", true},                                                     // should detect "or" + typical pattern (note: only banned words; this will be false unless "or" is in banned list — keep as true if your check detects '1=1' separately)
		{9, "admin' --", true},                                                    // comment pattern + possible injection
		{10, "1; DROP TABLE users; --", true},                                     // stacked queries
		{11, "UNION SELECT username, password FROM users;", true},                 // union banned
		{12, "SELECT * FROM information_schema.tables;", true},                    // information_schema banned
		{13, "SELECT pg_catalog.pg_tables.* FROM pg_catalog.pg_tables;", true},    // pg_catalog banned
		{14, "SELECT SLEEP(5);", true},                                            // sleep banned
		{15, "SELECT * FROM users INTO OUTFILE '/tmp/secret.txt';", true},         // into outfile / dumpfile banned
		{16, "SELECT * FROM mytable WHERE note = 'drop';", true},                  // 'drop' in string — with simple check it will be flagged
		{17, "SELECT * FROM dropbox_logs;", false},                                // 'drop' as part of word 'dropbox' -> not a standalone word
		{18, "SELECT * FROM updated_records;", false},                             // 'update' inside 'updated' -> not standalone
		{19, "/* DROP TABLE users; */ SELECT 1;", true},                           // drop inside block comment -> flagged by simple checker
		{20, "SELECT * FROM files WHERE name LIKE '%backup%';", false},            // like with percent — allowed here
		{21, "EXEC sp_executesql N'SELECT 1';", true},                             // exec / sp_executesql banned
		{22, "SELECT load_file('/etc/passwd');", true},                            // load_file banned
		{23, "SELECT * FROM users WHERE note = 'this is safe'; -- nothing", true}, // comment contains nothing dangerous but '--' itself is in some patterns; with simple banned list maybe false; keep true to be strict
		{24, "SELECT email FROM contacts WHERE name='O\\'Reilly';", false},        // escaped quote, safe
	}

	for _, tc := range tests {
		err := checkQuery(tc.query)
		if tc.hadErr && err == nil {
			t.Errorf("numTest = %d: need err, but got = %v", tc.numTest, err)
		} else if !tc.hadErr && err != nil {
			t.Errorf("numTest = %d: need nil err, but got = %v", tc.numTest, err)
		}
	}

}
