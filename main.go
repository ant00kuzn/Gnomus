package Gnomus

import (
	"github.com/Tnze/go-mc/net"
	"github.com/ant00kuzn/Gnomus/Gnomus/server"
	"log"
)

func main() {
	// Запуск сокета
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
		go Gnomus.AcceptConnection(connection)
	}
}
