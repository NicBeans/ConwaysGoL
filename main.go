package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"log"
	"sync"
)

//Underpopulation: Any live cell with fewer than two live neighbors dies.
//Next Generation: Any live cell with two or three live neighbors lives on to the next generation.
//Overpopulation: Any live cell with more than three live neighbors dies.
//Reproduction: Any dead cell with exactly three live neighbors becomes a live cell.

//aim to throw all cells in a go routine (individually), put whole lot in wait group and then update arrays to FE
//each thread shoul eval tile, call update to flip colour if needed

//https://ebitengine.org/en/documents/cheatsheet.html

const (
	screenWidth  = 600
	screenHeight = 600
	gridSize     = 10
	cellSize     = 60
)

type Cell struct {
	Color color.NRGBA
	X     int
	Y     int
}

type Game struct {
	//inserting a 6x6 grid
	grid               [gridSize][gridSize]Cell
	nextGrid           [gridSize][gridSize]Cell
	MouseButtonPressed bool
	isPlaying          bool
}

func (g *Game) startIterations() error {
	for x := 0; x < gridSize; x++ {
		for y := 0; y < gridSize; y++ {

			if g.evaluateUnderpopulationRule(x, y) {
				g.grid[x][y].Color = color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF} // Change color back to grey
			}
			//if g.evaluateSurvivalRule(x, y) {
			//	g.grid[x][y].Color = color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF} // Change color back to grey
			//}
			//if g.evaluateOverpopulationRule(x, y) {
			//	g.grid[x][y].Color = color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF} // Change color back to grey
			//}
			//if g.evaluateResurrectionRule(x, y) {
			//	g.grid[x][y].Color = color.NRGBA{R: 255, G: 255, B: 0, A: 255} // Change color to yellow
			//}
		}
	}
	return nil
}

func (g *Game) evaluateUnderpopulationRule(x, y int) bool {
	liveNeighbors := g.countLiveNeighbors(x, y)

	if g.grid[x][y].Color == (color.NRGBA{R: 255, G: 255, B: 0, A: 255}) { // assuming yellow color indicates a live cell
		if liveNeighbors < 2 {
			return true // cell changes
		}
	}
	return false // cell stays the same
}

func (g *Game) evaluateSurvivalRule(x, y int) bool {
	liveNeighbors := g.countLiveNeighbors(x, y)

	if g.grid[x][y].Color == (color.NRGBA{R: 255, G: 255, B: 0, A: 255}) { // assuming yellow color indicates a live cell
		if liveNeighbors == 2 || liveNeighbors == 3 {
			return false // cell survives
		}
	}
	return true // cell changes
}

func (g *Game) evaluateOverpopulationRule(x, y int) bool {
	liveNeighbors := g.countLiveNeighbors(x, y)

	if g.grid[x][y].Color == (color.NRGBA{R: 255, G: 255, B: 0, A: 255}) { // assuming grey yellow indicates a live cell
		if liveNeighbors > 3 {
			return true // cell changes (resurrects)
		}
	}
	return false // cell survives
}

func (g *Game) evaluateResurrectionRule(x, y int) bool {
	liveNeighbors := g.countLiveNeighbors(x, y)

	if g.grid[x][y].Color == (color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF}) { // assuming yellow color indicates a live cell
		if liveNeighbors == 3 {
			return true // cell changes (resurrects)
		}
	}
	return false // cell stays dead
}

// helper function to count live neighbors
func (g *Game) countLiveNeighbors(x, y int) int {
	liveNeighbors := 0

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			newX, newY := x+i, y+j

			if newX >= 0 && newX < gridSize && newY >= 0 && newY < gridSize { // check boundaries
				if !(i == 0 && j == 0) { // skip cell itself
					if g.grid[newX][newY].Color == (color.NRGBA{R: 255, G: 255, B: 0, A: 255}) {
						liveNeighbors++
					}
				}
			}
		}
	}
	fmt.Println(liveNeighbors)
	return liveNeighbors
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		err := g.startIterations()
		if err != nil {
			return err
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		xPos, yPos := ebiten.CursorPosition()
		// Translating the cursor's position to grid coordinates
		cellX := xPos / cellSize
		cellY := yPos / cellSize
		if cellX < gridSize && cellY < gridSize {
			// Interact with the cell
			//If cell is grey, set it to yellow.
			if g.grid[cellX][cellY].Color == (color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF}) {
				g.grid[cellX][cellY].Color = color.NRGBA{255, 255, 0, 255} // Change color to yellow
			} else if g.grid[cellX][cellY].Color == (color.NRGBA{255, 255, 0, 255}) {
				g.grid[cellX][cellY].Color = color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF} // Change color back to grey
			}
		}
	}
	if g.isPlaying {
		var cellGroup sync.WaitGroup
		for x := 0; x < gridSize; x++ {
			for y := 0; y < gridSize; y++ {
				cellGroup.Add(1)
				go func(x, y int) {
					defer cellGroup.Done()
					//liveNeighbors := g.countLiveNeighbors(x, y)
					// Apply the Game of Life rules
					//if g.grid[x][y].Color == (color.NRGBA{R: 255, G: 255, B: 0, A: 255}) { // assuming yellow color indicates a live cell
					//	// Rule 1 or Rule 3
					//	if liveNeighbors < 2 || liveNeighbors > 3 {
					//		g.nextGrid[x][y].Color = color.NRGBA{B: 0xB0, A: 0xFF} // cell dies
					//	}
					//	// Rule 2
					//	if liveNeighbors == 2 || liveNeighbors == 3 {
					//		g.nextGrid[x][y].Color = color.NRGBA{R: 255, G: 255, B: 0, A: 255} // cell lives
					//	}
					//} else {
					//	// Rule 4
					//	if liveNeighbors == 3 {
					//		g.nextGrid[x][y].Color = color.NRGBA{R: 255, G: 255, B: 0, A: 255} // cell births
					//	}
					//}
				}(x, y)
			}
		}
		cellGroup.Wait()
		// Swap the grids
		g.grid, g.nextGrid = g.nextGrid, g.grid
	}
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	d := &font.Drawer{
		Dst:  screen,
		Src:  image.Black,
		Face: basicfont.Face7x13,
	}
	for i := range g.grid {
		for j := range g.grid[i] {
			x := float32(i * cellSize)
			y := float32(j * cellSize)
			// Drawing the grid cells with the color from the grid
			vector.DrawFilledRect(screen, x, y, cellSize-2, cellSize-2, g.grid[i][j].Color, false)

			// Draw the text "Hello" at this position
			xStr := i
			yStr := j
			d.Dot = fixed.Point26_6{X: fixed.Int26_6((x + 4) * 64), Y: fixed.Int26_6((y + (cellSize / 2)) * 64)}
			d.DrawString(fmt.Sprintf("%d %d", xStr, yStr))
			fmt.Println(fmt.Sprintf("%d %d", xStr, yStr))
		}
	}
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Return the actual window size (must be positive numbers)
	return screenWidth, screenHeight
}
func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("press space to iterate")
	game := &Game{}
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			game.grid[i][j] = Cell{
				X:     i * cellSize,
				Y:     j * cellSize,
				Color: color.NRGBA{0xB0, 0xB0, 0xB0, 0xFF},
			}
		}
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
