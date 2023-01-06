package gem

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/stirboy/jh/pkg/utils"
)

type UnluckyError struct{}

func (e *UnluckyError) Error() string {
	return "Not today ;)"
}

func GetRandomGem() {
	rand.Seed(time.Now().Unix())

	min := 0
	max := 100

	r := min + rand.Intn(max-min)
	if r <= 100 {

		joke, err := fetchRandomDadJoke()
		if err != nil {
			// something is wrong, but who cares ¯\_(ツ)_/¯
			return
		}

		fmt.Print(heredoc.Docf(`
		-------------------------------------------------------------------
		YOU ARE IN LUCK!												     
		Here is a dad joke for you ¯\_(ツ)_/¯
			
		%s
		-------------------------------------------------------------------
	`, joke))
	}
}

func fetchRandomDadJoke() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("User-Agent", "jh - jira helper utility")
	req.Header.Set("Accept", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}

	p, err := utils.ParseResponse(resp)
	if err != nil {
		return "", nil
	}

	return string(p), nil
}
