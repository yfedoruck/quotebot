package web

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/yfedoruck/quotebot/pkg/env"
	"github.com/yfedoruck/quotebot/pkg/fail"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	Port string
}

func (s *Server) Start() {

	bot, err := tgbotapi.NewBotAPI(Token())
	fail.Check(err)

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// long pooling
	updates, err := bot.GetUpdatesChan(u)
	fail.Check(err)

	// web hooks for awake heroku from idling
	//updates := bot.ListenForWebhook("/" + bot.Token)

	var session map[int]string
	session = make(map[int]string)
	var library Library

	log.Println("Starting web server on", s.Port)
	go func() {
		if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	for update := range updates {
		if update.CallbackQuery != nil {
			philosopher := update.CallbackQuery.Data

			session[update.CallbackQuery.From.ID] = philosopher

			_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			fail.Check(err)

			library = quotes(library)
			author, err := quotesByAuthor(philosopher, library.AuthorList)
			fail.Check(err)

			l := len(author.Quote.TextList)
			num := rand.Intn(l)
			text := "*" + author.Name + " selected*\n\n"
			text += formatQuote(author.Quote.TextList[num], author.FullName) + " /next"
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			msg.ParseMode = "markdown"
			_, err = bot.Send(msg)
			fail.Check(err)
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
			fail.Check(err)

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
				fail.Check(err)

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
		fail.Check(err)
	}

}

func NewServer() *Server {
	s := &Server{}
	s.Port = env.Port()
	http.HandleFunc("/", MainHandler)
	return s
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	_, err := resp.Write([]byte("Hi there! I'm AnticQuoteBot!"))
	fail.Check(err)
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

func quotes(library Library) Library {

	if library.loaded {
		return library
	}

	file, err := os.Open(env.BasePath() + filepath.FromSlash("/data/quote.xml"))
	fail.Check(err)

	fi, err := file.Stat()
	fail.Check(err)

	var data = make([]byte, fi.Size())
	_, err = file.Read(data)
	fail.Check(err)

	err = xml.Unmarshal(data, &library)
	fail.Check(err)

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
