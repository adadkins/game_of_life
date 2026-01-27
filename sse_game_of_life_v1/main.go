package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"text/template"
	"time"
)

type GameServer struct {
	Board      [][]bool
	updateChan chan (bool)
	Running    bool
	mu         sync.Mutex
	tmpl       *template.Template
}

func main() {
	size := 75
	gs := &GameServer{
		updateChan: make(chan (bool)),
		Board:      make([][]bool, size),
		tmpl:       template.Must(template.ParseFiles("templates/home.html")),
	}
	for i := range gs.Board {
		gs.Board[i] = make([]bool, size)
	}

	// 2. Map routes to handler functions
	http.HandleFunc("/", gs.getIndex)
	http.HandleFunc("/board", gs.getBoard)
	http.HandleFunc("/start", gs.postStart)
	http.HandleFunc("/stop", gs.postStop)
	http.HandleFunc("/sse", gs.sse)

	// 3. Start the server
	println("Server starting at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func (gs *GameServer) getIndex(w http.ResponseWriter, r *http.Request) {
	_ = gs.tmpl.ExecuteTemplate(w, "home", gs)
}

func (gs *GameServer) gameLooper() {
	for {
		if gs.Running {
			gs.updateChan <- true
			gs.mu.Lock()
			gs.Board = generateNextFrame(gs.Board)
			gs.mu.Unlock()
		}
		time.Sleep(333 * time.Millisecond)
	}
}

// serveBoard renders ONLY the "board" template fragment
func (gs *GameServer) getBoard(w http.ResponseWriter, r *http.Request) {
	_ = gs.tmpl.ExecuteTemplate(w, "board", gs)
}

// handleStart sets the Running state to true
func (gs *GameServer) postStart(w http.ResponseWriter, r *http.Request) {
	gs.mu.Lock()
	gs.Running = true
	gs.Board = generateRandomBoard(gs.Board)
	gs.mu.Unlock()

	go gs.gameLooper()

	_ = gs.tmpl.ExecuteTemplate(w, "board-wrapper", gs)
}

// handleStop sets the Running state to false
func (gs *GameServer) postStop(w http.ResponseWriter, r *http.Request) {
	gs.mu.Lock()
	gs.Running = false
	gs.mu.Unlock()

	_ = gs.tmpl.ExecuteTemplate(w, "board-wrapper", gs)
}

// sse stays open and writes new updates to the writer. the \n\n tells the writer to send to the browser
// current implementation sends boolean for front end to GET new game state from get endpoint
func (gs *GameServer) sse(w http.ResponseWriter, r *http.Request) {
	// keep the sse connection open
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-r.Context().Done():
			// The user closed the tab. The "call" is over.
			return
		case <-gs.updateChan:
			// listen to updates from the game engine
			// send new event to the writer, sse defaults to "event: message" so not necessary
			// fmt.Fprintf(w, "event: message\n")
			fmt.Fprintf(w, "data: \n\n")
			w.(http.Flusher).Flush()
		}
	}
}

// for testing, lets just generate random board
func generateRandomBoard(b [][]bool) [][]bool {
	for i := range len(b) {
		for ii := range len(b[0]) {
			randomNum := rand.IntN(50)
			if randomNum%2 == 0 {
				b[i][ii] = false
			} else {
				b[i][ii] = true
			}
		}
	}
	return b
}

func generateNextFrame(b [][]bool) [][]bool {
	newBoard := make([][]bool, len(b))
	for i := range newBoard {
		newBoard[i] = make([]bool, len(b[0]))
	}

	rows := len(b)
	cols := len(b[0])

	for i := range rows {
		for ii := range cols {
			neighbors := 0
			for x := -1; x <= 1; x++ {
				for y := -1; y <= 1; y++ {
					if x == 0 && y == 0 {
						continue
					}

					// Calculate wrapped coordinates
					// Adding 'rows' before modulo handles negative results (like 0 - 1)
					edgeI := (i + x + rows) % rows
					edgeII := (ii + y + cols) % cols

					if b[edgeI][edgeII] == true {
						neighbors++
					}
				}
			}

			//if an alive cell has 2 or 3 neighbors it lives
			if b[i][ii] == true && (neighbors == 2 || neighbors == 3) {
				newBoard[i][ii] = true
			}
			//if a dead cell has exactly 3 neighbors it lives
			if b[i][ii] == false && neighbors == 3 {
				newBoard[i][ii] = true
			}
		}
	}
	return newBoard
}
