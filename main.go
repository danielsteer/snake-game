package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

const (
	running = "running"
	over    = "over"

	up    = "up"
	down  = "down"
	left  = "left"
	right = "right"

	snakeASCII      = 42
	foodASCII       = 48
	whiteSpaceASCII = 32
	borderASCII     = 43
)

type game struct {
	state string
	score int
}

func newGame() *game {
	return &game{
		state: running,
		score: 0,
	}
}

type point struct {
	x int
	y int
}

type snake struct {
	points    []point
	direction string
}

func newSnake() *snake {
	return &snake{
		points:    []point{{10, 10}},
		direction: right,
	}
}

type food struct {
	border *border
	point  point
	kind   int
}

func newFood(b *border) *food {

	food := food{
		border: b,
		kind:   foodASCII,
	}
	food.spawn()

	return &food
}

type border struct {
	x, y   int
	points map[point]bool
}

func newBorder(x int, y int) *border {

	m := map[point]bool{}

	// add top side
	for i := 0; i < x; i++ {
		m[point{x: i, y: 0}] = true
	}

	// add right side
	for i := 0; i < y; i++ {
		m[point{x: x - 1, y: i}] = true
	}

	// add left side
	for i := 0; i < y; i++ {
		m[point{x: 0, y: i}] = true
	}

	// add bottom side
	for i := 0; i < x; i++ {
		m[point{x: i, y: y - 1}] = true
	}

	return &border{x: x, y: y, points: m}
}

func main() {
	startup()

	// Initialize the game state.
	game := newGame()
	snake := newSnake()
	border := newBorder(20, 20)
	food := newFood(border)
	reader := bufio.NewReader(os.Stdin)

	// Start a goroutine to read user input and update the snake's direction.
	go func() {
		for {
			if char, err := reader.ReadByte(); err == nil {
				key := string(char)

				if key == "w" {
					snake.direction = up
				} else if key == "s" {
					snake.direction = down
				} else if key == "a" {
					snake.direction = left
				} else if key == "d" {
					snake.direction = right
				}
			}
		}
	}()

	// Start a goroutine to render the game state.
	go func() {
		for {
			// Clear the screen.
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()

			// move the snake
			snake.move(game, food, border)

			// Render the snake, border and food
			for y := 0; y < border.y; y++ {
				chars := make([]byte, border.x+1)
				for x := 0; x < border.x; x++ {
					chars[x] = whiteSpaceASCII

					// Render Border
					if border.points[point{x, y}] {
						chars[x] = borderASCII
					}

					// Render Snake
					for _, point := range snake.points {
						if x == point.x && y == point.y {
							chars[x] = snakeASCII
						}
					}

					// Render Food
					if x == food.point.x && y == food.point.y {
						chars[x] = byte(food.kind)
					}

				}

				fmt.Println(string(chars))
			}

			fmt.Println("Score: ", game.score)

			// Wait for a frame to elapse.
			time.Sleep(time.Millisecond * 200)
		}
	}()

	// wait for game over
	for {
		if game.state == over {
			break
		}
	}
	shutdown()
}

// move moves the snake
func (s *snake) move(g *game, f *food, b *border) {

	var headPoint point
	var newPoints []point

	switch s.direction {
	case up:
		headPoint = point{x: s.points[0].x, y: s.points[0].y - 1}
	case down:
		headPoint = point{x: s.points[0].x, y: s.points[0].y + 1}
	case left:
		headPoint = point{x: s.points[0].x - 1, y: s.points[0].y}
	case right:
		headPoint = point{x: s.points[0].x + 1, y: s.points[0].y}
	default:
	}

	newPoints = []point{headPoint}

	switch {
	case headPoint.onFood(f):
		g.score++
		newPoints = append(newPoints, s.points[:len(s.points)]...)
		s.points = newPoints
		f.spawn()
	case headPoint.onSnake(s):
		g.state = over
	case headPoint.onBorder(b):
		g.state = over
	default:
		newPoints = append(newPoints, s.points[:len(s.points)-1]...)
		s.points = newPoints
	}

}

func startup() {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// Hide the cursor
	fmt.Fprintf(os.Stdout, "\x1b[?25l")
}

func shutdown() {
	// Show the cursor
	fmt.Fprint(os.Stdout, "\x1b[?25h")
	// Enable input buffering
	exec.Command("stty", "-F", "/dev/tty", "echo").Run()
}

// spawn spawns a single food point
func (f *food) spawn() {
	x, y := 0, 0

	for x == 0 {
		x = rand.Intn(f.border.x - 1)
	}

	for y == 0 {
		y = rand.Intn(f.border.y - 1)
	}

	f.point.x = x
	f.point.y = y
}

// onFood returns true if point is on the food
func (p point) onFood(f *food) bool {
	return p == f.point
}

// onBorder returns true if point is on the border
func (p point) onBorder(b *border) bool {
	return b.points[p]
}

// onSnake returns true if point is on the snake
func (p point) onSnake(s *snake) bool {

	for _, sp := range s.points {
		if p == sp {
			return true
		}
	}

	return false
}
