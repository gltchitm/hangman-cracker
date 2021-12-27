# Hangman Cracker
Program to crack reliably crack the word in [Hangman](https://github.com/gltchitm/hangman).

## Running
Execute `cracker.go` to begin.

#### Options
```
-forever
    keep creating a new game and trying again
-quiet
    only log the game result
-url string
    url of the game websocket (default "ws://localhost:5522/ws")
-wordsurl string
    url of the rand_word.ex file to fetch (default "https://raw.githubusercontent.com/gltchitm/hangman/master/lib/hangman/rand_word.ex")
```

## Methodology
> **Note**: This section refers to [this specific implementation of Hangman](https://github.com/gltchitm/hangman).

The game operates in a way that prevents the client from knowing what the answer is until the game is finished. It runs over a WebSocket and the client is only told what it needs to know. However, the client can still determine the answer before it runs out of lives using the publicly available word list and a few of heuristics.

Do note that, while it would technically be possible, this program *does not* use the Guess Word functionality. The Guess Word operation allows a client to guess an entire word without losing a life if it is wrong, but the server enforces a 5-second delay between uses. Because of this delay, attempting all of the possible answers would be extremely slow, coming nowhere close to [this program's speeds](#speed).

Once started, this program downloads the complete word list used by the game. It then opens a WebSocket to the game server, pretending to be a normal client. It then starts a regular local game and begins the guess loop.

The game server sends an Update message to the client after each guess, informing it of the partial word (e.g. "\_E\_\_O" if the complete word is "HELLO" and the clent has guessed "E" and "O") and the lives remaining.

Each cycle of the guess loop begins by attempting to narrow down the amount of possible words in the word list. Three heuristics are used to do this:

**Length**

The client takes the length of the partial game state to remove all words from the word list with non-matching lengths (e.g. the partial game state of "E___" tells the client that the answer cannot be "EXAMPLE" as it is not 4 letters long).

**Guessed**

After guessing, the client can use the partial game state to eliminate even more possible answers from the word list. If the client has guessed the letters "E" and "O" and the partial game state is "\_E\_\_O", the client knows that the answer cannot be "HEAT" as it does not contain "O" while the partial game state does. Likewise, if the client has guessed "W" and "D" and the partial game state is "W\_\_\_" it knows the answer cannot be "WORD" as it does not contain D.

**Structure**

The structure of the partial word can also be used by the client to eliminate more possibilities from the word list. For example, if the client has guessed "I" and the partial word is "\_\_I\_\_" it knows that the answer cannot be "KIOSK" because, although it contains it, the letter "I" is not in the correct place.

Once the elimination is performed for the cycle, a letter guess is made. The letter to guess is determined by taking all of words in the word list and finding the most common letter.

This cycle continues until the client has either won or lost.

Because this program does not use the Guess Word operation at all, there are often times where the answer is already known but every letter must still be "guessed" manually.

## Accuracy
This program is very accurate. In a trial of 10,000 games, it was successfully able to determine the answer to every game without a single loss. Because the method used for finding the answer is deterministic, every single word can be tested for accuracy.

## Speed
Because it does not use the Guess Word operation, this program is able to find the answer in a fraction of a second even though it is not optimized for speed. Of course, the time this program takes to complete is impacted by other factors such as network speeds and the speed of the game server.

## License
[MIT](LICENSE)
