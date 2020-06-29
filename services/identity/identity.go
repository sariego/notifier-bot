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
func (d Driver) Register(username, userID, channelID string) (string, error) {
	// check if in private channel
	if !d.isChannelValid(channelID) {
		return "error :c\n" +
			"sólo puedo registrate en un canal directo,\n" +
			"prueba en una conversación privada conmigo", nil
	}
	// reject empty name
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

// WhoAmI - prints caller user identities
func WhoAmI(id string) (string, error) {
	return formatNames(
		"select username from identity where user_id = $1",
		id,
		"no se quién eres :c\n"+
			"usa !register [username] en una conversacion\n"+
			"privada conmigo para registrarte",
	)
}

// WhoIsHere - prints user identities in current channel
func (d Driver) WhoIsHere(id string) (string, error) {
	info, _ := d.Client.GetChannelInfo(id)

	return formatNames(
		"select distinct on(user_id) username from identity where user_id = any($1)",
		pq.Array(info.Participants),
		"no conozco a nadie acá :c",
	)
}

// scans and format names into a response, fallbacks if empty
func formatNames(query string, target interface{}, fallback string) (string, error) {
	names, _ := scanNames(query, target)
	if len(names) > 0 {
		return strings.Join(names, " "), nil
	}
	return fallback, nil
}

// return list of @ selectors of user or channel
func scanNames(query string, target interface{}) (names []string, err error) {
	rows, err := data.DB.Query(query, target)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		if err != nil {
			log.Println("error@identity_scan: ", err)
			return
		}
		names = append(names, "@"+username)
	}
	return
}

// returns true if channel has only one participant other than the bot
func (d Driver) isChannelValid(id string) bool {
	info, _ := d.Client.GetChannelInfo(id)
	p := info.Participants
	bid := d.Client.BotID()

	return len(p) == 2 &&
		(p[0] == bid || p[1] == bid)
}
