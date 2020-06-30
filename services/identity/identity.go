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
			"solo puedo registrate en un canal directo,\n" +
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
			log.Println("error@register:", err)
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
			"usa !register [username] en una conversación\n"+
			"privada conmigo para registrarte",
	)
}

// WhoIsHere - prints user identities in current channel
func (d Driver) WhoIsHere(id string) (string, error) {
	info, _ := d.Client.GetChannelInfo(id)

	return formatNames(
		"select distinct on(user_id) username from identity "+
			"where user_id = any($1) order by created",
		pq.Array(info.Participants),
		"no conozco a nadie acá :c",
	)
}

// NotifyMentions - notifies @ mentions in channels where user is in
func (d Driver) NotifyMentions(pkg base.Package) error {
	info, _ := d.Client.GetChannelInfo(pkg.Channel)
	channels, err := d.GetNotifyChannels(pkg)
	if err != nil {
		return err
	}

	for _, ch := range channels {
		log.Printf("exec: notify@%v\n", ch)

		// generate notification message
		sender := GetSenderName(pkg.Author)
		summary := base.Package{
			Channel: ch,
			Message: fmt.Sprintf(
				"%v te ha etiquetado en %v\n%v",
				sender,
				info.Name,
				fmt.Sprintf(d.Client.MentionsRedirectURL(), pkg.Channel),
			),
		}
		message := base.Package{
			Channel: ch,
			Message: fmt.Sprintf(
				"el mensaje fué:\n%v",
				pkg.Message,
			),
		}
		d.Client.Send(summary)
		// time.Sleep(500 * time.Millisecond)
		d.Client.Send(message)
	}

	return nil
}

// GetNotifyChannels - returns channels to notify
func (d Driver) GetNotifyChannels(pkg base.Package) (channels []string, err error) {
	info, _ := d.Client.GetChannelInfo(pkg.Channel)
	names, _ := scanNames(
		"select username from identity where user_id = any($1)",
		pq.Array(info.Participants),
	)
	args := strings.Split(pkg.Message, " ")
	// fmt.Println("names:", names)
	// fmt.Println("args:", args)

	rows, err := data.DB.Query(
		"select distinct channel_id from identity where '@'||username = any($1) and '@'||username = any($2)",
		pq.Array(names), pq.Array(args),
	)
	if err != nil {
		log.Println("error@notify_get_channels:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch string
		err := rows.Scan(&ch)
		if err != nil {
			log.Println("error@notify_each_channel:", err)
			continue
		}
		channels = append(channels, ch)
	}
	return
}

// GetSenderName - returns username if known, fallback to 'alguien'
func GetSenderName(id string) (sender string) {
	_ = data.DB.QueryRow(
		"select username from identity where user_id = $1",
		id,
	).Scan(&sender)
	if len(sender) == 0 {
		sender = "alguien"
	} else {
		sender = "@" + sender
	}
	return
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
			log.Println("error@identity_scan:", err)
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
