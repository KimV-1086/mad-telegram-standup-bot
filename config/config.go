package config

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken string `envconfig:"TELEGRAM_TOKEN" required:"false"`
	DatabaseURL   string `envconfig:"DATABASE_URL" required:"false" default:"telegram:telegram@tcp(localhost:3306)/telegram?parseTime=true"`
}

// Get config data from environment
func Get() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("", &c)
	return &c, err
}


// package config

// //import "github.com/kelseyhightower/envconfig"

// // BotConfig ...
// type BotConfig struct {
//     TelegramToken string //envconfig:"TELEGRAM_TOKEN" required:"false"
//     DatabaseURL   string //envconfig:"DATABASE_URL" required:"false" default:"telegram:telegram@tcp(localhost:3306)/telegram?parseTime=true"
//     Debug         bool   //envconfig:"DEBUG" default:"false"
// }// Get config data from environment
// func Get() (*BotConfig, error) {
//     var c = BotConfig{
//         TelegramToken: "709570732:AAFIbhYyuI9--Hnt_QCacZrnCkz_Z4N8Pf0",
//         DatabaseURL:   "telegram:telegram@tcp(localhost:3306)/telegram?parseTime=true",
//         Debug:         true,
//     }
//     //    err := envconfig.Process("", &c)
//     var err error = nil
//     return &c, err
// }