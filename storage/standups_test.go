package storage

import (
	"testing"
	"time"

	"github.com/maddevsio/mad-telegram-standup-bot/config"
	"github.com/maddevsio/mad-telegram-standup-bot/model"
	"github.com/stretchr/testify/require"
)

func TestStandup(t *testing.T) {
	conf, err := config.Get()
	require.NoError(t, err)
	mysql, err := NewMySQL(conf)
	require.NoError(t, err)

	s := &model.Standup{
		MessageID: 12345,
		Created:   time.Date(2019, 7, 10, 0, 0, 0, 0, time.Local),
		Modified:  time.Date(2019, 7, 10, 1, 2, 0, 0, time.Local),
		Username:  "User1",
		Text:      "SomeText",
		ChatID:    int64(3),
	}

	standup, err := mysql.CreateStandup(s)
	require.NoError(t, err)
	require.Equal(t, int64(3), standup.ChatID)
	require.Equal(t, "User1", standup.Username)
	require.Equal(t, "SomeText", standup.Text)

	standup.Text = "NewText"

	u, err := mysql.UpdateStandup(standup)
	require.NoError(t, err)
	require.Equal(t, "NewText", u.Text)

	ch, err := mysql.SelectStandup(standup.ID)
	require.NoError(t, err)
	require.Equal(t, standup.ID, ch.ID)

	standup, err = mysql.SelectStandupByMessageID(standup.MessageID, standup.ChatID)
	require.NoError(t, err)
	require.Equal(t, 12345, standup.MessageID)
	require.Equal(t, int64(3), standup.ChatID)

	ch, err = mysql.LastStandupFor(standup.Username, standup.ChatID)
	require.NoError(t, err)
	require.Equal(t, standup.ID, ch.ID)

	standup2, err := mysql.CreateStandup(s)
	require.NoError(t, err)

	standups, err := mysql.ListStandups()
	require.NoError(t, err)
	require.Equal(t, 2, len(standups))

	err = mysql.DeleteStandup(standup2.ID)
	require.NoError(t, err)

	err = mysql.DeleteStandup(standup.ID)
	require.NoError(t, err)

	standups, err = mysql.ListStandups()
	require.NoError(t, err)
	require.Equal(t, 0, len(standups))
}
