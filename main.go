package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	DEFAULT_WIDTH  = 30
	DEFAULT_HEIGHT = 30
	DEFAULT_MINES  = 99
)

func main() {
	m := initialModel(DEFAULT_WIDTH, DEFAULT_HEIGHT, DEFAULT_MINES)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel(width int, height int, numMines int) model {
	positions := make(stack[point], 0, width*height)
	minefield := make([][]cell, height)

	for y := range minefield {
		minefield[y] = make([]cell, width)
		for x := range minefield[y] {
			minefield[y][x] = cell{}
			positions.push(point{x: x, y: y})
		}
	}

	// TODO instantiate the mines after the first sweep to make sure first
	// click never hits a mine
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	for i := 0; i < numMines; i++ {
		p := positions.pop()
		minefield[p.x][p.y].isMine = true
	}

	return model{
		prefs: preferences{
			width:         width,
			height:        height,
			numberOfMines: numMines,
			showHelp:      true,
		},
		minefield: minefield,
		cursorX:   width/2 - 1,
		cursorY:   height/2 - 1,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	cursorMine := &m.minefield[m.cursorY][m.cursorX]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.cursorY--
			if m.cursorY < 0 {
				m.cursorY = m.prefs.height - 1
			}
		case "down", "j":
			m.cursorY++
			if m.cursorY > m.prefs.height-1 {
				m.cursorY = 0
			}
		case "left", "h":
			m.cursorX--
			if m.cursorX < 0 {
				m.cursorX = m.prefs.width - 1
			}
		case "right", "l":
			m.cursorX++
			if m.cursorX > m.prefs.width-1 {
				m.cursorX = 0
			}
		case "r":
			isDebug := m.prefs.isDebug
			y, x := m.cursorY, m.cursorX
			showHelp := m.prefs.showHelp
			m = initialModel(DEFAULT_WIDTH, DEFAULT_HEIGHT, DEFAULT_MINES)
			m.prefs.isDebug = isDebug
			m.cursorY, m.cursorX = y, x
			m.prefs.showHelp = showHelp
		case "enter", " ":
			if m.isGameOver {
				break
			}
			sweep(m.cursorX, m.cursorY, &m, true, make(set[point]))
			if checkDidWin(m) {
				m.isGameOver = true
			}
		case "f":
			if m.isGameOver {
				break
			}
			cursorMine.isFlagged = !cursorMine.isFlagged
		case "d":
			m.prefs.isDebug = !m.prefs.isDebug
		case "?":
			m.prefs.showHelp = !m.prefs.showHelp
		}
	}

	return m, nil
}

func (m model) View() string {
	var sb strings.Builder
	if m.isGameOver {
		sb.WriteString("Game Over! ")
		if checkDidWin(m) {
			sb.WriteString("You WON!!")
		} else {
			sb.WriteString("You lost...")
		}
	} else {
		sb.WriteString("...go sweep...")
		sb.WriteString(fmt.Sprintf(" %v mines left", minesLeft(m)))
	}
	sb.WriteString("\n\n")

	for y, row := range m.minefield {
		for x, mine := range row {
			switch {
			case x == m.cursorX && y == m.cursorY:
				sb.WriteString("🔳")
			case (m.isGameOver || m.prefs.isDebug) && mine.isMine:
				sb.WriteString("💣")
			case mine.isRevealed:
				sb.WriteString(viewForMineAtPosition(x, y, m))
			case mine.isFlagged:
				sb.WriteString("🟨")
			default:
				sb.WriteString("⬜️")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if m.prefs.showHelp {
		if !m.isGameOver {
			sb.WriteString("Press h/j/k/l or ←↓↑→ to move\n")
			sb.WriteString("Press enter or space to sweep\n")
			sb.WriteString("Press f to toggle flag.\n")
			sb.WriteString("Press d to toggle debug.\n")
		}
		sb.WriteString("Press q to quit.\n")
		sb.WriteString("Press r to start a new game.\n")
		sb.WriteString("Press ? to toggle help text\n")
	}
	return sb.String()
}

func sweep(x, y int, m *model, userInitiatedSweep bool, swept set[point]) {
	cell := &m.minefield[y][x]

	if cell.isRevealed && userInitiatedSweep {
		adjMines := countAdjacentMines(x, y, *m)
		adjFlags := countAdjacentFlags(x, y, *m)
		if adjFlags == adjMines {
			autoSweep(x, y, m)
		}
		return
	}

	if cell.isMine {
		if userInitiatedSweep {
			m.isGameOver = true
		}
		return
	}

	touching := countAdjacentMines(x, y, *m)

	w := m.prefs.width
	h := m.prefs.height
	p := point{x: x, y: y}
	if touching == 0 && !swept.has(p) {
		swept.add(p)
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if (dx == 0 && dy == 0) ||
					x+dx < 0 ||
					x+dx > w-1 ||
					y+dy < 0 ||
					y+dy > h-1 {
					continue
				}

				// if (dx*dy == 1 || dx*dy == -1) &&
				// 	countAdjacentMines(x+dx, y+dy, *m) == 0 {
				// 	continue
				// }

				sweep(x+dx, y+dy, m, false, swept)
			}
		}

	}

	cell.isRevealed = true
}

func minesLeft(m model) int {
	flags := 0
	for y := range m.minefield {
		for _, mine := range m.minefield[y] {
			if mine.isFlagged && !mine.isRevealed {
				flags++
			}
		}
	}
	return m.prefs.numberOfMines - flags
}

func checkDidWin(m model) bool {
	for y := range m.minefield {
		for _, mine := range m.minefield[y] {
			if !mine.isMine && !mine.isRevealed {
				return false
			}
		}
	}
	return true
}

func viewForMineAtPosition(x, y int, m model) string {
	if m.minefield[y][x].isMine {
		return "💣"
	}
	touching := countAdjacentMines(x, y, m)
	numViewMap := map[int]string{
		0: "⬛️",
		1: "1️⃣",
		2: "2️⃣",
		3: "3️⃣",
		4: "4️⃣",
		5: "5️⃣",
		6: "6️⃣",
		7: "7️⃣",
		8: "8️⃣",
	}
	return numViewMap[touching]
}

func forEachSurroundingCellDo(x, y int, m *model, do func(x, y int, m *model)) {
	w := m.prefs.width
	h := m.prefs.height
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if (dx == 0 && dy == 0) || x+dx < 0 || x+dx > w-1 || y+dy < 0 || y+dy > h-1 {
				continue
			}
			do(x+dx, y+dy, m)
		}
	}
}

func autoSweep(x, y int, m *model) {
	forEachSurroundingCellDo(x, y, m, func(x, y int, m *model) {
		cell := m.minefield[y][x]
		if !cell.isRevealed && !cell.isFlagged {
			sweep(x, y, m, true, make(set[point]))
		}
	})
}

func countAdjacentFlags(x, y int, m model) int {
	adj := 0
	forEachSurroundingCellDo(x, y, &m, func(x, y int, m *model) {
		if m.minefield[y][x].isFlagged {
			adj++
		}
	})
	return adj
}

func countAdjacentMines(x, y int, m model) int {
	adj := 0
	forEachSurroundingCellDo(x, y, &m, func(x, y int, m *model) {
		if m.minefield[y][x].isMine {
			adj++
		}
	})
	return adj
}
