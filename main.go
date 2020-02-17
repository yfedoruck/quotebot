package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

type Config struct {
	TelegramBotToken string
}

func token() string {
	file, err := os.Open(basePath() + filepath.FromSlash("/config.json"))
	check(err)

	decoder := json.NewDecoder(file)
	configuration := Config{}

	err = decoder.Decode(&configuration)
	check(err)

	return configuration.TelegramBotToken
}

func quotesByAuthor(name string, authorList []Author) (Author, error) {
	for _, author := range authorList {
		if author.Name == name {
			return author, nil
		}
	}
	return Author{}, errors.New("author not found")
}

func formatQuote(quote, author string) string {
	return fmt.Sprintf(
		"“_%s_“\n "+
			"``` — %s ```\n",
		quote,
		author)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(token())
	check(err)

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	check(err)

	var session map[int]string
	session = make(map[int]string)
	var library Library

	for update := range updates {
		if update.CallbackQuery != nil {
			philosopher := update.CallbackQuery.Data

			session[update.CallbackQuery.From.ID] = philosopher

			_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			check(err)

			library = quotes(library)
			author, err := quotesByAuthor(philosopher, library.AuthorList)
			check(err)

			l := len(author.Quote.TextList)
			num := rand.Intn(l)
			text := "*" + author.Name + " selected*\n\n"
			text += formatQuote(author.Quote.TextList[num], author.FullName) + " /next"
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			msg.ParseMode = "markdown"
			_, err = bot.Send(msg)
			check(err)
			continue
		}

		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "help":
			msg.Text = "Bot prints quotes of famous philosophers. Type /rand or /list"
		case "rand":
			library = quotes(library)
			l := len(library.AuthorList)
			num := rand.Intn(l)
			philosopher := library.AuthorList[num]

			author, err := quotesByAuthor(philosopher.Name, library.AuthorList)
			check(err)

			l2 := len(author.Quote.TextList)
			num2 := rand.Intn(l2)
			msg.Text = formatQuote(author.Quote.TextList[num2], author.FullName) + " /rand"
		case "list":
			library = quotes(library)
			msg = authorColumn(update.Message.Chat.ID, library)
			msg.Text = "*Select philosopher*"
		case "next":
			if philosopher, ok := session[update.Message.From.ID]; ok {
				library = quotes(library)
				author, err := quotesByAuthor(philosopher, library.AuthorList)
				check(err)

				l := len(author.Quote.TextList)
				num := rand.Intn(l)
				msg.Text = formatQuote(author.Quote.TextList[num], author.FullName) + " /next "
			} else {
				msg.Text = "*None philosopher is selected* /list"
			}
		case "clear":
			delete(session, update.Message.From.ID)
			msg.Text = "*None philosopher is selected* /list"
		default:
			msg.Text = "I don't know that command"
		}
		msg.ParseMode = "markdown"

		_, err := bot.Send(msg)
		check(err)
	}
}

type Library struct {
	XMLName    xml.Name `xml:"library"`
	AuthorList []Author `xml:"author"`
	loaded     bool
}

type Quote struct {
	XMLName  xml.Name `xml:"quote"`
	TextList []string `xml:"text"`
}

type Author struct {
	XMLName  xml.Name `xml:"author"`
	Name     string   `xml:"name"`
	FullName string   `xml:"full_name"`
	Quote    Quote    `xml:"quote"`
}

func basePath() string {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		log.Panic("Caller error")
	}
	return filepath.Dir(b)
}

func quotes(library Library) Library {

	if library.loaded {
		return library
	}

	file, err := os.Open(basePath() + filepath.FromSlash("/data/quote.xml"))
	check(err)

	fi, err := file.Stat()
	check(err)

	var data = make([]byte, fi.Size())
	_, err = file.Read(data)
	check(err)

	err = xml.Unmarshal(data, &library)
	check(err)

	library.loaded = true
	return library
}

func authorGrid(chatId int64, library Library) tgbotapi.MessageConfig {
	var row1, row2 []tgbotapi.InlineKeyboardButton

	msg := tgbotapi.NewMessage(chatId, "")
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	for i, author := range library.AuthorList {
		btn := tgbotapi.NewInlineKeyboardButtonData(author.Name, author.Name)
		if i%2 == 0 {
			row1 = append(row1, tgbotapi.NewInlineKeyboardRow(btn)...)
		} else {
			row2 = append(row2, tgbotapi.NewInlineKeyboardRow(btn)...)
		}
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row1, row2)
	msg.ReplyMarkup = keyboard
	return msg
}

func authorColumn(chatId int64, library Library) tgbotapi.MessageConfig {
	var (
		row []tgbotapi.InlineKeyboardButton
		btn tgbotapi.InlineKeyboardButton
	)

	msg := tgbotapi.NewMessage(chatId, "")
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	for _, author := range library.AuthorList {
		btn = tgbotapi.NewInlineKeyboardButtonData(author.Name, author.Name)
		row = tgbotapi.NewInlineKeyboardRow(btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	msg.ReplyMarkup = keyboard
	return msg
}
