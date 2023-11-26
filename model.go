package main

type model struct {
	prefs      preferences
	minefield  [][]cell
	cursorX    int
	cursorY    int
	isGameOver bool
}

type preferences struct {
	width         int
	height        int
	numberOfMines int
	isDebug       bool
}

type cell struct {
	isMine     bool
	isFlagged  bool
	isRevealed bool
	isUnknown  bool
}

type point struct {
	x int
	y int
}
