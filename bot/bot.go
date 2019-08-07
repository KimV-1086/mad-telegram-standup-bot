package bot

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maddevsio/mad-internship-bot/config"
	"github.com/maddevsio/mad-internship-bot/model"
	"github.com/maddevsio/mad-internship-bot/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

const (
	telegramAPIUpdateInterval = 60
)

// Bot structure
type Bot struct {
	bundle       *i18n.Bundle
	c            *config.BotConfig
	tgAPI        *tgbotapi.BotAPI
	updates      tgbotapi.UpdatesChannel
	db           *storage.MySQL
	watchersChan chan *model.Group
	teams        []*model.Team
	wg           sync.WaitGroup
}

var yesterdayWorkKeywords = []string{"yesterday", "friday", "вчера", "пятниц"}
var todayPlansKeywords = []string{"today", "сегодня"}
var issuesKeywords = []string{"blocker", "problem", "issue", "мешает", "проблем", "сложност", "блок"}

// New creates a new bot instance
func New(c *config.BotConfig, bundle *i18n.Bundle) (*Bot, error) {
	newBot, err := tgbotapi.NewBotAPI(c.TelegramToken)
	if err != nil {
		return nil, err
	}

	log.Info("authorized with ", newBot.Self.UserName)

	u := tgbotapi.NewUpdate(0)

	u.Timeout = telegramAPIUpdateInterval

	updates, err := newBot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}

	wch := make(chan *model.Group)
	var teams []*model.Team

	b := &Bot{
		c:            c,
		tgAPI:        newBot,
		updates:      updates,
		db:           conn,
		watchersChan: wch,
		teams:        teams,
		bundle:       bundle,
	}

	return b, nil
}

// Start bot
func (b *Bot) Start() error {
	b.wg.Add(1)
	go b.StartWatchers()

	groups, err := b.db.ListGroups()
	if err != nil {
		return err
	}

	for _, g := range groups {
		b.watchersChan <- g
	}

	log.Info("Listening for updates... \n")
	for update := range b.updates {
		_, err := b.handleUpdate(update)
		if err != nil {
			log.Error(err)
		}
	}

	b.wg.Wait()

	return nil
}

func (b *Bot) findTeam(chatID int64) *model.Team {
	for _, team := range b.teams {
		if team.Group.ChatID == chatID {
			return team
		}
	}
	return nil
}

func (b *Bot) removeTeam(chatID int64) bool {
	result := false
	for i, team := range b.teams {
		if team.Group.ChatID == chatID {
			b.teams = append(b.teams[:i], b.teams[i+1:]...)
			result = true
		}
	}
	return result
}
