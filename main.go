package Gnomus

import (
	"github.com/Tnze/go-mc/net"
	"log"
)

// InitSRV - Функция запуска сервера
func InitSRV() {
	// Запускаем сокет по адрессу 0.0.0.0:25565
	loop, err := net.ListenMC(":25565")
	// Если есть ошибка, то выводим её
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	// Цикл обрабатывающий входящие подключеня
	for {
		// Принимаем подключение или ждём
		connection, err := loop.Accept()
		// Если произошла ошибка - пропускаем соденение
		if err != nil {
			continue
		}
		// Принимаем подключение и обрабатываем его не блокируя основной поток
		go acceptConnection(connection)
	}
}
