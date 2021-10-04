package main

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	url := getUrl()
	var offset int
	for {
		updates, err := getUpdates(url, offset)
		if err != nil {
			log.Println("Error updates", err.Error())
		}
		for _, update := range updates {
			err := respond(url, update)
			offset = update.UpdateId + 1
			if err != nil {
				log.Println("Error respond", err.Error())
			}
		}
	}
}

func getUrl() string {
	botApi := goEnvVariable("TELEGRAM_API")
	botToken := goEnvVariable("TELEGRAM_TOKEN")
	botUrl := botApi + botToken

	return botUrl
}

func goEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func getUpdates(url string, offset int) ([]Update, error) {
	res, err := http.Get(url + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var restResponse Response

	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Result, nil
}

func respond(url string, update Update) error {
	var botMessage BotMessage
	variants := [5]string{
		"камень",
		"ножницы",
		"бумага",
		"ящерица",
		"спок",
	}
	computerAnswer := variants[getComputerAnswer()]
	userAnswer := strings.ToLower(update.Message.Text)

	gameResult := getGameResult(userAnswer, computerAnswer)
	botMessage.ChatId = update.Message.Chat.Id
	botMessage.Text = "Вариант компьютера: " + computerAnswer + ". Значит вы " + gameResult

	buf, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}

	_, err = http.Post(url+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	return nil
}

func getComputerAnswer() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(5)
}

func getGameResult(userAnswer, computerAnswer string) string {
	/*
		правила
		ножницы -> бумагу
		бумага -> камень
		камень -> ящерицу
		ящерица -> спок
		спок -> ножницы
		ножницы -> ящерицу
		ящерица -> бумагу
		бумага -> спок
		спок -> камень
		камень -> ножницы
	*/
	scripts := map[string][]string{
		"ножницы": {"бумага", "ящерица"},
		"бумага":  {"камень", "спок"},
		"ящерица": {"спок", "бумага"},
		"спок":    {"ножницы", "камень"},
		"камень":  {"ножницы", "ящерица"},
	}

	if computerAnswer == userAnswer {
		return "добились ничьи"
	}

	for _, value := range scripts[userAnswer] {
		if value == computerAnswer {
			return "победили!"
		}
	}
	return "проиграли"
}
