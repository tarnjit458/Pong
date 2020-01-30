package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// Default window size
const windowW int = 800
const windowH int = 600

type aiDifficulty float32

// Enum for the difficulty levels of the AI paddle
const (
	Easy       aiDifficulty = 1.0
	Medium     aiDifficulty = 1.5
	Hard       aiDifficulty = 2.0
	Impossible aiDifficulty = 3.0
)

// Select the difficulty level here
var difficulty = Hard

type gameState int

// Enum for the different game states
const (
	start   gameState = iota
	play    gameState = iota
	restart gameState = iota
)

var state = start

// Score letters, goes up to 9
var numbers = [][]byte{
	{ // 0
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{ // 1
		1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		1, 1, 1,
	},
	{ // 2
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{ // 3
		1, 1, 1,
		0, 0, 1,
		0, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	{ // 4
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		0, 0, 1,
	},
	{ // 5
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	{ // 6
		1, 0, 0,
		1, 0, 0,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{ // 7
		1, 1, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
	},
	{ // 8
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{ // 9
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	}}

// Color struct for the rgba color scheme
type color struct {
	r, g, b, a byte
}

// Helper function for setting the default color to white
func defaultColor() color {
	return color{255, 255, 255, 0}
}

// Position struct which holds the x and y coordinates
type position struct {
	x, y float32
}

// Ball struct which holds the information related to the ball
type ball struct {
	position
	radius int
	xv     float32
	yv     float32
	color  color
}

// Helper function for getting the center of the screen
func centerBall() position {
	return position{float32(windowW) / 2, float32(windowH) / 2}
}

// Used for drawing the ball to the screen
func (ball *ball) draw(pixels []byte) {
	// Start drawing from the top down
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			// Only draw for the radius of the ball
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x)+x, int(ball.y)+y, ball.color, pixels)
			}
		}
	}
}

// Used to determine the position of the ball and handle collisions with the paddle
func (ball *ball) update(score *score, leftPaddle *paddle, rightPaddle *paddle, pixels []byte) {
	ball.x += ball.xv
	ball.y += ball.yv

	// checks if the ball hits the top of the screen, if so bounce back
	if int(ball.y)-ball.radius < 0 {
		ball.yv = -ball.yv
	}
	// checks if the ball hits the bottom of the screen, if so bounce back
	if int(ball.y)+ball.radius > windowH {
		ball.yv = -ball.yv
	}

	if ball.x < 0 { // If ball goes off the screen to the left, player 2 scores
		score.scoreCount2++
		if score.scoreCount2 == 9 {
			state = restart
		} else {
			// Pauses game after player scores
			state = start
		}
		ball.position = centerBall()
	} else if int(ball.x) > windowW { // If ball goes off the screen to the right, player 1 scores
		score.scoreCount1++
		if score.scoreCount1 == 9 {
			state = restart
		} else {
			// Pauses game after player scores
			state = start
		}
		ball.position = centerBall()
	}

	// Ball collision with paddle
	// If the ball collides with the left paddle
	if int(ball.x)-ball.radius < int(leftPaddle.x)+leftPaddle.w/2 {
		if int(ball.y) > int(leftPaddle.y)-leftPaddle.h/2 && int(ball.y) < int(leftPaddle.y)+leftPaddle.h/2 {
			ball.xv = -ball.xv
			// Fixes the buggy collision that happens sometimes due to the ball going inside the paddle
			ball.x = leftPaddle.x + float32(leftPaddle.w)/2 + float32(ball.radius)
		}
	}
	// If the ball collides with the right paddle
	if int(ball.x)+ball.radius > int(rightPaddle.x)-rightPaddle.w/2 {
		if int(ball.y) > int(rightPaddle.y)-rightPaddle.h/2 && int(ball.y) < int(rightPaddle.y)+rightPaddle.h/2 {
			ball.xv = -ball.xv
			// Fixes the buggy collision that happens sometimes due to the ball going inside the paddle
			ball.x = rightPaddle.x - float32(rightPaddle.w)/2 - float32(ball.radius)
		}
	}
}

// Paddle struct which holds the information related to a paddle
type paddle struct {
	position
	w     int
	h     int
	speed float32
	color color
}

// Helper function for setting the position of a paddle
func (paddle *paddle) setPosition(position position) {
	paddle.position = position
}

// Used to determine the direction and speed of the paddle via keyboard input
func (paddle *paddle) update(keyState []uint8) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		paddle.y -= paddle.speed
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += paddle.speed
	}
}

// AI logic for the opposing player, difficulty adjusts the speed in which the AI follows the ball
func (paddle *paddle) aiUpdate(ball *ball) {
	if paddle.y < ball.y {
		paddle.y += float32(difficulty)
	} else {
		paddle.y -= float32(difficulty)
	}
}

// Draws the paddle to the screen
func (paddle *paddle) draw(pixels []byte) {
	// Draws the paddle with the position in the middle
	startX := int(paddle.x) - paddle.w/2
	startY := int(paddle.y) - paddle.h/2

	// Draws the paddle from top down
	for y := 0; y < paddle.h; y++ {
		for x := 0; x < paddle.w; x++ {
			setPixel(startX+x, startY+y, paddle.color, pixels)
		}
	}
}

// Score struct which holds the score of both players
type score struct {
	scoreCount1 int
	scoreCount2 int
}

// Draws the score onto the screen at the specified location
func drawScore(pos position, color color, num int, pixels []byte) {
	size := 10
	startX := int(pos.x) - int(size*3)/2
	startY := int(pos.y) - int(size*5)/2

	for i, v := range numbers[num] {
		// Draws a square at the given position if the value is a 1
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, color, pixels)
				}
			}
		}
		// Once 3 squares are drawn for the row, go down to the next row
		startX += size
		if (i+1)%3 == 0 {
			// Moves Y down one square
			startY += size
			// Puts X back to the start
			startX -= size * 3
		}
	}
}

// Helper function for updating the score during the game
func updateScore(scoreCount *score, leftPaddle *paddle, rightPaddle *paddle, pixels []byte) {
	drawScore(position{float32(windowW/2) + 100, 35}, leftPaddle.color, scoreCount.scoreCount2, pixels)
	drawScore(position{float32(windowW/2) - 100, 35}, rightPaddle.color, scoreCount.scoreCount1, pixels)
}

// Helper function for drawing a line to the middle of the screen
func drawLine(pixels []byte) {
	mid := windowW / 2

	for y := mid; y < mid+10; y++ {
		for x := 0; x < windowH; x++ {
			setPixel(y, x, defaultColor(), pixels)
			x += 5
		}
	}
}

// Helper function for clearing the pixel array
func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

// Draws a pixel at the specified location
func setPixel(x, y int, c color, pixels []byte) {
	// Multiply by the width of the screen to move up and down, add by the x to change the position left and right
	index := (y*windowW + x) * 4

	// Makes sure that the pixels stay on the screen and do not go out of bounds
	if index < len(pixels) && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
		pixels[index+3] = c.a
	}
}

func main() {
	// Creates a basic window
	window, err := sdl.CreateWindow("Pong", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(windowW), int32(windowH), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Closes window and clears up all the resources it was using
	defer window.Destroy()

	// Allows us to draw things on our window
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	// Creates a texture for our things that we draw (PIXELFORMAT_ABGR8888 is the standard RGBA scheme)
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(windowW), int32(windowH))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	// Times by 4 because we need 4 bytes for each pixel (RGBA)
	pixels := make([]byte, windowW*windowH*4)

	// Creates the player paddles
	player1 := paddle{position{50, float32(windowH) / 2}, 15, 100, 5, defaultColor()}
	player2 := paddle{position{float32(windowW) - 50, float32(windowH) / 2}, 15, 100, 5, defaultColor()}

	// Creates the ball
	ball := ball{centerBall(), 20, 3, 3, defaultColor()}

	// Creates the score counter
	scoreCounter := score{0, 0}

	// Gets the state of which key is being pressed
	keyState := sdl.GetKeyboardState()

	for {
		event := sdl.PollEvent()

		// Required or else the sdl window does not stay open
		// Exiting the window will cause the program to end
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		}

		if state == play {
			// Default playing state
			player1.update(keyState)
			player2.aiUpdate(&ball)
			ball.update(&scoreCounter, &player1, &player2, pixels)
		} else if state == start {
			// After someone scores, game pauses and can be un-paused with spacebar
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				state = play
			}
		} else if state == restart {
			// Once the game ends, option to continue playing by pressing spacebar
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				scoreCounter.scoreCount1 = 0
				scoreCounter.scoreCount2 = 0
				player1.setPosition(position{50, float32(windowH) / 2})
				player2.setPosition(position{float32(windowW) - 50, float32(windowH) / 2})
				state = start
			}
		}

		clear(pixels)

		// Drawing the pixels to the window
		player1.draw(pixels)
		player2.draw(pixels)
		ball.draw(pixels)
		drawLine(pixels)

		// Updating the window
		updateScore(&scoreCounter, &player1, &player2, pixels)
		// The pitch is the width of the screen times how many bytes per pixel
		tex.Update(nil, pixels, windowW*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		sdl.Delay(5)
	}
}
