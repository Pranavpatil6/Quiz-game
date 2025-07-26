package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const totalQuestions = 10

type Question struct {
	question string
	answer   string
}

func main() {
	filename, timeLimit := readArguments()
	f, err := openFile(filename)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	questions, err := readCSV(f)
	if err != nil {
		fmt.Println("Error reading questions:", err)
		return
	}

	if len(questions) < totalQuestions {
		fmt.Printf("Not enough questions in the file. Found %d, need %d\n", len(questions), totalQuestions)
		return
	}

	fmt.Printf("Loaded %d questions. You have %d seconds per question.\n", len(questions), timeLimit)
	score, err := askQuestion(questions, timeLimit)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Your Score %d/%d\n", score, totalQuestions)
}

func readArguments() (string, int) {
	filename := flag.String("filename", "problem.csv", "CSV file containing quiz questions")
	timeLimit := flag.Int("limit", 10, "Time limit per question (seconds)")
	flag.Parse()
	return *filename, *timeLimit
}

func readCSV(f io.Reader) ([]Question, error) {
	allQuestions, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(allQuestions) == 0 {
		return nil, fmt.Errorf("no questions found in CSV file")
	}

	var data []Question
	for _, line := range allQuestions {
		if len(line) < 2 {
			continue
		}
		ques := Question{
			question: line[0],
			answer:   line[1],
		}
		data = append(data, ques)
	}
	return data, nil
}

func openFile(filename string) (io.Reader, error) {
	return os.Open(filename)
}

func getInput(input chan string) {
	in := bufio.NewReader(os.Stdin)
	for {
		result, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input <- result
	}
}

func askQuestion(questions []Question, timeLimit int) (int, error) {
	totalScore := 0
	done := make(chan string)

	go getInput(done)

	for i := 0; i < totalQuestions; i++ {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)

		ans, err := eachQuestion(questions[i].question, questions[i].answer, timer.C, done)

		timer.Stop()

		if err != nil {
			if ans == -1 {
				fmt.Println("\nTime's up!")
				break
			}
			fmt.Println(err)
		}
		totalScore += ans
	}
	return totalScore, nil
}

func eachQuestion(quest string, answer string, timer <-chan time.Time, done <-chan string) (int, error) {
	fmt.Printf("Question: %s = ", quest)

	for {
		select {
		case <-timer:
			return -1, fmt.Errorf("timeout")
		case ans := <-done:
			trimmedAns := strings.TrimSpace(strings.ToLower(ans))
			trimmedAnswer := strings.ToLower(answer)

			if trimmedAns == trimmedAnswer {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	}
}
