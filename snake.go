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

type Point struct {
	X, Y int
}

type Snake struct {
	Body      []Point
	Direction Point
}

type Game struct {
	Snake    *Snake
	Food     Point
	Score    int
	GameOver bool
	Width    int
	Height   int
}

const (
	InitialSpeed = 150 * time.Millisecond
	BoardWidth   = 40
	BoardHeight  = 20
)

func main() {
	// Disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	game := NewGame()
	game.Run()
}

func NewGame() *Game {
	snake := &Snake{
		Body: []Point{
			{X: BoardWidth / 2, Y: BoardHeight / 2},
			{X: BoardWidth/2 - 1, Y: BoardHeight / 2},
			{X: BoardWidth/2 - 2, Y: BoardHeight / 2},
		},
		Direction: Point{X: 1, Y: 0}, // Moving right
	}

	game := &Game{
		Snake:    snake,
		Width:    BoardWidth,
		Height:   BoardHeight,
		Score:    0,
		GameOver: false,
	}

	game.SpawnFood()
	return game
}

func (g *Game) Run() {
	// Channel for input events
	inputChan := make(chan rune, 10)

	// Goroutine for handling input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := reader.ReadRune()
			if err != nil {
				return
			}
			inputChan <- char
		}
	}()

	// Game loop
	ticker := time.NewTicker(InitialSpeed)
	defer ticker.Stop()

	for !g.GameOver {
		select {
		case char := <-inputChan:
			switch char {
			case 'w', 'W':
				if g.Snake.Direction.Y == 0 {
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
			g.Update()
			g.Draw()
		}
	}

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
	if newHead.X < 0 || newHead.X >= g.Width || newHead.Y < 0 || newHead.Y >= g.Height {
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

	// Add new head
	g.Snake.Body = append([]Point{newHead}, g.Snake.Body...)

	// Check if food is eaten
	if newHead.X == g.Food.X && newHead.Y == g.Food.Y {
		g.Score++
		g.SpawnFood()
	} else {
		// Remove tail if no food eaten
		g.Snake.Body = g.Snake.Body[:len(g.Snake.Body)-1]
	}
}

func (g *Game) Draw() {
	// Clear screen
	fmt.Print("\033[H\033[2J")

	// Create the board
	board := make([][]rune, g.Height)
	for i := range board {
		board[i] = make([]rune, g.Width)
		for j := range board[i] {
			board[i][j] = ' '
		}
	}

	// Place food
	board[g.Food.Y][g.Food.X] = '♥'

	// Place snake
	for i, segment := range g.Snake.Body {
		if segment.Y >= 0 && segment.Y < g.Height && segment.X >= 0 && segment.X < g.Width {
			if i == 0 {
				board[segment.Y][segment.X] = '■' // Head
			} else {
				board[segment.Y][segment.X] = '●' // Body
			}
		}
	}

	// Draw top border
	fmt.Print("┌" + strings.Repeat("─", g.Width) + "┐\n")

	// Draw board
	for _, row := range board {
		fmt.Print("│")
		for _, cell := range row {
			fmt.Printf("%c", cell)
		}
		fmt.Print("│\n")
	}

	// Draw bottom border
	fmt.Print("└" + strings.Repeat("─", g.Width) + "┘\n")

	// Draw score and controls
	fmt.Printf("Score: %d\n", g.Score)
	fmt.Println("Controls: W/A/S/D to move | Q to quit")
}

func (g *Game) SpawnFood() {
	for {
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

		if !collision {
			g.Food = food
			return
		}
	}
}

func (g *Game) ShowGameOver() {
	fmt.Print("\033[H\033[2J")
	fmt.Println("\n\n")
	fmt.Println("   ╔═══════════════════════════╗")
	fmt.Println("   ║                           ║")
	fmt.Println("   ║       GAME OVER!          ║")
	fmt.Println("   ║                           ║")
	fmt.Printf("   ║    Final Score: %-3d       ║\n", g.Score)
	fmt.Println("   ║                           ║")
	fmt.Println("   ╚═══════════════════════════╝")
	fmt.Println()
}
