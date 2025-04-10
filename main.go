package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	. "github.com/xpetit/x/v5"
	"golang.org/x/term"
)

const (
	// Keys
	up byte = iota + 65
	down
	right
	left
	ctrlC  = 3
	escape = 27
	arrows = 91

	width  = 20
	height = 20
)

type coord struct{ x, y int }

var vector = map[byte]coord{
	left:  {-1, 0},
	right: {+1, 0},
	up:    {0, -1},
	down:  {0, +1},
}

func main() {
	os.Stdout.WriteString("\x1b[?25l\x1b[2J") // Hide cursor & clear screen
	defer os.Stdout.WriteString("\x1b[?25h")  // Show cursor

	oldState := Must(term.MakeRaw(0))
	defer func() { Check(term.Restore(0, oldState)) }()

	key := make(chan byte)
	go func() { // Start keyboard handler
		var b [256]byte
		for {
			switch Must(os.Stdin.Read(b[:])) {
			case 1:
				switch b[0] {
				case ctrlC, escape, 'q':
					key <- escape
				}
			case 3:
				if b[1] == arrows {
					key <- b[2]
				}
			}
		}
	}()

	var grid [height][width]byte
	for y, row := range grid {
		for x := range row {
			grid[y][x] = ' '
		}
	}

	direction := vector[right]
	snakePositions := []coord{{2, 2}, {1, 2}, {0, 2}}
	placeApple := func() {
		for {
			x := rand.Intn(width)
			y := rand.Intn(height)
			if grid[y][x] == ' ' { // Free spot
				grid[y][x] = '#'
				break
			}
		}
	}
	placeApple()

	var score int
	b := make([]byte, 0, 2*(width*2)*(height+2))
	for {
		// Refresh display
		fmt.Print("\x1b[1;1H") // Move cursor to the upper-left corner
		fmt.Printf("score: %d\r\n", score)
		b = b[:0]
		b = append(b, "╭"...)
		b = append(b, strings.Repeat("─", width*2)...)
		b = append(b, "╮\r\n"...)
		for _, row := range grid {
			b = append(b, "│"...)
			for _, sq := range row {
				b = append(b, sq, ' ')
			}
			b = append(b, "│\r\n"...)
		}
		b = append(b, "╰"...)
		b = append(b, strings.Repeat("─", width*2)...)
		b = append(b, "╯\r\n"...)
		os.Stdout.Write(b)

		time.Sleep(time.Second / 8)

		for { // Process incoming keys
			select {
			case key := <-key:
				switch key {
				case escape:
					return
				case left, right, up, down:
					if v := vector[key]; direction.x+v.x != 0 && direction.y+v.y != 0 {
						direction = v
					}
				}
				continue
			default: // Nothing left, will exit the loop
			}
			break
		}

		// Update
		headX := snakePositions[0].x + direction.x
		headY := snakePositions[0].y + direction.y

		outside := headX == -1 || headY == -1 || headX == width || headY == height
		stumbling := grid[headY][headX] == '@'
		if outside || stumbling {
			fmt.Println("game over\r")
			return
		}
		if grid[headY][headX] == '#' { // Eat the apple
			placeApple()
			score++
		} else { // Erase tail
			tail := snakePositions[len(snakePositions)-1]
			grid[tail.y][tail.x] = ' '
			snakePositions = snakePositions[:len(snakePositions)-1]
		}
		// Place head
		snakePositions = append([]coord{{headX, headY}}, snakePositions...)
		grid[headY][headX] = '@'
	}
}
