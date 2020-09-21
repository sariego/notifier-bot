package topics

import (
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"sariego.dev/notifier-bot/base"
	"sariego.dev/notifier-bot/services/data"
	"sariego.dev/notifier-bot/services/identity"
)

// Driver [Topics] - manages user subscriptions
type Driver struct {
	Client base.Client
}

// Subscribe - add new sub to user
func (d Driver) Subscribe(topic, userID, channelID string) (string, error) {
	// check if in private channel
	if !d.Client.IsValidManagementChannel(channelID) {
		return "error :c\n" +
			"solo puedo suscribirte en un canal directo,\n" +
			"prueba en una conversación privada conmigo", nil
	}
	// reject empty topic
	if len(topic) > 0 {
		// clean topic of initial # to avoid confusion
		for topic[0] == '#' {
			topic = topic[1:]
		}
		_, err := data.DB.Exec(
			"insert into subscription(topic,user_id,channel_id) values($1,$2,$3)",
			topic, userID, channelID,
		)
		if err != nil {
			log.Println("error@subscribe", err)
			// let user know if already subscribed
			if err, ok := err.(*pq.Error); ok &&
				err.Code.Name() == "unique_violation" {
				return "ya estas suscrito ;D", nil
			}
			return "", err
		}

		return fmt.Sprintf(
			"listo, usaré este canal para notificarte sobre #%v",
			topic,
		), nil
	}

	return "debes ingresar un tópico", nil
}

// Unsubscribe - removes sub from user
func Unsubscribe(topic, userID string) (string, error) {
	result, err := data.DB.Exec(
		"delete from subscription where topic = $1 AND user_id = $2",
		topic, userID,
	)
	if n, _ := result.RowsAffected(); err != nil || n == 0 {
		return "", err
	}
	return fmt.Sprintf(
		"listo, no te notificaré más sobre #%v",
		topic,
	), nil
}

// Subscriptions - prints caller user subscriptions
func Subscriptions(id string) (string, error) {
	return data.FormatTerms(
		"select topic from subscription where user_id = $1 order by created",
		id,
		"#",
		"no tienes suscripciones\n"+
			"usa !subscribe [topic] en una conversación\n"+
			"privada conmigo para suscribirte",
	)
}

// NotifySubscriptions - notifies # tags in channels where user is in
func (d Driver) NotifySubscriptions(pkg base.Package) error {
	info, _ := d.Client.GetChannelInfo(pkg.Channel)
	channels, err := d.getNotifyChannelsForParticipantsOnly(
		pkg.Author,
		pkg.Channel,
		pkg.Message,
	)
	if err != nil {
		return err
	}

	for _, ch := range channels {
		log.Printf("exec: notify_subs@%v\n", ch)

		// generate notification message
		sender := identity.GetSenderName(pkg.Author)
		out := base.Package{
			Channel: ch,
			Message: fmt.Sprintf(
				"%v ha mencionado una de tus suscripciones en %v\n\"%v\"\n\n%v",
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

// returns channels to notify users in current channel from subs in args
func (d Driver) getNotifyChannelsForParticipantsOnly(author, ch, msg string) (channels []string, err error) {
	info, _ := d.Client.GetChannelInfo(ch)
	topics, _ := data.ScanTerms(
		"select distinct topic from subscription where user_id = any($1)",
		pq.Array(info.Participants), "#",
	)
	args := strings.Split(msg, " ")

	rows, err := data.DB.Query(
		"select distinct channel_id from subscription "+
			"where user_id != $1 and user_id = any($2) and '#'||topic in("+
			"select topic from unnest($3::varchar[]) as u(topic), "+
			"unnest($4::varchar[]) as a(word) "+
			"where word ~* ('^'||topic||'[^\\w\\s]*$'))",
		author,
		pq.Array(info.Participants),
		pq.Array(topics),
		pq.Array(args),
	)
	if err != nil {
		log.Println("error@get_notify_subs_all:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch string
		err := rows.Scan(&ch)
		if err != nil {
			log.Println("error@get_notify_subs_each:", err)
			continue
		}
		channels = append(channels, ch)
	}
	return
}
