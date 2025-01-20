package letterbox

import (
	"bufio"
	"errors"
	"log"
	"math/bits"
	"os"
	"sort"
	"strings"
)

const (
	L1 EncodedLetter = 1 << iota
	L2
	L3
	L4
	L5
	L6
	L7
	L8
	L9
	L10
	L11
	L12

	PUZZLE_SIZE      = 12
	MIN_WORD_LENGTH  = 3
	MAX_SEARCH_DEPTH = 3
)

type EncodedLetter uint16
type EncodedWord uint16
type ValidationMask uint16

type Word struct {
	// The text representation of the word directly from the list of words
	Text string
	// A binary representation of the word where each letter is represented by a single bit.
	// Duplicate letters do not produce additional information.
	Binary EncodedWord
	// The binary representation of the first letter in the word. Use this to chain words
	// together.
	FirstLetter rune
	// The binary representation of the final letter in the word. Use this to chain words
	// together.
	LastLetter rune
}

func (w *Word) String() string {
	return w.Text
}

type GameBoard struct {
	Board    string
	Mapping  map[rune]EncodedLetter
	WordList []Word
}

func NewGameBoard(input string) *GameBoard {
	if len(input) != PUZZLE_SIZE {
		log.Panicf("Input '%s' is invalid.", input)
	}
	return &GameBoard{
		Board:    input,
		Mapping:  encodingMap(input),
		WordList: make([]Word, 0),
	}
}

// Load the list of value words from a dictionary file.
func (board *GameBoard) LoadWordList(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal("Unable to open dictionary file.")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		if len(word) < MIN_WORD_LENGTH {
			continue
		}
		line, err := board.NewWord(word)
		if err != nil {
			continue
		}
		board.WordList = append(board.WordList, line)
	}
	sort.Slice(board.WordList, func(i, j int) bool {
		return board.WordList[i].Score() > board.WordList[j].Score()
	})
}

func encodingMap(input string) map[rune]EncodedLetter {
	runes := []rune(input)
	if len(runes) != PUZZLE_SIZE {
		log.Panic("Unable to create mapping because the board size is incorrect.")
	}
	return map[rune]EncodedLetter{
		runes[0]:  L1,
		runes[1]:  L2,
		runes[2]:  L3,
		runes[3]:  L4,
		runes[4]:  L5,
		runes[5]:  L6,
		runes[6]:  L7,
		runes[7]:  L8,
		runes[8]:  L9,
		runes[9]:  L10,
		runes[10]: L11,
		runes[11]: L12,
	}
}

// Encode a string to a binary representation
func (board *GameBoard) NewWord(word string) (Word, error) {
	mapping := board.Mapping
	runes := []rune(word)
	encodedWord := Word{
		Text:        word,
		FirstLetter: runes[0],
		LastLetter:  runes[len(runes)-1],
		Binary:      0,
	}
	previousLetter := EncodedLetter(0)
	for i, letter := range runes {
		encodedLetter := mapping[letter]
		if encodedLetter == 0 {
			return encodedWord, errors.New("Unable to encode a letter.")
		}
		if i > 0 && !isValidJump(previousLetter, encodedLetter) {
			return encodedWord, errors.New("Invalid jump in word.")
		}
		previousLetter = encodedLetter
		encodedWord.Binary = encodedWord.Binary | EncodedWord(encodedLetter)
	}
	return encodedWord, nil
}

// Check two encoded letters to confirm if the move from one to the other is valid.
func isValidJump(previousLetter EncodedLetter, proposedLetter EncodedLetter) bool {
	allValidMask := EncodedLetter(0b111_111_111_111)
	var mask ValidationMask
	switch previousLetter {
	case L1, L2, L3:
		mask = ValidationMask(allValidMask ^ L1 ^ L2 ^ L3)
	case L4, L5, L6:
		mask = ValidationMask(allValidMask ^ L4 ^ L5 ^ L6)
	case L7, L8, L9:
		mask = ValidationMask(allValidMask ^ L7 ^ L8 ^ L9)
	case L10, L11, L12:
		mask = ValidationMask(allValidMask ^ L10 ^ L11 ^ L12)
	default:
		return false
	}
	state := uint16(mask & ValidationMask(proposedLetter))
	// the encoded letter combined with the validation mask should have zero overlap causing
	// state to be zero and to have zero ones.
	return bits.OnesCount16(state) == 1
}

func (word *Word) Score() int {
	return bits.OnesCount16(uint16(word.Binary))
}

type Solution []Word

func (solution Solution) String() string {
	result := make([]string, len(solution))
	for i, word := range solution {
		result[i] = word.String()
	}
	return strings.Join(result, ", ")
}

func (solution *Solution) Score() int {
	sum := EncodedWord(0)
	for _, word := range *solution {
		sum = sum | word.Binary
	}
	return bits.OnesCount16(uint16(sum))
}

func (solution *Solution) isSolved() bool {
	return solution.Score() == PUZZLE_SIZE
}

func (board *GameBoard) GenerateSolution() Solution {
	solutions := make([]Solution, len(board.WordList))
	for i, v := range board.WordList {
		solution := Solution([]Word{v})
		// if the solution is already found, we can stop early.
		if solution.isSolved() {
			return solution
		}
		solutions[i] = Solution([]Word{v})
	}
	for range MAX_SEARCH_DEPTH {
		nextSolutions := make([]Solution, 0)
		for _, solution := range solutions {
			for _, v := range board.WordList {
				previousWord := solution[len(solution)-1]
				if previousWord.LastLetter != v.FirstLetter {
					continue
				}
				score := solution.Score()
				newSolution := append(solution, v)
				if newSolution.isSolved() {
					return newSolution
				}
				// this takes care of duplicate words and words that embed other words.
				if newSolution.Score() > score {
					nextSolutions = append(nextSolutions, newSolution)
				}
			}
		}
		sort.Slice(nextSolutions, func(i, j int) bool {
			return nextSolutions[i].Score() > nextSolutions[j].Score()
		})
		solutions = nextSolutions
	}
	log.Fatalf("Looped %d times and did not find a solution", MAX_SEARCH_DEPTH)
	return nil
}
