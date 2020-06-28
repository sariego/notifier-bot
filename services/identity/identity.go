package identity

import (
	"fmt"
	"log"

	"github.com/lib/pq"
	"sariego.dev/cotalker-bot/services/data"
)

// Register - add new identity to registry
func Register(username, userID, channelID string) (string, error) {
	// todo check is private channel
	if len(username) > 0 {
		// clean name of initial @ to avoid confusion
		for username[0] == '@' {
			username = username[1:]
		}
		_, err := data.DB.Exec(
			"insert into identity(username,user_id,channel_id) values($1,$2,$3)",
			username, userID, channelID,
		)
		if err != nil {
			log.Println("error@register: ", err)
			// let user know if name is already registered
			if err, ok := err.(*pq.Error); ok &&
				err.Code.Name() == "unique_violation" {
				return "lo siento, este nombre ya está registrado", nil
			}
			return "", err
		}

		return fmt.Sprintf(
			"hola @%v, usaré este canal para notificarte",
			username,
		), nil
	}

	return "debes ingresar un nombre de usuario", nil
}

