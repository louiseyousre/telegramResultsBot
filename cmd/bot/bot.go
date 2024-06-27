package main

import (
	"context"
	"errors"
	"fmt"
	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"net/http"
	"strings"
	"telegramResultsBot/internal/portal"
)

type bot struct {
	portalService *portal.Service
}

func newBot() *bot {
	return &bot{portalService: portal.NewService()}
}

func (_ *bot) matchGetResults(update *models.Update) bool {
	if update.Message == nil {
		return false
	}

	text := strings.TrimSpace(update.Message.Text)
	if text == "" {
		return false
	}

	textLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(textLines) != 2 {
		return false
	}

	return true
}

func (r *bot) getResultsHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	textLines := strings.Split(strings.ReplaceAll(strings.TrimSpace(update.Message.Text), "\r\n", "\n"), "\n")

	nationalID := strings.TrimSpace(textLines[0])
	if len(nationalID) != 14 {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "السطر الأول في الرسالة اللي بعتها مفيهاش رقم قومي صح لأنه المفروض يبقى 14 رقم. إتأكد تاني وابعت من جديد.",
			ReplyParameters: replyParametersTo(update.Message),
		})

		if err != nil {
			log.Print(err)
		}

		return
	}

	password := strings.TrimSpace(textLines[1])
	if len(password) == 0 {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "السطر التاني مفيهوش باسورد. إتأكد إنك مش حاطط مثلاً مسافات أو حاجة شبه كدا وإبعت تاني من جديد.",
			ReplyParameters: replyParametersTo(update.Message),
		})

		if err != nil {
			log.Print(err)
		}

		return
	}

	message, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID:          update.Message.Chat.ID,
		Text:            "ثواني بنحاول نسجل الدخول...",
		ReplyParameters: replyParametersTo(update.Message),
	})
	if err != nil {
		log.Print(err)
	}

	var cookie *http.Cookie
	cookie, err = r.portalService.Login(nationalID, password)
	if err != nil {
		log.Print(err)

		if errors.Is(err, portal.ErrInvalidCredentials) {
			_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID:          update.Message.Chat.ID,
				Text:            "راجع بيانات الدخول تاني كدا الموقع مش قابلها. إفتكر إن الرقم القومي تكتبه في السطر الأول والباسورد في السطر التاني.",
				ReplyParameters: replyParametersTo(update.Message),
			})
		} else {
			_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID:          message.ID,
				Text:            "معلهش متأسفين. حصل خطأ مش عارفينه إيه هو بالظبط ولكن هنحاول نشوفه إيه هو ولو ينفع يتصلح من عندنا هنصلحه.",
				ReplyParameters: replyParametersTo(message),
			})
		}

		if err != nil {
			log.Print(err)
		}

		return
	}
	message, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID:          update.Message.Chat.ID,
		Text:            "تسجيل الدخول نجح. لحظات نجيب بياناتك...",
		ReplyParameters: replyParametersTo(message),
	})
	if err != nil {
		log.Print(err)
	}

	var data *portal.StudentData
	data, err = r.portalService.GetStudentData(cookie)
	if err != nil {
		log.Print(err)

		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "معلهش متأسفين. حصل خطأ مش عارفينه إيه هو بالظبط ولكن هنحاول نشوفه إيه هو ولو ينفع يتصلح من عندنا هنصلحه.",
			ReplyParameters: replyParametersTo(message),
		})

		return
	}

	studentName, err := portal.GetFirstTranslation(data.StdName)
	if err != nil {
		log.Printf("Student name is malformed: %s", data.StdName)

		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "البيانات اللي جات من الموقع مش سليمة ممكن يكون فيه مشكلة دلوقتي أو يكون البوت قدم والموقع حصل فيه تغييرات.",
			ReplyParameters: replyParametersTo(message),
		})

		if err != nil {
			log.Print(err)
		}

		return
	}

	var photo *models.InputFileUpload
	photo, err = getStudentPhoto(data.ImagePath)
	if err != nil {
		log.Print(err)

		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          message.ID,
			Text:            "معلهش متأسفين. حصل خطأ مش عارفينه إيه هو بالظبط ولكن هنحاول نشوفه إيه هو ولو ينفع يتصلح من عندنا هنصلحه.",
			ReplyParameters: replyParametersTo(message),
		})

		return
	}

	message, err = b.SendPhoto(ctx, &telegramBot.SendPhotoParams{
		ChatID:          update.Message.Chat.ID,
		Photo:           photo,
		Caption:         fmt.Sprintf("جبنا بياناتك بنجاح يا %s، حد قالك إنك حد جميل وشكلك حلو ❤️ ثواني هنجيب النتيجة بقى.", *studentName),
		ReplyParameters: replyParametersTo(message),
	})
	if err != nil {
		log.Print(err)
	}

	var results *[]portal.StudentResult
	results, err = r.portalService.GetResults(cookie, data.UUID)
	if err != nil {
		log.Print(err)
		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "معلهش متأسفين. حصل خطأ مش عارفينه إيه هو بالظبط ولكن هنحاول نشوفه إيه هو ولو ينفع يتصلح من عندنا هنصلحه.",
			ReplyParameters: replyParametersTo(message),
		})
		return
	}

	var formattedResults *string
	formattedResults, err = portal.FormatResults(results)
	if err != nil {
		log.Print(err)

		if errors.Is(err, portal.ErrMalformedResults) {
			_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID:          update.Message.Chat.ID,
				Text:            "النتيجة اللي جات من الموقع مش سليمة ممكن يكون فيه مشكلة دلوقتي أو يكون البوت قدم والموقع حصل فيه تغييرات.",
				ReplyParameters: replyParametersTo(message),
			})
		} else {
			_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID:          update.Message.Chat.ID,
				Text:            "معلهش متأسفين. حصل خطأ مش عارفينه إيه هو بالظبط ولكن هنحاول نشوفه إيه هو ولو ينفع يتصلح من عندنا هنصلحه.",
				ReplyParameters: replyParametersTo(message),
			})
		}

		return
	}

	_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID:          update.Message.Chat.ID,
		Text:            *formattedResults,
		ReplyParameters: replyParametersTo(message),
	})
}

func (_ *bot) startHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if update.Message != nil {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "يا مرحب بيك يا صديقي. علشان تجيب النتيجة ابعت رسالة عبارة عن سطرين، السطر الأول منهم فيه الرقم القومي (14 رقم) والسطر التاني في الباسورد بتاعك وإبعت.",
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}
	}
}

func (_ *bot) defaultHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if update.Message != nil {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "بص يا صديقي البوت دا مبيفهمش أي حاجة غير إنك تكتب الرقم القومي في السطر الأول والباسورد في السطر التاني، وبكدا هيحاول يجيب النتيجة بتاعتك ويبعتهالك.",
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}
	}
}
