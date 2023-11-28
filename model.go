package main

type model struct {
	prefs          preferences
	minefield      [][]cell
	cursorX        int
	cursorY        int
	isGameOver     bool
	secondsElapsed int
}

type preferences struct {
	width         int
	height        int
	numberOfMines int
	isDebug       bool
	showHelp      bool
}

type cell struct {
	isMine     bool
	isFlagged  bool
	isRevealed bool
}

type point struct {
	x int
	y int
}
