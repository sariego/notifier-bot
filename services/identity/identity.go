package identity

import (
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/services/data"
)

// Driver [Identity] - manages user identities
type Driver struct {
	Client base.Client
}

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

// Deregister - removes identity from registry
func Deregister(username, userID string) (string, error) {
	result, err := data.DB.Exec(
		"delete from identity where username = $1 AND user_id = $2",
		username, userID,
	)
	if n, _ := result.RowsAffected(); err != nil || n == 0 {
		return "", err
	}
	return fmt.Sprintf(
		"listo @%v, no te enviaré mas mensajes",
		username,
	), nil
}

// WhoAmI - prints user identities
func WhoAmI(id string) (string, error) {
	var names []string
	rows, err := data.DB.Query(
		"select username from identity where user_id = $1",
		id,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			log.Println("error@identity_scan: ", err)
			return "", err
		}
		names = append(names, "@"+username)
	}

	if len(names) > 0 {
		return strings.Join(names, " "), nil
	}
	return "no se quién eres :c\n" +
		"usa !register [username] en una conversacion\n" +
		"privada conmigo para registrarte", nil
// returns true if channel has only one participant other than the bot
func (d Driver) isChannelValid(id string) bool {
	info, _ := d.Client.GetChannelInfo(id)
	p := info.Participants
	bid := d.Client.BotID()

	return len(p) == 2 &&
		(p[0] == bid || p[1] == bid)
}
