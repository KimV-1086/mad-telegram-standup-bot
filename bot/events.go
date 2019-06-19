package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/model"
	log "github.com/sirupsen/logrus"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) error {

	message := update.Message

	if message == nil {
		//need to check for edited message content.
		return nil
	}

	if message.IsCommand() {
		return b.HandleCommand(update)
	}

	if message.Text != "" {
		return b.HandleMessageEvent(update)
	}

	if message.LeftChatMember != nil {
		return b.HandleChannelLeftEvent(update)
	}

	if message.NewChatMembers != nil {
		return b.HandleChannelJoinEvent(update)
	}

	//? need to handle user change username
	return nil
}

//HandleMessageEvent function to analyze and save standups
func (b *Bot) HandleMessageEvent(event tgbotapi.Update) error {

	if !strings.Contains(event.Message.Text, b.tgAPI.Self.UserName) {
		return nil
	}

	if !isStandup(event.Message.Text) {
		return fmt.Errorf("Message is not a standup")
	}

	_, err := b.db.CreateStandup(&model.Standup{
		MessageID: event.Message.MessageID,
		Created:   time.Now().UTC(),
		Modified:  time.Now().UTC(),
		Username:  event.Message.From.UserName,
		Text:      event.Message.Text,
		ChatID:    event.Message.Chat.ID,
	})

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(event.Message.Chat.ID, "Спасибо, стендап принят!")
	msg.ReplyToMessageID = event.Message.MessageID
	_, err = b.tgAPI.Send(msg)
	return err
}

//HandleChannelLeftEvent function to remove bot and standupers from channels
func (b *Bot) HandleChannelLeftEvent(event tgbotapi.Update) error {
	member := event.Message.LeftChatMember
	// if user is a bot
	if member.UserName == b.tgAPI.Self.UserName {
		team := b.findTeam(event.Message.Chat.ID)
		if team == nil {
			return fmt.Errorf("Could not find sutable team")
		}
		team.Stop()

		err := b.db.DeleteGroupStandupers(event.Message.Chat.ID)
		if err != nil {
			return err
		}
		err = b.db.DeleteGroup(team.Group.ID)
		if err != nil {
			return err
		}
		return nil
	}

	standuper, err := b.db.FindStanduper(member.UserName, event.Message.Chat.ID)
	if err != nil {
		return nil
	}
	err = b.db.DeleteStanduper(standuper.ID)
	if err != nil {
		return err
	}
	return nil
}

//HandleChannelJoinEvent function to add bot and standupers t0 channels
func (b *Bot) HandleChannelJoinEvent(event tgbotapi.Update) error {
	for _, member := range *event.Message.NewChatMembers {
		// if user is a bot
		if member.UserName == b.tgAPI.Self.UserName {

			_, err := b.db.FindGroup(event.Message.Chat.ID)
			if err != nil {
				log.Info("Could not find the group, creating...")
				group, err := b.db.CreateGroup(&model.Group{
					ChatID:          event.Message.Chat.ID,
					Title:           event.Message.Chat.Title,
					Username:        event.Message.Chat.UserName,
					Description:     event.Message.Chat.Description,
					StandupDeadline: "10:00",
					TZ:              "Asia/Bishkek", // default value...
				})
				if err != nil {
					return err
				}

				b.watchersChan <- group
			}
			// Send greeting message after success group save
			text := "Всем привет! Я буду помогать вам не забывать о сдаче стендапов вовремя. За все мои ошибки отвечает @anatoliyfedorenko :)"
			_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
			return err
		}
		//if it is a regular user, greet with welcoming message and add to standupers
		_, err := b.db.FindStanduper(member.UserName, event.Message.Chat.ID) // user[1:] to remove leading @
		if err == nil {
			return nil
		}

		_, err = b.db.CreateStanduper(&model.Standuper{
			UserID:       member.ID,
			Username:     member.UserName,
			ChatID:       event.Message.Chat.ID,
			LanguageCode: member.LanguageCode,
			TZ:           "Asia/Bishkek", // default value...
		})
		if err != nil {
			log.Error("CreateStanduper failed: ", err)
			return nil
		}

		group, err := b.db.FindGroup(event.Message.Chat.ID)
		if err != nil {
			group, err = b.db.CreateGroup(&model.Group{
				ChatID:          event.Message.Chat.ID,
				Title:           event.Message.Chat.Title,
				Description:     event.Message.Chat.Description,
				StandupDeadline: "10:00",
				TZ:              "Asia/Bishkek", // default value...
			})
			if err != nil {
				return err
			}
		}

		var welcome, onbording, deadline, closing string

		welcome = fmt.Sprintf("Привет, @%v! Добро пожаловать в %v!\n", member.UserName, event.Message.Chat.Title)

		onbording = b.c.OnbordingMessage + "\n"

		if group.StandupDeadline != "" {
			deadline = fmt.Sprintf("Срок сдачи стендапов ежедневно до %s. В выходные пишите стендапы по желанию.\n", group.StandupDeadline)
		}

		closing = "Если по каким-либо серьезным причинам нужно перестать ждать стендапы от вас, сделайте /leave .\n\nЗа все мои ошибки отвечает @anatolifedorenko"

		text := welcome + onbording + deadline + closing

		_, err = b.tgAPI.Send(tgbotapi.NewMessage(event.Message.Chat.ID, text))
		return err
	}
	return nil
}
