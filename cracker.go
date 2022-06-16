package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
)

type ServerboundPacket struct {
	Action string `json:"action"`

	// create_game
	GameType string `json:"game_type,omitempty"`

	// guess_letter
	Letter string `json:"letter,omitempty"`
}
type ClientboundPacket struct {
	Message string `json:"message"`

	// update
	GameState string   `json:"word,omitempty"`
	Guessed   []string `json:"letters,omitempty"`
	Lives     int      `json:"lives,omitempty"`
}

func fetchWords(url string) []string {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	wordRegexp, err := regexp.Compile(`"([A-Z]+)\"`)
	if err != nil {
		panic(err)
	}

	matches := wordRegexp.FindAllStringSubmatch(string(body), -1)
	if matches == nil {
		panic("could not find any words!")
	}

	words := make([]string, len(matches))

	for i, match := range matches {
		words[i] = match[1]
	}

	return words
}

func mostCommonLetter(words []string, guessed []string) string {
	seen := make(map[string]int)

	for _, word := range words {
	letterSearch:
		for _, letter := range strings.Split(word, "") {
			for _, guessedLetter := range guessed {
				if guessedLetter == letter {
					continue letterSearch
				}
			}

			_, ok := seen[letter]
			if ok {
				seen[letter] += 1
			} else {
				seen[letter] = 1
			}
		}
	}

	var result string
	var max int

	for letter := range seen {
		if seen[letter] > max {
			result = letter
			max = seen[letter]
		}
	}

	return result
}

func filterWordsByLength(words []string, length int) []string {
	var result []string

	for _, word := range words {
		if len(word) == length {
			result = append(result, word)
		}
	}

	return result
}
func filterWordsByGuessed(words []string, gameState string, guessed []string) []string {
	var result []string

	var correctGuesses []string
	var incorrectGuesses []string
	for _, letter := range guessed {
		if strings.Contains(gameState, letter) {
			correctGuesses = append(correctGuesses, letter)
		} else {
			incorrectGuesses = append(incorrectGuesses, letter)
		}
	}

wordsSearch:
	for _, word := range words {
		for _, correctLetter := range correctGuesses {
			if !strings.Contains(word, correctLetter) {
				continue wordsSearch
			}
		}

		for _, incorrectLetter := range incorrectGuesses {
			if strings.Contains(word, incorrectLetter) {
				continue wordsSearch
			}
		}

		result = append(result, word)
	}

	return result
}
func filterWordsByStructure(words []string, gameState string) []string {
	var result []string

wordSearch:
	for _, word := range words {
		for i, letter := range strings.Split(word, "") {
			if string(gameState[i]) != "_" && string(gameState[i]) != letter {
				continue wordSearch
			}
		}

		result = append(result, word)
	}

	return result
}

func main() {
	url := flag.String("url", "ws://localhost:5522/ws", "url of the game websocket")
	wordsUrl := flag.String(
		"wordsurl",
		"https://raw.githubusercontent.com/gltchitm/hangman/master/server/game/wordlist.go",
		"url of the wordlist.go file to fetch",
	)
	quiet := flag.Bool("quiet", false, "only log the game result")
	forever := flag.Bool("forever", false, "keep creating a new game and trying again")

	flag.Parse()

	words := fetchWords(*wordsUrl)

	connection, _, err := websocket.DefaultDialer.Dial(*url, nil)
	if err != nil {
		panic(err)
	}

	defer connection.Close()

	connection.WriteJSON(ServerboundPacket{
		Action:   "create_game",
		GameType: "local",
	})

	i := 1
	for {
		if !*quiet {
			println("--Attempt " + fmt.Sprint(i) + "--")
		}

		var response ClientboundPacket

		err = connection.ReadJSON(&response)
		if err != nil {
			panic(err)
		}

		if response.Message != "update" {
			panic("received unexpected packet: " + response.Message)
		}

		if !*quiet {
			println("Game State: " + response.GameState)
			println("Lives Remaining: " + fmt.Sprint(response.Lives))
		}

		if !strings.Contains(response.GameState, "_") {
			if !*quiet {
				println()
			}
			if response.Lives > 0 {
				if *quiet {
					println("Win: " + response.GameState)
				} else {
					println("--Game Over: We win! (word was " + response.GameState + ")--")
				}
			} else {
				if *quiet {
					println("Loss: " + response.GameState)
				} else {
					println("--Game Over: We lost :( (word was" + response.GameState + ")--")
				}
			}

			if *forever {
				i = 1
				words = fetchWords(*wordsUrl)
				connection.WriteJSON(ServerboundPacket{Action: "new_game"})
				if !*quiet {
					println("--New Game--")
					println()
				}
				continue
			}

			return
		}

		words = filterWordsByLength(words, len(response.GameState))
		words = filterWordsByGuessed(words, response.GameState, response.Guessed)
		words = filterWordsByStructure(words, response.GameState)

		if !*quiet {
			if len(words) == 1 {
				println("The answer is " + words[0] + " but we are manually sending the letters")
			} else {
				for _, word := range words {
					println("Possible Word: " + word)
				}
			}
		}

		letter := mostCommonLetter(words, response.Guessed)
		if !*quiet {
			println("Guessing: " + letter)
		}

		connection.WriteJSON(ServerboundPacket{
			Action: "guess_letter",
			Letter: letter,
		})

		if !*quiet {
			println()
		}

		i++
	}
}
