package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type QuizQuestionRepository struct {
	QuizQuestions []*QuizQuestion
}

type QuizQuestion struct {
	Question string
	Answer   string
	Correct  bool
}

func NewQuizQuestion(question string, answer string) *QuizQuestion {
	return &QuizQuestion{
		Question: question,
		Answer:   answer,
		Correct:  false,
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func askQuestion(QuizQuestion *QuizQuestion, Reader *bufio.Reader, answerChannel chan<- string) {

	fmt.Printf("Question: %s", QuizQuestion.Question)
	fmt.Println()
	text, _ := Reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	answerChannel <- text
}

func showSummary(QuizQuestionRepository *QuizQuestionRepository) {
	numCorrectAnswers := 0
	for _, quizQuestion := range QuizQuestionRepository.QuizQuestions {
		if quizQuestion.Correct {
			numCorrectAnswers++
		}
	}

	fmt.Printf("You got %d out of %d correct answers.", numCorrectAnswers, len(QuizQuestionRepository.QuizQuestions))
}

func main() {

	reader := bufio.NewReader(os.Stdin)

	fileLocation := flag.String("file", "questions.csv", "A CSV with questions")
	questionsTimeout := flag.Int("question-timeout", 5, "Time Limit for the Question")

	flag.Parse()

	content, err := ioutil.ReadFile(*fileLocation)

	check(err)

	repository := &QuizQuestionRepository{
		QuizQuestions: make([]*QuizQuestion, 0),
	}

	r := csv.NewReader(strings.NewReader(string(content)))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		quizQuestion := NewQuizQuestion(record[0], record[1])
		repository.QuizQuestions = append(repository.QuizQuestions, quizQuestion)
	}

	answerChannel := make(chan string)
	timer := time.NewTimer(time.Duration(*questionsTimeout) * time.Second)

	for _, quizQuestion := range repository.QuizQuestions {
		go askQuestion(quizQuestion, reader, answerChannel)
		select {
		case <-timer.C:
			fmt.Println("Time is up!")
			showSummary(repository)
			return
		case answer := <-answerChannel:
			if answer == quizQuestion.Answer {
				quizQuestion.Correct = true
			}
		}
	}

	showSummary(repository)
}
