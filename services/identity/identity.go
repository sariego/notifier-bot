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
	if !d.Client.IsValidManagementChannel(channelID) {
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
	return data.FormatTerms(
		"select username from identity where user_id = $1 order by created",
		id,
		"@",
		"no se quién eres :c\n"+
			"usa !register [username] en una conversación\n"+
			"privada conmigo para registrarte",
	)
}

// WhoIsHere - prints user identities in current channel
func (d Driver) WhoIsHere(id string) (string, error) {
	info, _ := d.Client.GetChannelInfo(id)

	return data.FormatTerms(
		"select distinct on(user_id) username from identity "+
			"where user_id = any($1) order by user_id, created",
		pq.Array(info.Participants),
		"@",
		"no conozco a nadie acá :c",
	)
}

// NotifyMentions - notifies @ mentions in channels where user is in
func (d Driver) NotifyMentions(pkg base.Package) error {
	info, _ := d.Client.GetChannelInfo(pkg.Channel)

	// do NOT use GetNotifyChannels, we want participants only
	channels, err := d.getNotifyChannelsForParticipantsOnly(
		pkg.Channel,
		pkg.Message,
	)
	if err != nil {
		return err
	}

	for _, ch := range channels {
		log.Printf("exec: notify_mention@%v\n", ch)

		// generate notification message
		sender := GetSenderName(pkg.Author)
		out := base.Package{
			Channel: ch,
			Message: fmt.Sprintf(
				"%v te ha etiquetado en %v\n\"%v\"\n\n%v",
				sender,
				info.Name,
				pkg.Message,
				fmt.Sprintf(d.Client.ChannelURLTemplate(), pkg.Channel),
			),
		}
		d.Client.Send(out)
	}

	return nil
}

// GetNotifyChannels - returns channels to notify from @ mentions in args
func GetNotifyChannels(args []string) (channels []string, err error) {
	rows, err := data.DB.Query(
		"select distinct channel_id from identity where '@'||username = any($1)",
		pq.Array(args),
	)
	if err != nil {
		log.Println("error@get_notify_channels_all:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch string
		err := rows.Scan(&ch)
		if err != nil {
			log.Println("error@get_notify_channels_each:", err)
			continue
		}
		channels = append(channels, ch)
	}
	return
}

// GetSenderName - returns username if known, fallback to 'alguien'
func GetSenderName(id string) (sender string) {
	_ = data.DB.QueryRow(
		"select username from identity where user_id = $1 order by created",
		id,
	).Scan(&sender)
	if len(sender) == 0 {
		sender = "alguien"
	} else {
		sender = "@" + sender
	}
	return
}

// similar to GetNotifyChannels but only for users in channel ch,
// also matches symbols after username
func (d Driver) getNotifyChannelsForParticipantsOnly(ch, msg string) (channels []string, err error) {
	info, _ := d.Client.GetChannelInfo(ch)
	names, _ := data.ScanTerms(
		"select username from identity where user_id = any($1)",
		pq.Array(info.Participants), "@",
	)
	args := strings.Split(msg, " ")
	// fmt.Println("names:", names)
	// fmt.Println("args:", args)

	rows, err := data.DB.Query(
		"select distinct channel_id from identity where '@'||username in("+
			"select name from unnest($1::varchar[]) as u(name), "+
			"unnest($2::varchar[]) as a(word) "+
			"where word ~* ('^'||name||'[^\\w\\s]*$'))",
		pq.Array(names), pq.Array(args),
	)
	if err != nil {
		log.Println("error@get_notify_participants_all:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch string
		err := rows.Scan(&ch)
		if err != nil {
			log.Println("error@get_notify_participants_each:", err)
			continue
		}
		channels = append(channels, ch)
	}
	return
}
