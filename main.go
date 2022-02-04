package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/fatih/color"
)

type user struct {
	Id    int64  `json:id`
	Login string `json:login`
}

type guess struct {
	Guess   string   `json:"guess"`
	Matches []string `json:"matches"`
	IsMatch bool     `json:"isMatch"`
}

type gameStatus struct {
	Guesses []guess `json:"guesses"`
	Status  string  `json:"status"`
}

type gameLocator struct {
	host string
	test bool
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Expected 'wordle <org[/repo] <status | guess <word>>'")
		os.Exit(1)
	}

	user, err := getUser()
	if err != nil {
		fmt.Println("Couldn't figure out who you are. Did you login?")
		fmt.Println(err)
		os.Exit(1)
	}

	var locator = gameLocator{test: false, host: os.Args[1]}
	if os.Args[len(os.Args)-1] == "-test" {
		locator.test = true
	}

	switch os.Args[2] {
	case "guess":
		if len(os.Args) < 4 {
			fmt.Println("You need a guess to guess")
			os.Exit(1)
		}
		locator.doGuess(os.Args[3], user)
	case "status":
		locator.doStatus(user)
	default:
		fmt.Println("Expected 'wordle <org[/repo] <status | guess <word>>'")
		os.Exit(1)
	}
}

func getUser() (*user, error) {
	client, err := gh.RESTClient(nil)
	if err != nil {
		return nil, err
	}
	var response user
	err = client.Get("user", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (x gameLocator) doStatus(user *user) {
	game, err := x.sendStatus(user)
	if err != nil {
		fmt.Println("Error getting your Wordle game")
		fmt.Printf("%+v\n", err)
		return
	}

	if game == nil {
		fmt.Println("No game running today. Yet...")
		return
	}
	x.printGame(game)
}

func (x gameLocator) sendStatus(user *user) (*gameStatus, error) {
	url := x.getUrl() + "/status"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return x.send(req, user)
}

func (x gameLocator) doGuess(guess string, user *user) {
	game, err := x.sendGuess(guess, user)
	if err != nil {
		fmt.Printf("Error getting your Wordle game\n")
		fmt.Printf("%+v\n", err)
		return
	}

	if game == nil {
		fmt.Printf("Hmmm, that was an invalid guess\n")
		return
	}

	fmt.Printf("%s\n", x.getGuessComment(game.Guesses[len(game.Guesses)-1]))
	x.printGame(game)
}

func (x gameLocator) sendGuess(word string, user *user) (*gameStatus, error) {
	url := x.getUrl() + "/" + word
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	return x.send(req, user)
}

func (x gameLocator) getGuessComment(latest guess) string {
	if latest.IsMatch {
		return "Awesome! You won!"
	}
	score := 0
	for _, match := range latest.Matches {
		switch match {
		case "green":
			score = score + 2
		case "yellow":
			score++
		}
	}
	if score > 6 {
		return "Pretty good. You're getting there"
	}
	if score > 3 {
		return "Not too shabby."
	}
	return "Bummer. Try something like 'tears'"
}

func (x gameLocator) printGame(game *gameStatus) {
	green := color.New(color.FgGreen).PrintfFunc()
	yellow := color.New(color.FgYellow).PrintfFunc()
	gray := color.New(color.FgHiBlack).PrintfFunc()

	for _, guess := range game.Guesses {
		for i := range guess.Guess {
			letter := string(guess.Guess[i])
			switch guess.Matches[i] {
			case "green":
				green(" %s", letter)
			case "yellow":
				yellow(" %s", letter)
			case "gray":
				gray(" %s", letter)
			}
		}
		fmt.Println()
	}
}

func (x gameLocator) send(req *http.Request, user *user) (*gameStatus, error) {
	req.Header.Set("Extendo-ActorLogin", user.Login)
	req.Header.Set("Extendo-ActorId", strconv.FormatInt(user.Id, 10))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		return nil, nil
	} else if resp.StatusCode > 299 {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result gameStatus
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (x gameLocator) getUrl() string {
	kind := "repos"
	if !strings.Contains(x.host, "/") {
		kind = "orgs"
	}
	server := "extendocompute"
	if x.test {
		server = "extendotest"
	}
	return fmt.Sprintf("http://%s.eastus.cloudapp.azure.com:3000/rest/%s/%s/wordle", server, kind, x.host)
}
