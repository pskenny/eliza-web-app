package main

// IMPORTANT: I was unable to resolve error in adding responses to corresponding keys, making this non-functioning

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type key struct {
	keyword   string
	responses []string
}

// Clean up input text a bit
func sanatiseInput(input string) string {
	// Set text to lower case and remove spacing from left and right
	return strings.TrimSpace(strings.ToLower(input))
}

// read script file and set up keys
func loadKeys() []key {
	// Open the file from local file system
	file, err := os.Open("scripts/original.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	keyList := []key{}

	// Read the file line by line.
	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		// Get the next line to check
		switch line := scanner.Text(); {
		// If line starts with "key: " it's a key, add to key list
		case strings.HasPrefix(line, "key: "):
			// Get keyword text
			keyword := strings.Replace(line, "key: ", "", -1)
			newKey := key{keyword: keyword, responses: []string{}}
			// Add key to key list
			keyList = append(keyList, newKey)
		// If text is response, add it to previous key (in file no response is before "key: ")
		case strings.HasPrefix(line, "reasmb: "):
			// Get last key in key list
			currentKey := keyList[len(keyList)-1]
			// Get response text
			response := strings.Replace(line, "reasmb: ", "", -1)

			// Add response text to response list
			currentKey.responses = append(currentKey.responses, response)
			// NOTE: I was unable to resolve the error in the above lines, not adding response to key, making Eliza non-functioning
			// It also causes cascading errors in elizaReply method getting random numbers
		}
	}

	return keyList
}

func elizaReply(input string) string {
	input = sanatiseInput(input)
	output := ""
	anyMatch := true

	// Key words and responses, keys in regex and response using string formatting
	keyList := loadKeys()

	// No response (empty) replies
	noResponse := []string{"What?", "Huh?", "Come again?", "I don't really understand."}
	// No key found replies
	noKeys := []string{"Please go on...", "I see", "And?", "Could you elaborate?"}

	// No input
	if input == "" {
		// Return one of the no response replies
		return noResponse[rand.Intn(len(noResponse))]
	}

	// Loop over keys so check matches
	for _, currentKey := range keyList {
		reg := regexp.MustCompile(currentKey.keyword)
		matched := reg.MatchString(input)

		// Match found
		if matched {
			// Get one of the responses
			newResponse := currentKey.responses[rand.Intn(len(currentKey.responses))]

			// replace keywork with reply and fill in remaining
			reg := regexp.MustCompile(currentKey.keyword)
			// Replace keyword with response in input
			input = reg.ReplaceAllString(input, newResponse)

			// Check if response is using string formatting, if so, add reflected statement to key response
			match, _ := regexp.MatchString("%v", newResponse)
			if match {
				// Remove keyword from input, leaving things to be reflected]
				reflected := reflect(reg.ReplaceAllString(input, ""))
				output = reg.ReplaceAllString(reflected, newResponse)
				// Output uses
				input = fmt.Sprintf(newResponse, output)
			} else {
				// Response is all that's outputted
				input = newResponse
			}

			anyMatch = false
			break
		}
	}

	// No key matches found, choose one of the no key found replies
	if anyMatch {
		output = noKeys[rand.Intn(len(noKeys))]
	}

	return output
}

// Reflect replaces some personal pronouns and other words
func reflect(input string) string {
	// Split the input on word boundaries (ASCII only)
	boundaries := regexp.MustCompile(`\b`)
	// Split input at boundaries
	splitInput := boundaries.Split(input, -1)

	// List the reflections.
	reflections := [][]string{
		{`are`, `am`},
		{`am`, `are`},
		{`was`, `were`},
		{`i`, `you`},
		{`i'd`, `you would`},
		{`i've`, `you have`},
		{`i'll`, `you will`},
		{`my`, `your`},
		{`your`, `my`},
		{`you've`, `i have`},
		{`you'll`, `i will`},
		{`you will`, `i will`},
		{`you`, `me`},
		{`me`, `you`}}

	// Loop over every word
	for index, word := range splitInput {
		// Loop over every reflection
		for _, reflections := range reflections {
			// check if current word matches reflection (first index in array)
			if matched, _ := regexp.MatchString(reflections[0], word); matched {
				// change word to reflection
				splitInput[index] = reflections[1]
			}
		}
	}

	// return all words rejoined
	return strings.Join(splitInput, ``)
}

func main() {
	// Serve static HTML index file, adapted from http://www.alexedwards.net/blog/serving-static-sites-with-go
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/user-input", elizaHandler)

	http.ListenAndServe(":8080", nil)
}

func elizaHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, elizaReply(r.URL.Query().Get("value")))
}
