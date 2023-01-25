package gem

import (
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/stirboy/jh/pkg/utils"
	"math/rand"
	"net/http"
	"time"
)

func GetRandomGem() {
	rand.Seed(time.Now().Unix())

	min := 0
	max := 100

	r := min + rand.Intn(max-min)
	if r < 10 {

		joke, err := fetchOfficeUsQuote()
		if err != nil {
			// something is wrong, but who cares ¯\_(ツ)_/¯
			return
		}

		fmt.Print(heredoc.Docf(`
		-------------------------------------------------------------------
		YOU ARE IN LUCK!												     
		Here is an office quote for you ¯\_(ツ)_/¯
			
		%s
		-------------------------------------------------------------------
	`, joke))
	}
}

type OfficeRandomQuoteResponse struct {
	Data struct {
		Id        string `json:"_id"`
		Content   string `json:"content"`
		Character struct {
			Id        string `json:"_id"`
			Firstname string `json:"firstname"`
			Lastname  string `json:"lastname"`
			V         int    `json:"__v"`
		} `json:"character"`
		V int `json:"__v"`
	} `json:"data"`
}

func fetchOfficeUsQuote() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://officeapi.dev/api/quotes/random", nil)
	if err != nil {
		return "", nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}

	p, err := utils.ParseResponse(resp)
	if err != nil {
		return "", nil
	}

	var quote OfficeRandomQuoteResponse
	err = json.Unmarshal(p, &quote)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(quote.Data.Content, "\n", " © ", quote.Data.Character.Firstname, " ", quote.Data.Character.Lastname), nil
}
