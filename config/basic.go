package config

import "github.com/Tnze/go-mc/chat"

var (
	ProtocolVersion uint16       = 340
	MOTD            chat.Message = chat.Text("§bGnomus is the best server kernel")
	ADDRESS         string       = "127.0.0.1"
	ADDRESS_PORT    uint16       = 25565
)
