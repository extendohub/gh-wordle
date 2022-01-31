package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'guess' or 'status' commands")
		os.Exit(1)
	}

	user, err := getUser()
	if err != nil {
		fmt.Println("Couldn't figure out who you are. Did you login?")
		fmt.Println(err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "guess":
		if len(os.Args) < 3 {
			fmt.Println("You need a guess to guess")
			os.Exit(1)
		}
		doGuess(os.Args[2], user)
	case "status":
		doStatus(user)
	default:
		fmt.Println("Expected 'guess' or 'status' commands")
		os.Exit(1)
	}
}

func doStatus(user *user) {
	game, err := sendStatus(user)
	if err != nil {
		fmt.Println("Error getting your Wordle game")
		fmt.Printf("%+v\n", err)
		return
	}

	if game == nil {
		fmt.Println("No game running today. Yet...")
		return
	}
	printGame(game)
}

func sendStatus(user *user) (*gameStatus, error) {
	url := "http://extendocompute.eastus.cloudapp.azure.com:3000/rest/repos/app-extensions/extendo-wordle/wordle/status"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return send(req, user)
}

func doGuess(guess string, user *user) {
	game, err := sendGuess(guess, user)
	if err != nil {
		fmt.Printf("Error getting your Wordle game\n")
		fmt.Printf("%+v\n", err)
		return
	}

	if game == nil {
		fmt.Printf("Hmmm, that was an invalid guess\n")
		return
	}

	fmt.Printf("%s\n", getGuessComment(game.Guesses[len(game.Guesses)-1]))
	printGame(game)
}

func sendGuess(word string, user *user) (*gameStatus, error) {
	url := "http://extendocompute.eastus.cloudapp.azure.com:3000/rest/repos/app-extensions/extendo-wordle/wordle/" + word
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	return send(req, user)
}

func getGuessComment(latest guess) string {
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

func printGame(game *gameStatus) {
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

func send(req *http.Request, user *user) (*gameStatus, error) {
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
