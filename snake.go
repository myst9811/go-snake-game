package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Game configuration constants
const (
	InitialSpeed = 150 * time.Millisecond // Game tick speed
	BoardWidth   = 40                     // Width of game board
	BoardHeight  = 20                     // Height of game board
)

// Point represents a coordinate on the game board
type Point struct {
	X, Y int
}

// Snake represents the player-controlled snake
type Snake struct {
	Body      []Point // Array of body segments (head is index 0)
	Direction Point   // Current movement direction
}

// Game holds the entire game state
type Game struct {
	Snake    *Snake // Pointer to the snake
	Food     Point  // Current food location
	Score    int    // Player's score
	GameOver bool   // Game state flag
	Width    int    // Board width
	Height   int    // Board height
}

func NewGame() *Game {
	// Create initial snake in the center, 3 segments long
	snake := &Snake{
		Body: []Point{
			{X: BoardWidth / 2, Y: BoardHeight / 2},   // Head
			{X: BoardWidth/2 - 1, Y: BoardHeight / 2}, // Body segment 1
			{X: BoardWidth/2 - 2, Y: BoardHeight / 2}, // Body segment 2
		},
		Direction: Point{X: 1, Y: 0}, // Start moving right
	}

	// Create game instance
	game := &Game{
		Snake:    snake,
		Width:    BoardWidth,
		Height:   BoardHeight,
		Score:    0,
		GameOver: false,
	}

	// Place first food
	game.SpawnFood()
	return game
}
func main() {
	// Configure terminal for raw input (no buffering, no echo)
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	// Hide cursor for cleaner display
	fmt.Print("\033[?25l")

	// Show cursor when program exits
	defer fmt.Print("\033[?25h")

	// Create and run the game
	game := NewGame()
	game.Run()
}

func (g *Game) Run() {
	// Create buffered channel for input
	inputChan := make(chan rune, 10)

	// Start goroutine for input handling
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := reader.ReadRune()
			if err != nil {
				return
			}
			inputChan <- char // Send input to channel
		}
	}()

	// Create ticker for game updates
	ticker := time.NewTicker(InitialSpeed)
	defer ticker.Stop()

	// Main game loop
	for !g.GameOver {
		select {
		case char := <-inputChan:
			// Handle user input
			switch char {
			case 'w', 'W':
				if g.Snake.Direction.Y == 0 { // Prevent 180° turn
					g.Snake.Direction = Point{X: 0, Y: -1}
				}
			case 's', 'S':
				if g.Snake.Direction.Y == 0 {
					g.Snake.Direction = Point{X: 0, Y: 1}
				}
			case 'a', 'A':
				if g.Snake.Direction.X == 0 {
					g.Snake.Direction = Point{X: -1, Y: 0}
				}
			case 'd', 'D':
				if g.Snake.Direction.X == 0 {
					g.Snake.Direction = Point{X: 1, Y: 0}
				}
			case 'q', 'Q':
				g.GameOver = true
				return
			}
		case <-ticker.C:
			// Update game state on each tick
			g.Update()
			g.Draw()
		}
	}

	// Show game over screen
	g.ShowGameOver()
	time.Sleep(3 * time.Second)
}
func (g *Game) Update() {
	// Calculate new head position
	head := g.Snake.Body[0]
	newHead := Point{
		X: head.X + g.Snake.Direction.X,
		Y: head.Y + g.Snake.Direction.Y,
	}

	// Check wall collision
	if newHead.X < 0 || newHead.X >= g.Width ||
		newHead.Y < 0 || newHead.Y >= g.Height {
		g.GameOver = true
		return
	}

	// Check self collision
	for _, segment := range g.Snake.Body {
		if newHead.X == segment.X && newHead.Y == segment.Y {
			g.GameOver = true
			return
		}
	}

	// Add new head to front
	g.Snake.Body = append([]Point{newHead}, g.Snake.Body...)

	// Check if food is eaten
	if newHead.X == g.Food.X && newHead.Y == g.Food.Y {
		g.Score++
		g.SpawnFood()
		// Don't remove tail - snake grows!
	} else {
		// Remove tail - snake maintains size
		g.Snake.Body = g.Snake.Body[:len(g.Snake.Body)-1]
	}
}

func (g *Game) Draw() {
	// Clear screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")

	// Create 2D board array
	board := make([][]rune, g.Height)
	for i := range board {
		board[i] = make([]rune, g.Width)
		for j := range board[i] {
			board[i][j] = ' ' // Empty space
		}
	}

	// Place food on board
	board[g.Food.Y][g.Food.X] = '♥'

	// Place snake on board
	for i, segment := range g.Snake.Body {
		if segment.Y >= 0 && segment.Y < g.Height &&
			segment.X >= 0 && segment.X < g.Width {
			if i == 0 {
				board[segment.Y][segment.X] = '■' // Head
			} else {
				board[segment.Y][segment.X] = '●' // Body
			}
		}
	}

	// Draw top border
	fmt.Print("┌" + strings.Repeat("─", g.Width) + "┐\n")

	// Draw board with side borders
	for _, row := range board {
		fmt.Print("│")
		for _, cell := range row {
			fmt.Printf("%c", cell)
		}
		fmt.Print("│\n")
	}

	// Draw bottom border
	fmt.Print("└" + strings.Repeat("─", g.Width) + "┘\n")

	// Draw HUD
	fmt.Printf("Score: %d\n", g.Score)
	fmt.Println("Controls: W/A/S/D to move | Q to quit")
}

func (g *Game) SpawnFood() {
	for {
		// Random position
		food := Point{
			X: rand.Intn(g.Width),
			Y: rand.Intn(g.Height),
		}

		// Check if food spawns on snake
		collision := false
		for _, segment := range g.Snake.Body {
			if food.X == segment.X && food.Y == segment.Y {
				collision = true
				break
			}
		}

		// If no collision, place food
		if !collision {
			g.Food = food
			return
		}
		// Otherwise, try again (loop continues)
	}
}
