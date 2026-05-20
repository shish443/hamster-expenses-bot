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
	//если енв есть-пропускаем
	if _, err := os.Stat(".env"); err == nil {
		fmt.Println("Файл .env уже существует. Пропускаем настройку переменных.")
	} else if os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)
		//запрашиваем данные от пользователя
		fmt.Print("📝 Введите имя для базы данных: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("📝 Введите токен Telegram-бота: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		fmt.Print("📝 Введите имя пользователя для базы данных: ")
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)

		var pwd string
		for {
			fmt.Print("📝 Введите пароль для базы данных: ")
			pwd1, _ := reader.ReadString('\n')
			pwd1 = strings.TrimSpace(pwd1)

			fmt.Print("📝 Введите пароль повторно: ")
			pwd2, _ := reader.ReadString('\n')
			pwd2 = strings.TrimSpace(pwd2)

			if pwd1 == pwd2 {
				pwd = pwd1
				break
			}
			fmt.Println("Пароли не совпадают! Пожалуйста, повторите ввод.")
			fmt.Println()
		}
		//записываем данные в енв файл
		envContent := fmt.Sprintf(
			"DB_NAME=%s\nApiTelegramBot=%s\nDB_USER=%s\nDB_PASSWORD=%s\n",
			name, token, user, pwd,
		)

		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			fmt.Printf("Ошибка записи .env: %v\n", err)
			return
		}
		fmt.Println("Файл .env успешно создан!")
	}
	// хопа меджик и докер запущен
	fmt.Println("\nЗапускаем все через Docker Compose")
	dockerCmd := exec.Command("docker", "compose", "up", "--build", "-d")
	dockerCmd.Stdout = os.Stdout //это гарри магия, что бы юзер видел что происходит
	dockerCmd.Stderr = os.Stderr
	if err := dockerCmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска Docker: %v\n", err)
		return
	}
	fmt.Println("\nВсе работае")
	fmt.Println("Чтобы посмотреть, что происходит внутри бота, выполни команду:")
	fmt.Println("docker logs -f hamster_bot")
}
