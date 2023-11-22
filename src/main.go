package main

import (
	"os"
	"fmt"
	"errors"
	"regexp"
	"strings"
	"io/ioutil"
	"os/exec"
	"net/url"
	"net/http"
	"net/http/cookiejar"
	"golang.org/x/net/html"
)

type AdventOfCode struct {
	year int
	day int
}

type AdventOfCodeProblem struct {
	adventOfCode AdventOfCode
	description string
	input string

	answers []string
}

func AdventOfCodeRootURL() string {
	return fmt.Sprintf("http://adventofcode.com")
}

func (self *AdventOfCode) BaseURL() string {
	return fmt.Sprintf(AdventOfCodeRootURL() + "/%d/day/%d", self.year, self.day)
}

func (self *AdventOfCode) DescriptionURL() string {
	return self.BaseURL()
}

func (self *AdventOfCode) AnswerURL() string {
	return self.BaseURL() + "/answer"
}

func (self *AdventOfCode) InputURL() string {
	return self.BaseURL() + "/input"
}

func (self *AdventOfCode) BasePath() string {
	return "./problems/" + fmt.Sprintf("%d/%d", self.year, self.day)
}

func (self *AdventOfCode) DescriptionPath() string {
	return self.BasePath() + "/README.html"
}

func (self *AdventOfCode) AnswerPath() string {
	return self.BasePath() + "/answers"
}

func (self *AdventOfCode) InputPath() string {
	return self.BasePath() + "/input.txt"
}

func (self *AdventOfCode) SrcPath() string {
	return self.BasePath() + "/src"
}

func (self *AdventOfCode) BinPath() string {
	return self.BasePath() + "/bin"
}

func NewAdventOfCodeClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Printf("error creating cookie jar: %s\n", err)
		return nil, err
	}

	rootUrl := AdventOfCodeRootURL()
	adventOfCodeRootURL, err := url.Parse(rootUrl)
	if err != nil {
		fmt.Printf("url for string %s was malformed: %s\n", rootUrl, err)
		return nil, err
	}
	cookies := []*http.Cookie {
		&http.Cookie{
			Name: "session",
			Value: os.Getenv("SESSION"),
		},
	}
	jar.SetCookies(adventOfCodeRootURL, cookies)

	return &http.Client{
		Jar: jar,
	}, nil
}

func fetchURL(url string) (string, error) {
	client, err := NewAdventOfCodeClient()
	if err != nil {
		fmt.Printf("error creating advent of code client: %s\n", err)
		return "", err
	}

	response, err := client.Get(url)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", url, err)
		return "", err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("url %s could not be fetched status code: %d\n", url, response.StatusCode)
		return "", errors.New("url could not be fetched")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error reading response body form %s: %s\n", url, err)
		return "", err
	}

	return string(body), nil
}

func AttributeShouldBeRemoved(attribute []html.Attribute) bool {
	for _, attr := range attribute {
		if attr.Val == "sidebar" ||
			attr.Val == "sponsor" {
			return true
		}
	}
	return false
}

func FormatWebpage(node *html.Node) {
	child := node.FirstChild;
	for child != nil {
		if child.Data == "script" ||
			child.Data == "form" ||
			child.Data == "a" ||
			child.Data == "span" ||
			child.Data == "header" ||
			child.Data == "nav" ||
			child.Type == html.CommentNode ||
			AttributeShouldBeRemoved(child.Attr) {

			toRemove := child
			child = child.NextSibling
			node.RemoveChild(toRemove)
		} else if (child.Data == "article") {
			FormatWebpage(child)
			child = child.NextSibling
			for child != nil {
				toRemove := child
				child = child.NextSibling
				node.RemoveChild(toRemove)
			}
		} else {
			for index, attr := range child.Attr {
				if attr.Key == "href" {
					attr.Val = AdventOfCodeRootURL() + attr.Val
					child.Attr[index] = attr
				}
			}
			FormatWebpage(child)
			child = child.NextSibling
		}
	}
}

func formatDescriptionAsWebpage(description string) (string, error) {
	webpage, err := html.Parse(strings.NewReader(description))
	if err != nil {
		fmt.Printf("error parsing html: %s\n", err)
		return "", err
	}

	FormatWebpage(webpage)

	buffer := strings.Builder{}
	err = html.Render(&buffer, webpage)

	if err != nil {
		fmt.Printf("could not render webpage\n")
		return "", err
	}

	return buffer.String(), nil
}

func cacheFile(path string, content string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		fmt.Printf("cache file could not be created at %s: %s\n", path, err)
	}
	file.WriteString(content)
}

func (self *AdventOfCode) Description() (string, error) {
	path := self.DescriptionPath()
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		response, err := fetchURL(self.DescriptionURL())
		if err != nil {
			fmt.Printf("advent of code description was not found\n")
			return "", err
		}

		wepage, err := formatDescriptionAsWebpage(response)
		if err != nil {
			fmt.Printf("could not render webpage\n")
			return "", err
		}

		cacheFile(path, wepage)
		return wepage, nil
	} else {
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("could not read file path %s: %s\n", path, err)
			return "", err
		}
		return string(content), nil
	}
}

func (self *AdventOfCode) Input() (string, error) {
	path := self.InputPath()
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		response, err := fetchURL(self.InputURL())
		if err != nil {
			fmt.Printf("advent of code input was not found\n")
			return "", err
		}

		cacheFile(path, response)
		return response, nil
	} else {
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("could not read file path %s: %s\n", path, err)
			return "", err
		}
		return string(content), nil
	}
}

func runCommand(command string) (string, error) {
	args := strings.Fields(command)
	out, err := exec.Command(args[0], args[1:]...).Output()

	if err != nil {
		// fmt.Printf("Command %s failed: %s\n"	, command, err)
		return "", err
	}

	return string(out), nil
}

func compile(srcPath string, binPath string) {
	command := fmt.Sprintf("/usr/local/go/bin/go build -o %s %s", binPath, srcPath)
	runCommand(command)
}

func execute(binPath string, inputPath string) string {
	command := fmt.Sprintf("%s %s", binPath, inputPath)
	answer, err := runCommand(command)
	if err != nil {
		return ""
	} else {
		return strings.Trim(answer, " \r\n")
	}
}

func NewAdventOfCodeProblem(adventOfCode AdventOfCode) (*AdventOfCodeProblem, error) {
	requiredPaths := []string {
		adventOfCode.AnswerPath(),
		adventOfCode.SrcPath(),
		adventOfCode.BinPath(),
	}
	for _, requiredPath := range requiredPaths {
		if err := os.MkdirAll(requiredPath, os.ModePerm); err != nil {
			fmt.Printf("required path could not be created: %d\n", requiredPath)
			return nil, err
		}
	}

	description, err := adventOfCode.Description()
	if err != nil {
		fmt.Printf("advent of code description was not found\n")
		return nil, err
	}

	input, err := adventOfCode.Input()
	if err != nil {
		fmt.Printf("advent of code input was not found\n")
		return nil, err
	}

	problem := AdventOfCodeProblem {
		adventOfCode: adventOfCode,
		description: description,
		input: input,

		answers: []string{ "", "" },
	}
	for answerIndex := range problem.answers {
		srcPath, _ := problem.SrcPath(answerIndex)
		binPath, _ := problem.BinPath(answerIndex)
		answerPath, _ := problem.AnswerPath(answerIndex)
		compile(srcPath, binPath)
		answer := execute(binPath, adventOfCode.InputPath())
		if answer != "" {
			cacheFile(answerPath, answer)
			problem.answers[answerIndex] = answer
		}
	}

	return &problem, nil
}

func (self *AdventOfCodeProblem) AnswerBody(answer int) (string, error) {
	if answer < 0 || answer >= len(self.answers) {
		return "", errors.New("answer index out of range")
	}
	return fmt.Sprintf(`{
		"answer": "%s"
		"level": "%d",
	}`, self.answers[answer] , answer + 1), nil
}

func (self *AdventOfCodeProblem) AnswerPath(answer int) (string, error) {
	if answer < 0 || answer >= len(self.answers) {
		return "", errors.New("answer index out of range")
	}
	return self.adventOfCode.AnswerPath() + fmt.Sprintf("/answer%d.txt", answer + 1), nil
}

func (self *AdventOfCodeProblem) SrcPath(answer int) (string, error) {
	if answer < 0 || answer >= len(self.answers) {
		return "", errors.New("answer index out of range")
	}
	return self.adventOfCode.SrcPath() + fmt.Sprintf("/solution%d.go", answer + 1), nil
}

func (self *AdventOfCodeProblem) BinPath(answer int) (string, error) {
	if answer < 0 || answer >= len(self.answers) {
		return "", errors.New("answer index out of range")
	}
	return self.adventOfCode.BinPath() + fmt.Sprintf("/solution%d.exe", answer + 1), nil
}

func postJSON(url string, payload string) (string, error) {
	client, err := NewAdventOfCodeClient()
	if err != nil {
		fmt.Printf("error creating advent of code client: %s\n", err)
		return "", err
	}

	response, err := client.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		fmt.Printf("error posting json to %s: %s\n", url, err)
		return "", err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("url %s could not be fetched status code: %d\n", url, response.StatusCode)
		return "", errors.New("payload could not be posted")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error reading response body form %s: %s\n", url, err)
		return "", err
	}

	return string(body), nil
}

func (self *AdventOfCodeProblem) Answer(answer int) (string, error) {
	body, err := self.AnswerBody(answer)
	if err != nil {
		fmt.Printf("could not create body for answer[%d]: %s\n", answer, err)
		return "", err
	}

	url := self.adventOfCode.AnswerURL()
	response, err := postJSON(url, body)
	if err != nil {
		fmt.Printf("error posting answer[%d] to %s: %s\n", answer, url, err)
		return "", err
	}

	if strings.Contains(response, "You gave an answer too recently; you have to wait after submitting an answer before trying again.") {
		return "", errors.New("Too many checks for the answer, wait a bit and try again")
	}

	answerCheck := []string{
		"The first half of this puzzle is complete! It provides one gold star: *",
		"Both parts of this puzzle are complete! They provide two gold stars: **",
	}
	for answerCheckIndex := len(answerCheck)-1; answerCheckIndex >= answer; answerCheckIndex-- {
		if strings.Contains(response, answerCheck[answerCheckIndex]) {
			re := regexp.MustCompile(`(?s)Your puzzle answer was <code>(.*?)</code>`)
			solution := re.FindAllStringSubmatch(response, -1)[answer][1]
			return solution, nil
		}
	}

	if strings.Contains(response, "That's the right answer!") {
		return self.answers[answer], nil
	} else {
		return "", nil
	}
}

func sum(numbers []int) int {
	sum := 0
	for _, number := range numbers {
		sum += number
	}
	return sum
}

func app() int {
	problems := []*AdventOfCodeProblem{}
	for year := 2022; year <= 2022; year++ {
		for day := 1; day <= 25; day++ {
			adventOfCode := AdventOfCode{
				year: year,
				day: day,
			}
			problem, err := NewAdventOfCodeProblem(adventOfCode)
			if err != nil {
				fmt.Printf("advent of code problem could not be created\n")
			} else {
				problems = append(problems, problem)
			}
		}
	}

	correctAnswersStars := []int{0, 0}
	uncheckedAnswers := 0
	for problemIndex, problem := range problems {
		for answerIndex, answer := range problem.answers {
			answerHeader := fmt.Sprintf("TEST %4d [%d, %2d, %2s]:", problemIndex*2+answerIndex, problem.adventOfCode.year, problem.adventOfCode.day, strings.Repeat("*", answerIndex+1))

			if answer == "" {
				fmt.Printf("%s Answer is unknown! (skipping solution fetch to be friendly)\n", answerHeader)
				uncheckedAnswers += 1
				continue
			}

			solution, err := problem.Answer(answerIndex)
			if err != nil {
				fmt.Printf("%s Answer '%s' is unkedcked!: %s\n", answerHeader, answer, err)
				uncheckedAnswers += 1
				continue
			}

			if solution == answer && solution != ""{
				fmt.Printf("%s Answer '%s' is correct!\n", answerHeader, answer)
				correctAnswersStars[answerIndex] += 1
			} else if solution != "" {
				fmt.Printf("%s Answer '%s' is wrong! Correct solution is: '%s'\n", answerHeader, answer, solution)
			} else if answer != ""{
				fmt.Printf("%s Answer '%s' is wrong! Solution is unknown\n", answerHeader, answer)
			} else {
				fmt.Printf("%s Answer is unknown!\n", answerHeader)
			}
		}
	}

	correctAnswers := sum(correctAnswersStars)
	wrongAnswers := len(problems)*len(problems[0].answers) - correctAnswers - uncheckedAnswers
	fmt.Printf("SUMMARY: Correct answers: %d, Wrong answers: %d, Unkecked answers: %d\n", correctAnswers, wrongAnswers, uncheckedAnswers)
	fmt.Printf("SUMMARY: Problems with *: %d, Problems with **: %d\n", correctAnswersStars[0], correctAnswersStars[1])
	fmt.Printf("SUMMARY: Solved problems: %d, Unsolved problems: %d\n", correctAnswersStars[1], len(problems)*2-correctAnswers)
	return wrongAnswers
}

func main() {
	os.Exit(app())
}
