package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"text/template"
)

type GameServer struct {
	Board   [][]bool
	Running bool
	mu      sync.Mutex
	tmpl    *template.Template
}

func main() {
	size := 75
	gs := &GameServer{
		Board: make([][]bool, size),
		tmpl:  template.Must(template.ParseFiles("templates/home.html")),
	}
	for i := range gs.Board {
		gs.Board[i] = make([]bool, size)
	}

	// 2. Map routes to handler functions
	http.HandleFunc("/", gs.serveIndex)
	http.HandleFunc("/board", gs.serveBoard)
	http.HandleFunc("/start", gs.handleStart)
	http.HandleFunc("/stop", gs.handleStop)

	// 3. Start the server
	println("Server starting at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func (gs *GameServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Serve index")
	_ = gs.tmpl.ExecuteTemplate(w, "home", gs)
}

// serveBoard renders ONLY the "board" template fragment
// Logic: If GameServer.Running is true, calculate the next generation
// before rendering the template.
func (gs *GameServer) serveBoard(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside Serve board")
	gs.Board = generateNextFrame(gs.Board)
	_ = gs.tmpl.ExecuteTemplate(w, "board", gs)
}

// handleStart sets the Running state to true
func (gs *GameServer) handleStart(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside handle Start")

	gs.mu.Lock()
	gs.Running = true
	gs.Board = generateRandomBoard(gs.Board)
	gs.mu.Unlock()

	_ = gs.tmpl.ExecuteTemplate(w, "board-wrapper", gs)
}

// handleStop sets the Running state to false
func (gs *GameServer) handleStop(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside handle stop")

	gs.mu.Lock()
	gs.Running = false
	gs.mu.Unlock()

	_ = gs.tmpl.ExecuteTemplate(w, "board-wrapper", gs)
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
			//get the left
			// if b[i-1][ii] == true {
			// 	neighbors++
			// }
			// //get the right
			// if b[i+1][ii] == true {
			// 	neighbors++
			// }
			// //get the top
			// if b[i][ii+1] == true {
			// 	neighbors++
			// }
			// //get the bottom
			// if b[i][ii-1] == true {
			// 	neighbors++
			// }
			// //get the top right
			// if b[i+1][ii+1] == true {
			// 	neighbors++
			// }
			// //get the top left
			// if b[i-1][ii+1] == true {
			// 	neighbors++
			// }
			// //get the bottom right
			// if b[i+1][ii-1] == true {
			// 	neighbors++
			// }
			// //get teh bottom left
			// if b[i-1][ii-1] == true {
			// 	neighbors++
			// }

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
