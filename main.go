package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/scukonick/teachbot/app"
	"github.com/scukonick/teachbot/cfg"
	"github.com/scukonick/teachbot/db"
	log "github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

func cmdRun(c *cli.Context) {
	log.SetLevel(log.DebugLevel)

	config := cfg.GetConfig()

	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	bot.Debug = true
	if err != nil {
		log.WithError(err).Fatal("Failed to create bot API")
	}

	dbConn, err := gorm.Open("postgres", config.DBConn)
	if err != nil {
		log.WithError(err).WithField("DBConn", config.DBConn).
			Fatal("Failed to open db connection")
	}

	storage := db.NewStorage(dbConn)

	appInstance := app.NewServer(storage, bot)

	err = appInstance.Run()
	if err != nil {
		log.WithError(err).Fatal("Failed to run app instance")
	}
}

func main() {
	a := cli.NewApp()

	a.Version = "0.0.1"

	a.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Run backend process",
			Action: cmdRun,
		},
	}

	a.Run(os.Args)
}
