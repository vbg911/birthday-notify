package services

import (
	"birthday-notify/internal/config"
	"birthday-notify/internal/models"
	"birthday-notify/internal/storage"
	"github.com/lmittmann/tint"
	"log/slog"
	"net/smtp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

func RunNotifier(repo storage.UserRepository, smtpcfg config.SMTPServer, logger *slog.Logger) {
	currentTime := time.Now()
	peoples, err := repo.FindBirthdayPeoples(currentTime.Day(), int(currentTime.Month()))
	if err != nil {
		logger.Error("[Notifier] Failed to get birthday people", tint.Err(err))
		return
	}

	if len(peoples) == 0 {
		logger.Info("[Notifier] There are no birthday peoples today")
		return
	}

	msg := generateMessages(peoples)
	logger.Info("[Notifier] Text of the emails has been generated", slog.Int("amount", len(msg)))

	sendEmails(msg, smtpcfg, logger)
}

func sendEmails(msg []models.Message, smtpcfg config.SMTPServer, logger *slog.Logger) {
	auth := smtp.PlainAuth("", smtpcfg.From, smtpcfg.Password, smtpcfg.SMTPAddress)
	logger.Info("[Notifier] Start sending emails", slog.Int("amount", len(msg)))
	var wg sync.WaitGroup
	for _, m := range msg {
		wg.Add(1)
		var msgBuilder strings.Builder
		msgBuilder.WriteString("To: ")
		msgBuilder.WriteString(m.Recipient)
		msgBuilder.WriteString("\r\n")
		msgBuilder.WriteString("Subject: ")
		msgBuilder.WriteString(m.EmailSubject)
		msgBuilder.WriteString("\r\n\r\n")
		msgBuilder.WriteString(m.EmailText)
		message := msgBuilder.String()

		go func(m models.Message) {
			defer wg.Done()
			err := smtp.SendMail(smtpcfg.SMTPAddress+":"+smtpcfg.SMTPPort, auth, smtpcfg.From, []string{m.Recipient}, []byte(message))
			if err != nil {
				logger.Error("[Notifier] Failed to send email to ", slog.String("adr", m.Recipient), tint.Err(err))
				return
			}
			logger.Info("[Notifier] Email successfully sent to", slog.String("adr", m.Recipient))
		}(m)
	}
	wg.Wait()
	logger.Info("[Notifier] All emails were sent on ", slog.String("date", time.Now().Format(time.DateOnly)))
}

func generateMessages(users []models.User) []models.Message {
	var recipients []string
	var messages []models.Message

	for _, user := range users {
		recipients = append(recipients, user.Subscribers...)
	}
	recipients = removeDuplicates(recipients)
	emails := make(map[string][]models.User)
	for _, item := range recipients {
		for _, user := range users {
			if slices.Contains(user.Subscribers, item) {
				emails[item] = append(emails[item], user)
			}
		}
	}

	for u, item := range emails {
		var text strings.Builder
		text.WriteString("Сегодня ")
		text.WriteString(time.Now().Format("02-01"))
		text.WriteString(" день рождения празднуют:\n")
		for i, u := range item {
			text.WriteString(strconv.Itoa(i+1) + ") ")
			text.WriteString(u.Email)
			text.WriteString("\n")
		}
		text.WriteString("\n\n не забудь поздравить своих коллег!")
		messages = append(messages, models.Message{
			Recipient:    u,
			EmailSubject: "Дни рождения коллег " + time.Now().Format("02-01-2006"),
			EmailText:    text.String(),
		})
	}

	return messages
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for _, element := range elements {
		if !encountered[element] {
			encountered[element] = true
			result = append(result, element)
		}
	}

	return result
}
