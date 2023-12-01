package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const (
	DEFAULT_WIDTH  = 30
	DEFAULT_HEIGHT = 30
	DEFAULT_MINES  = 99
)

var wFlag, hFlag, numMinesFlag int
var shouldUseAscii bool

func main() {
	flag.IntVar(&wFlag, "w", DEFAULT_WIDTH, "minefield width")
	flag.IntVar(&hFlag, "h", DEFAULT_HEIGHT, "minefield height")
	flag.IntVar(&numMinesFlag, "n", DEFAULT_MINES, "number of mines")
	flag.BoolVar(&shouldUseAscii, "a", false, "use ascii characters")

	flag.Parse()

	m := initialModel(wFlag, hFlag, numMinesFlag, shouldUseAscii)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel(width int, height int, numMines int, ascii bool) model {
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
	numMines = min(numMines, width*height)
	for i := 0; i < numMines && i < width*height; i++ {
		p := positions.pop()
		minefield[p.y][p.x].isMine = true
	}

	return model{
		prefs: preferences{
			width:         width,
			height:        height,
			numberOfMines: numMines,
			showHelp:      true,
			ascii:         ascii,
		},
		minefield: minefield,
		cursorX:   width/2 - 1,
		cursorY:   height/2 - 1,
	}
}

type tickMsg struct{}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return tickMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	cursorMine := &m.minefield[m.cursorY][m.cursorX]

	switch msg := msg.(type) {
	case tickMsg:
		if m.isGameOver {
			break
		}
		m.secondsElapsed++
		return m, tick()
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
			ascii := m.prefs.ascii
			m = initialModel(wFlag, hFlag, numMinesFlag, ascii)
			m.prefs.isDebug = isDebug
			m.cursorY, m.cursorX = y, x
			m.prefs.showHelp = showHelp
			break
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
			if cursorMine.isRevealed {
				sweep(m.cursorX, m.cursorY, &m, true, make(set[point]))
			} else {
				cursorMine.isFlagged = !cursorMine.isFlagged
			}
		case "d":
			m.prefs.isDebug = !m.prefs.isDebug
		case "?":
			m.prefs.showHelp = !m.prefs.showHelp
		case "a":
			m.prefs.ascii = !m.prefs.ascii
		}
	}

	return m, nil
}

func (m model) View() string {
	var sb strings.Builder
	writeHeader(&sb, m)
	sb.WriteString("\n\n")
	if m.prefs.ascii {
		writeAsciiMinefield(&sb, m)
	} else {
		writeMinefield(&sb, m)
	}
	sb.WriteString("\n\n")
	writeHelp(&sb, m)
	return sb.String()
}

func writeHeader(sb *strings.Builder, m model) {
	if m.isGameOver {
		sb.WriteString("Game Over! ")
		if checkDidWin(m) {
			sb.WriteString("You WON!!")
		} else {
			sb.WriteString("You lost...")
		}
	} else {
		sb.WriteString("...go sweep...")
		sb.WriteString(fmt.Sprintf("\n%v mines left", minesLeft(m)))
	}
	sb.WriteString(fmt.Sprintf("\n%v seconds elapsed", m.secondsElapsed))
}

func writeMinefield(sb *strings.Builder, m model) {
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
}

func writeAsciiMinefield(sb *strings.Builder, m model) {
	strs := make([][]string, m.prefs.height)
	for y, row := range m.minefield {
		strs[y] = make([]string, m.prefs.width)
		for x, mine := range row {
			switch {
			case x == m.cursorX && y == m.cursorY:
				strs[y][x] = "*"
			case (m.isGameOver || m.prefs.isDebug) && mine.isMine:
				strs[y][x] = "B"
			case mine.isRevealed:
				strs[y][x] = asciiViewForMineAtPosition(x, y, m)
			case mine.isFlagged:
				strs[y][x] = "F"
			default:
				strs[y][x] = " "
			}
		}
	}
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderRow(true).
		BorderColumn(true).
		Rows(strs...).
		StyleFunc(func(row, col int) lipgloss.Style {
			var fg lipgloss.TerminalColor = lipgloss.NoColor{}
			var bg lipgloss.TerminalColor = lipgloss.NoColor{}

			switch strs[row-1][col] {
			case "0":
				fg = lipgloss.Color("#292929")
			case "1":
				fg = lipgloss.Color("#74adf2")
			case "2":
				fg = lipgloss.Color("#00FF00")
			case "3":
				fg = lipgloss.Color("#FF0000")
			case "4":
				fg = lipgloss.Color("#28706d")
			case "5":
				fg = lipgloss.Color("#b06446")
			case "6":
				fg = lipgloss.Color("#FF0000")
			case "7":
				fg = lipgloss.Color("#8a7101")
			case "8":
				fg = lipgloss.Color("#111")
				bg = lipgloss.Color("#bfbfbf")
			case "*":
				fg = lipgloss.Color("#FF33FF")
			case "F":
				bg = lipgloss.Color("#ffee00")
				fg = lipgloss.Color("#111")
			case "B":
				bg = lipgloss.Color("#FF0000")
				fg = lipgloss.Color("#111")
			case " ":
				bg = lipgloss.Color("#bfbfbf")
			}

			return lipgloss.NewStyle().
				Foreground(fg).
				Background(bg).
				Padding(0, 1)
		})
	sb.WriteString(t.Render())
}

func writeHelp(sb *strings.Builder, m model) {
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
		sb.WriteString("Press a to toggle ascii view\n")
	}
}

func sweep(x, y int, m *model, userInitiatedSweep bool, swept set[point]) {
	cell := &m.minefield[y][x]

	if cell.isRevealed && userInitiatedSweep {
		adjMines := countAdjacentMines(x, y, *m)
		adjFlags := countAdjacentFlags(x, y, *m)
		if adjFlags >= adjMines {
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

	p := point{x: x, y: y}
	if touching == 0 && !swept.has(p) {
		swept.add(p)
		forEachSurroundingCellDo(x, y, m, func(x, y int, m *model) {
			sweep(x, y, m, false, swept)
		})
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

func asciiViewForMineAtPosition(x, y int, m model) string {
	if m.minefield[y][x].isMine {
		return "B"
	}
	return fmt.Sprint(countAdjacentMines(x, y, m))
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

func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}
