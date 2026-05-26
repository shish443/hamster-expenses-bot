// main_test.go
// тесты написаны ии (гемини) просто потому что до написания тестов я сам еще не добрался, планирую в сл проекте освоить это
package main

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ==========================================
// 1. ТЕСТЫ ДЛЯ ПАРСЕРА (parser.go)
// ==========================================
func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr string
		wantCmd *Command
	}{
		{
			name: "Успешный парсинг команды с аргументами",
			text: "/add 150 корм",
			wantCmd: &Command{
				Name: "add",
				Args: []string{"150", "корм"},
			},
		},
		{
			name: "Команда в верхнем регистре и с лишними пробелами",
			text: "   /Add   300   сено  ",
			wantCmd: &Command{
				Name: "add",
				Args: []string{"300", "сено"},
			},
		},
		{
			name:    "Текст не является командой",
			text:    "привет хомяк",
			wantErr: "не команда",
		},
		{
			name:    "Просто слэш без текста",
			text:    "/",
			wantCmd: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text:      tt.text,
					From:      &tgbotapi.User{ID: 123},
					Chat:      &tgbotapi.Chat{ID: 456},
					MessageID: 789,
				},
			}

			cmd, err := Parser(nil, update)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("Ожидалась ошибка %q, получено: %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Неожиданная ошибка: %v", err)
			}

			if tt.wantCmd == nil {
				if cmd != nil {
					t.Fatalf("Ожидался nil-результат, получено: %+v", cmd)
				}
				return
			}

			if cmd.Name != tt.wantCmd.Name {
				t.Errorf("Получено имя команды %q, ожидалось %q", cmd.Name, tt.wantCmd.Name)
			}

			if len(cmd.Args) != len(tt.wantCmd.Args) {
				t.Errorf("Получено аргументов %d, ожидалось %d", len(cmd.Args), len(tt.wantCmd.Args))
			} else {
				for i := range cmd.Args {
					if cmd.Args[i] != tt.wantCmd.Args[i] {
						t.Errorf("Аргумент [%d]: получено %q, ожидалось %q", i, cmd.Args[i], tt.wantCmd.Args[i])
					}
				}
			}
		})
	}
}

// ==========================================
// 2. ТЕСТЫ ДЛЯ СТАТИСТИКИ (stats.go)
// ==========================================
func TestStatsCounter(t *testing.T) {
	// Фиксируем текущие показатели атомиков
	initMsgs := totalMessages
	initErrs := totalErrors
	initComplaints := totalComplaints

	incrementMessages()
	if totalMessages != initMsgs+1 {
		t.Errorf("incrementMessages не увеличил счетчик. Было %d, стало %d", initMsgs, totalMessages)
	}

	incrementErrors()
	if totalErrors != initErrs+1 {
		t.Errorf("incrementErrors не увеличил счетчик. Было %d, стало %d", initErrs, totalErrors)
	}

	incrementComplaints()
	if totalComplaints != initComplaints+1 {
		t.Errorf("incrementComplaints не увеличил счетчик. Было %d, стало %d", initComplaints, totalComplaints)
	}

	uptime := getUptime()
	if uptime == "" {
		t.Error("getUptime вернул пустую строку")
	}
}

// ==========================================
// 3. ТЕСТЫ ЛОГИКИ НЕ-КОМАНД (botLoop.go)
// ==========================================
func TestNonCommandLogic(t *testing.T) {
	userID := int64(777999)

	// Сбрасываем состояние счетчика для чистоты теста
	resetNonCommandCounter(userID)

	// Должен отвечать (возвращать true) ровно на каждый 5-й раз
	for i := 1; i <= 4; i++ {
		if handleNonCommand(userID) {
			t.Errorf("Итерация %d: ожидалось false (не отвечать), получено true", i)
		}
	}

	if !handleNonCommand(userID) {
		t.Error("Итерация 5: ожидалось true (показать подсказку), получено false")
	}

	// Сбрасываем и проверяем, что счетчик обнулился
	resetNonCommandCounter(userID)
	if handleNonCommand(userID) {
		t.Error("После сброса: первая же не-команда вернула true, а должна false")
	}
}

// ==========================================
// 4. ТЕСТЫ СТАТИЧЕСКИХ ХЕНДЛЕРОВ (handlers.go)
// ==========================================
func TestStaticHandlers(t *testing.T) {
	helpText := HandleHelp()
	if !strings.Contains(helpText, "Доступные команды") {
		t.Error("HandleHelp не содержит базовое приветствие")
	}

	authorText := InfoAuthor()
	if !strings.Contains(authorText, "github.com/shish443") {
		t.Error("InfoAuthor не содержит ссылку на гитхаб автора")
	}

	adminHelpText := HandleAdminHelp()
	if !strings.Contains(adminHelpText, "Админ-панель") {
		t.Error("HandleAdminHelp возвращает невалидный текст")
	}
}

// ==========================================
// 5. КОМПЛЕКСНЫЕ ТЕСТЫ ХЕНДЛЕРОВ С БД (handlers.go)
// ==========================================
func TestHandlersWithMockDB(t *testing.T) {
	// Инициализируем mock БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания mock-соединения с БД: %v", err)
	}
	defer db.Close()

	// Подменяем глобальную переменную DataBase на нашу заглушку
	oldDB := DataBase
	DataBase = db
	defer func() { DataBase = oldDB }() // Возвращаем исходную после теста

	ctx := context.Background()
	userID := int64(555)

	// --- Тест HandleAdd (Валидация) ---
	t.Run("HandleAdd - пустые аргументы", func(t *testing.T) {
		_, err := HandleAdd(ctx, []string{}, userID)
		if err == nil || !strings.Contains(err.Error(), "укажите сумму") {
			t.Errorf("Ожидалась ошибка валидации суммы, получено: %v", err)
		}
	})

	t.Run("HandleAdd - некорректный формат суммы", func(t *testing.T) {
		_, err := HandleAdd(ctx, []string{"abc"}, userID)
		if err == nil || !strings.Contains(err.Error(), "положительным числом") {
			t.Errorf("Ожидалась ошибка формата числа, получено: %v", err)
		}
	})

	t.Run("HandleAdd - сумма превышает лимит", func(t *testing.T) {
		_, err := HandleAdd(ctx, []string{"1500000"}, userID)
		if err == nil || !strings.Contains(err.Error(), "слишком большая сумма") {
			t.Errorf("Ожидалась ошибка превышения лимита, получено: %v", err)
		}
	})

	// --- Тест HandleAdd (Успешная запись) ---
	t.Run("HandleAdd - успешная вставка в БД", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO expenses").
			WithArgs(userID, float64(150), "корм для хомяка").
			WillReturnResult(sqlmock.NewResult(1, 1)) // Имитируем успешный insert 1 строки
		mock.ExpectCommit()

		msg, err := HandleAdd(ctx, []string{"150", "корм", "для", "хомяка"}, userID)
		if err != nil {
			t.Fatalf("Неожиданная ошибка при добавлении: %v", err)
		}
		if !strings.Contains(msg, "150.00") || !strings.Contains(msg, "корм для хомяка") {
			t.Errorf("Ответ хендлера некорректен: %q", msg)
		}
	})

	// --- Тест HandleDelete ---
	t.Run("HandleDelete - запись успешно удалена", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM expenses WHERE id = \\$1 AND user_id = \\$2").
			WithArgs(42, userID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка затронута

		msg, err := HandleDelete(ctx, []string{"42"}, userID)
		if err != nil {
			t.Fatalf("Неожиданная ошибка при удалении: %v", err)
		}
		if !strings.Contains(msg, "удалена") {
			t.Errorf("Ожидалось подтверждение удаления, получено: %q", msg)
		}
	})

	t.Run("HandleDelete - запись не найдена / чужая", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM expenses WHERE id = \\$1 AND user_id = \\$2").
			WithArgs(99, userID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто

		_, err := HandleDelete(ctx, []string{"99"}, userID)
		if err == nil || !strings.Contains(err.Error(), "не найдена или не принадлежит вам") {
			t.Errorf("Ожидалась ошибка отсутствия записи, получено: %v", err)
		}
	})

	// --- Тест HandleComplaint ---
	t.Run("HandleComplaint - успешная отправка жалобы", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO complaints").
			WithArgs(userID, "не работает калькулятор").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Передаем nil вместо бота, так как os.Getenv("ADMIN_ID") не задан
		// функция notifyAdmin завершится на первой же проверке и не вызовет панику.
		msg, err := HandleComplaint(ctx, nil, []string{"не", "работает", "калькулятор"}, userID)
		if err != nil {
			t.Fatalf("Неожиданная ошибка при отправке жалобы: %v", err)
		}
		if !strings.Contains(msg, "Жалоба отправлена админу") {
			t.Errorf("Неверный ответ на жалобу: %q", msg)
		}
	})

	// --- Тест HandleCalc ---
	t.Run("HandleCalc - подсчет за всё время (calcAllTime)", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(450.75)
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE user_id = \\$1").
			WithArgs(userID).
			WillReturnRows(rows)

		msg, err := HandleCalc(ctx, []string{"all"}, userID)
		if err != nil {
			t.Fatalf("Неожиданная ошибка при подсчете: %v", err)
		}
		if !strings.Contains(msg, "450.75") {
			t.Errorf("В ответе нет ожидаемой суммы: %q", msg)
		}
	})

	// Проверяем, что все ожидания mock-БД были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все запланированные SQL-запросы были выполнены в тестах: %v", err)
	}
}
