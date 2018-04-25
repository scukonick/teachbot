package app

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/scukonick/teachbot/db"
	"github.com/scukonick/teachbot/db/structs"
	"github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

type UpdateProcessor func(c *UpdateContext)

type UpdateContext struct {
	update *tgbotapi.Update
	stop   bool
	bot    *tgbotapi.BotAPI
}

func (c *UpdateContext) Stop() {
	c.stop = true
}

type Server struct {
	storage    *db.Storage
	bot        *tgbotapi.BotAPI
	processors []UpdateProcessor
}

func NewServer(storage *db.Storage, bot *tgbotapi.BotAPI) *Server {
	return &Server{
		storage: storage,
		bot:     bot,
	}
}

func (s *Server) AddProcessor(p UpdateProcessor) {
	s.processors = append(s.processors, p)
}

func (s *Server) Run() error {
	logctx := logrus.WithField("action", "Run")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	s.AddProcessor(s.updateUserInfo)
	s.AddProcessor(s.processStartMessage)
	s.AddProcessor(s.processInfoMessage)
	s.AddProcessor(s.processImageTaskMessage)
	s.AddProcessor(s.processTextTaskMessage)

	s.AddProcessor(s.processDefaultMessage)

	updates, err := s.bot.GetUpdatesChan(u)
	if err != nil {
		logctx.WithError(err).Error("Failed to get updates chan")
		return err
	}

	logctx.Info("AppServer started")
	for update := range updates {
		s.ProcessMessage(&update)
	}

	return nil
}

// ProcessMessage routes the message to the handler.
func (s *Server) ProcessMessage(update *tgbotapi.Update) {
	ctx := &UpdateContext{
		update: update,
		bot:    s.bot,
	}
	for _, p := range s.processors {
		p(ctx)
		if ctx.stop {
			break
		}
	}

	if update.InlineQuery != nil {
		q := update.InlineQuery
		logrus.WithField("data", q.Query).Info("Got inline query")
		return
	}
	if update.CallbackQuery != nil {
		q := update.CallbackQuery
		logrus.WithField("data", q.Data).Info("Got callback query")
		return
	}
}

// updateUserInfo updates user information in the database
func (s *Server) updateUserInfo(c *UpdateContext) {
	update := c.update
	if update.Message == nil {
		return
	}

	msg := update.Message
	tgUser := msg.From

	var username sql.NullString
	if tgUser.UserName != "" {
		username.String = tgUser.UserName
		username.Valid = true
	}

	user := &structs.User{
		FirstName: tgUser.FirstName,
		LastName:  tgUser.LastName,
		Username:  username,
		TgID:      int64(tgUser.ID),
	}

	_, err := s.storage.UpsertUserByTgID(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to upsert user by tg id")
		return
	}
	return
}

func (s *Server) processInfoMessage(c *UpdateContext) {
	if c.update.Message == nil {
		return
	}
	if c.update.Message.Text != "ℹ️ Инфо" {
		return
	}
	defer c.Stop()

	msg := c.update.Message

	resp := "Я пришлю тебе картинку или текст"

	reply := tgbotapi.NewMessage(msg.Chat.ID, resp)
	reply.ParseMode = "Markdown"
	reply.DisableWebPagePreview = true

	_, err := s.bot.Send(reply)
	if err != nil {
		logrus.WithError(err).Error("Failed to send reply")
		return
	}
}

func (s *Server) processStartMessage(c *UpdateContext) {
	if c.update.Message == nil {
		return
	}
	if c.update.Message.Text != "/start" {
		return
	}
	defer c.Stop()

	msg := c.update.Message
	response := "Привет, %v!\n\n" +
		"Я — бот, присылающий задания"

	response = fmt.Sprintf(response, msg.From.FirstName)

	reply := tgbotapi.NewMessage(msg.Chat.ID, response)
	reply.ParseMode = "Markdown"
	reply.DisableWebPagePreview = true

	locButton := tgbotapi.NewKeyboardButton("📍 Получить задание текстом")
	locRow := []tgbotapi.KeyboardButton{locButton}

	imgButton := tgbotapi.NewKeyboardButton("📍 Получить задание картинкой")
	imgRow := []tgbotapi.KeyboardButton{imgButton}

	homeButton := tgbotapi.NewKeyboardButton("ℹ️ Инфо")
	homeRow := []tgbotapi.KeyboardButton{homeButton}

	markup := [][]tgbotapi.KeyboardButton{locRow, imgRow, homeRow}
	reply.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       markup,
		ResizeKeyboard: true,
	}

	_, err := s.bot.Send(reply)
	if err != nil {
		logrus.WithError(err).Error("Failed to send reply")
		return
	}

	return
}

func (s *Server) processDefaultMessage(c *UpdateContext) {
	if c.update.Message == nil {
		return
	}
	defer c.Stop()

	msg := c.update.Message

	response := "Я ещё только учусь, и не понимаю некоторые сообщения. "

	user, err := s.storage.GetUserByTgID(int64(msg.From.ID))
	if err != nil {
		logrus.WithError(err).Error("Failed to find user by telegram ID")
		return
	}

	err = s.storage.CreateInvalidMessage(user.ID, msg.Text)
	if err != nil {
		logrus.WithError(err).Error("Failed to store invalid message")
		// not exiting here, need to send response
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, response)
	_, err = s.bot.Send(reply)
	if err != nil {
		logrus.WithError(err).Error("Failed to send response")
		return
	}

	return
}

func (s *Server) processImageTaskMessage(c *UpdateContext) {
	if c.update.Message == nil {
		return
	}
	if c.update.Message.Text != "📍 Получить задание картинкой" {
		return
	}
	defer c.Stop()

	msg := c.update.Message

	user, err := s.storage.GetUserByTgID(int64(msg.From.ID))
	if err != nil {
		logrus.WithError(err).Error("Failed to find user by telegram ID")
		return
	}

	task, err := s.storage.GetImageTaskForUser(user.ID)
	if err == db.ErrNotFound {
		logrus.Warningf("no more image tasks for user: %v", user.ID)
		resp := "К сожалению, задания с картинками для тебя кончились."
		reply := tgbotapi.NewMessage(msg.Chat.ID, resp)

		_, err := s.bot.Send(reply)
		if err != nil {
			logrus.WithError(err).Error("Failed to send reply")
			return
		}
		return
	} else if err != nil {
		logrus.Errorf("failed to lookup image task: %+v", err)
		return
	}

	err = s.storage.CreateUserTask(user.ID, task.ID)
	if err != nil {
		logrus.Errorf("failed to create user task: %+v", err)
	}

	filePath := path.Join("./images", task.Image)

	f, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("failed to open image: %+v", err)
	}
	defer f.Close()

	x := tgbotapi.NewPhotoUpload(msg.Chat.ID, tgbotapi.FileReader{Size: -1, Name: "sadsd", Reader: f})
	x.Caption = task.Task

	_, err = s.bot.Send(x)
	if err != nil {
		logrus.WithError(err).Error("Failed to send reply")
		return
	}
}

func (s *Server) processTextTaskMessage(c *UpdateContext) {
	if c.update.Message == nil {
		return
	}
	if c.update.Message.Text != "📍 Получить задание текстом" {
		return
	}
	defer c.Stop()

	msg := c.update.Message

	user, err := s.storage.GetUserByTgID(int64(msg.From.ID))
	if err != nil {
		logrus.WithError(err).Error("Failed to find user by telegram ID")
		return
	}

	task, err := s.storage.GetTextTaskForUser(user.ID)
	if err == db.ErrNotFound {
		logrus.Warningf("no more image tasks for user: %v", user.ID)
		resp := "К сожалению, задания с текстом для тебя кончились."
		reply := tgbotapi.NewMessage(msg.Chat.ID, resp)

		_, err := s.bot.Send(reply)
		if err != nil {
			logrus.WithError(err).Error("Failed to send reply")
			return
		}
		return
	} else if err != nil {
		logrus.Errorf("failed to lookup image task: %+v", err)
		return
	}

	err = s.storage.CreateUserTask(user.ID, task.ID)
	if err != nil {
		logrus.Errorf("failed to create user task: %+v", err)
	}

	resp := task.Task

	reply := tgbotapi.NewMessage(msg.Chat.ID, resp)
	reply.ParseMode = "Markdown"
	reply.DisableWebPagePreview = true

	_, err = s.bot.Send(reply)
	if err != nil {
		logrus.WithError(err).Error("Failed to send reply")
		return
	}
}
