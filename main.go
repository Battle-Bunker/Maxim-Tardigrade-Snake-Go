package main

import (
	"log"
	"math"
)

const (
	UP    = "up"
	DOWN  = "down"
	LEFT  = "left"
	RIGHT = "right"
)

func ManhattanDistance(p1, p2 Coord) int {
	return int(math.Abs(float64(p1.X-p2.X)) + math.Abs(float64(p1.Y-p2.Y)))
}

func FindClosestFood(start Coord, foodPoints []Coord) Coord {
	if len(foodPoints) == 0 {
		return Coord{} // Returning a default Point if the foodPoints list is empty
	}

	closest := foodPoints[0]
	minDistance := ManhattanDistance(start, closest)

	for _, food := range foodPoints[1:] {
		distance := ManhattanDistance(start, food)
		if distance < minDistance {
			closest = food
			minDistance = distance
		}
	}

	return closest
}

func info() BattlesnakeInfoResponse {
	log.Println("INFO")

	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "crazycat911", // TODO: Your Battlesnake username
		Color:      "#002d62",     // TODO: Choose color
		Head:       "comet",       // TODO: Choose head
		Tail:       "mlh-gene",    // TODO: Choose tail
	}
}

func start(state GameState) {
	log.Println("GAME START")
}

func end(state GameState) {
	log.Printf("GAME OVER\n\n")
}

func detectDanger(state GameState) map[Coord]bool {
	mySnake := state.You
	snakes := state.Board.Snakes
	dangerZones := make(map[Coord]bool)

	for _, snake := range snakes {
		for i, bodypart := range snake.Body {
			dangerZones[bodypart] = true
			if i == len(snake.Body)-1 && snake.Health == 100 {
				dangerZones[bodypart] = true
			}
		}

		if snake.ID != mySnake.ID && snake.Length >= mySnake.Length {
			dangerZones[Coord{X: snake.Head.X + 1, Y: snake.Head.Y}] = true
			dangerZones[Coord{X: snake.Head.X - 1, Y: snake.Head.Y}] = true
			dangerZones[Coord{X: snake.Head.X, Y: snake.Head.Y + 1}] = true
			dangerZones[Coord{X: snake.Head.X, Y: snake.Head.Y - 1}] = true
		}
	}

	for _, location := range getNeighbors(state.You.Body[0]) {
		if location == state.You.Body[1] {
			dangerZones[location] = true
		}
	}

	return dangerZones
}

func isCoordInBoard(c Coord, board Board) bool {
	return c.X >= 0 && c.Y >= 0 && c.X < board.Width && c.Y < board.Height
}

func getNeighbors(c Coord) []Coord {
	return []Coord{
		{X: c.X + 1, Y: c.Y},
		{X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1},
		{X: c.X, Y: c.Y - 1},
	}
}

func floodFill(start Coord, board Board, dangerZones map[Coord]bool) int {
	queue := []Coord{start}
	visited := map[Coord]bool{start: true}
	area := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		area++

		for _, neighbor := range getNeighbors(current) {
			if isCoordInBoard(neighbor, board) && !visited[neighbor] && !dangerZones[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return area
}

func move(state GameState) BattlesnakeMoveResponse {
	myHead := state.You.Body[0] // Coordinates of your head

	// Initialize safe moves
	isMoveSafe := map[string]bool{
		UP:    true,
		DOWN:  true,
		LEFT:  true,
		RIGHT: true,
	}

	// Check wall collisions
	if myHead.X == 0 {
		isMoveSafe[LEFT] = false
	} else if myHead.X == state.Board.Width-1 {
		isMoveSafe[RIGHT] = false
	}

	if myHead.Y == 0 {
		isMoveSafe[DOWN] = false
	} else if myHead.Y == state.Board.Height-1 {
		isMoveSafe[UP] = false
	}

	dangerZones := detectDanger(state)

	// Check self and other snakes using the dangerZones map
	directions := map[string]Coord{
		UP:    {X: myHead.X, Y: myHead.Y + 1},
		DOWN:  {X: myHead.X, Y: myHead.Y - 1},
		LEFT:  {X: myHead.X - 1, Y: myHead.Y},
		RIGHT: {X: myHead.X + 1, Y: myHead.Y},
	}

	for direction, newHead := range directions {
		if isMoveSafe[direction] && (dangerZones[newHead] || floodFill(newHead, state.Board, dangerZones) < state.You.Length) {
			isMoveSafe[direction] = false
		}
	}

	// Determine safe moves
	safeMoves := []string{}
	for move, safe := range isMoveSafe {
		if safe {
			safeMoves = append(safeMoves, move)
		}
	}

	// No safe moves
	if len(safeMoves) == 0 {
		log.Printf("MOVE %d: No safe moves detected :( Moving up\n", state.Turn)
		return BattlesnakeMoveResponse{Move: UP}
	}

	food := state.Board.Food
	if len(food) > 0 {
		closestFood := FindClosestFood(myHead, food)
		dx := closestFood.X - myHead.X
		dy := closestFood.Y - myHead.Y

		var moveTowardsFood string

		if dx > 0 && isMoveSafe[RIGHT] {
			moveTowardsFood = RIGHT
		} else if dx < 0 && isMoveSafe[LEFT] {
			moveTowardsFood = LEFT
		} else if dy > 0 && isMoveSafe[UP] {
			moveTowardsFood = UP
		} else if dy < 0 && isMoveSafe[DOWN] {
			moveTowardsFood = DOWN
		}

		if moveTowardsFood != "" {
			log.Printf("MOVE %d: Moving towards food %s\n", state.Turn, moveTowardsFood)
			return BattlesnakeMoveResponse{Move: moveTowardsFood}
		}
	}

	// Choose the best move based on flood fill algorithm
	bestMove := safeMoves[0]
	maxArea := -1
	for _, move := range safeMoves {
		newHead := directions[move]
		area := floodFill(newHead, state.Board, dangerZones)
		if area > maxArea {
			maxArea = area
			bestMove = move
		}
	}

	log.Printf("MOVE %d: %s\n", state.Turn, bestMove)
	return BattlesnakeMoveResponse{Move: bestMove}
}

func main() {
	RunServer()
}
