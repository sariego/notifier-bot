package feedback

import (
	"log"

	"sariego.dev/notifier-bot/services/data"
)

// Create - creates new feedback
func Create(userID, channelID, tag, content string) (string, error) {
	_, err := data.DB.Exec(
		"insert into feedback(user_id,channel_id,tag,content) values($1,$2,$3,$4)",
		userID, channelID, tag, content,
	)
	if err != nil {
		log.Println("error@create_feedback: ", err)
	}

	return "recibido c:\ngracias por tus comentarios!", err
}
