//go:build ignore

// start.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("Накатываем окружение для Хамстер-Бота")

	// === Создание .env ===
	if _, err := os.Stat(".env"); err == nil {
		fmt.Println("файл .env уже существует. Пропускаем настройку переменных.")
	} else if os.IsNotExist(err) {
		if err := setupEnv(); err != nil {
			fmt.Printf("Ошибка настройки окружения: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Ошибка проверки .env: %v\n", err)
		os.Exit(1)
	}

	//запуск дкера
	fmt.Println("\nЗапускаем все контейнеры через docker-compose.yml")
	dockerCmd := exec.Command("docker", "compose", "up", "--build", "-d")
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	if err := dockerCmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска Docker Compose: %v\n", err)
		fmt.Println("Попробуйте запустить вручную: docker compose up --build -d")
		os.Exit(1)
	}

	fmt.Println("\nБот успешно запущен в фоне!")
	fmt.Println("\nПолезные команды:")
	fmt.Println("   • Посмотреть логи бота:       docker logs -f hamster_bot")
	fmt.Println("   • Перезапустить:             docker compose restart")
	fmt.Println("   • Остановить:                docker compose down")
	fmt.Println("   • Посмотреть все контейнеры: docker compose ps")

	fmt.Println("\nВозвращаемся в терминал.")
}

func setupEnv() error {
	reader := bufio.NewReader(os.Stdin)

	//запрашиваем данные от пользователя
	fmt.Print("📝 Введите имя базы данных (например hamster_db): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = "hamster_db"
		fmt.Println("вы ничего не ввели. Применено имя по умолчанию: hamster_db")
	}

	fmt.Print("📝 Введите токен Telegram-бота (от @BotFather): ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("токен бота обязателен")
	}

	fmt.Print("📝 Введите ваш Telegram ID (для админ-панели): ")
	adminStr, _ := reader.ReadString('\n')
	adminStr = strings.TrimSpace(adminStr)
	if adminStr == "" {
		adminStr = "0"
	}

	fmt.Print("📝 Введите имя пользователя для базы данных:")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	if user == "" {
		user = "hamster"
		fmt.Println("вы ничего не ввели. Применено имя по умолчанию: hamster")

	}

	// Пароль с подтверждением
	var pwd string
	for {
		fmt.Print("📝 Введите пароль для базы данных: ")
		pwd1, _ := reader.ReadString('\n')
		pwd1 = strings.TrimSpace(pwd1)

		fmt.Print("📝 Повторите пароль: ")
		pwd2, _ := reader.ReadString('\n')
		pwd2 = strings.TrimSpace(pwd2)

		if pwd1 == pwd2 {
			pwd = pwd1
			break
		}
		fmt.Println("❌ Пароли не совпадают! Попробуйте ещё раз.")
	}

	//какие времена-такие и решения
	fmt.Print("\n ваш сервер в РФ / нужен обход блокировок Telegram? (y/n): ")
	proxyChoice, _ := reader.ReadString('\n')
	proxyChoice = strings.ToLower(strings.TrimSpace(proxyChoice))

	var proxyEnvLine string
	if strings.HasPrefix(proxyChoice, "y") || proxyChoice == "да" || proxyChoice == "д" {
		proxyEnvLine = setupProxy(reader)
	} else {
		_ = os.WriteFile("sing_box_config.json", []byte(`{"log":{"level":"info"},"outbounds":[{"type":"direct","tag":"direct"}]}`), 0644)
		proxyEnvLine = ""
	}

	//записываем данные в енв файл
	envContent := fmt.Sprintf(
		"DB_NAME=%s\nApiTelegramBot=%s\nDB_USER=%s\nDB_PASSWORD=%s\nADMIN_ID=%s\n%s",
		name, token, user, pwd, adminStr, proxyEnvLine,
	)

	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		return fmt.Errorf("не удалось записать .env: %w", err)
	}

	fmt.Println("✅ Файл .env успешно создан!")
	return nil
}

func setupProxy(reader *bufio.Reader) string {
	fmt.Println("\n📋 Вставьте ПОЛНЫЙ JSON конфиг sing-box (или Enter для минимального и отредактируйте его вручную):")
	fmt.Println("(вставьте и нажмите Enter два раза)")

	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" && len(lines) > 0 {
			break
		}
		if line != "" {
			lines = append(lines, line)
		}
	}

	configStr := strings.Join(lines, "\n")
	if configStr == "" {
		fmt.Println("Создаём минимальный конфиг")
		configStr = `{"log":{"level":"info"},"outbounds":[{"type":"direct","tag":"direct"}]}`
	}

	if err := os.WriteFile("sing_box_config.json", []byte(configStr), 0644); err != nil {
		fmt.Println("Не удалось сохранить конфиг, используется минимальный")
		configStr = `{"log":{"level":"info"},"outbounds":[{"type":"direct","tag":"direct"}]}`
		_ = os.WriteFile("sing_box_config.json", []byte(configStr), 0644)
	}

	return "HTTPS_PROXY=http://127.0.0.1:2080\n"
}
