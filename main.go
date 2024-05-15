package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type EncodedResponse struct {
	Value string `json:"value"`
}

type Response struct {
	Children   []string    `json:"children"`
	Pagination interface{} `json:"pagination"`
}

type NestedResponse struct {
	BlockHeight string   `json:"blockHeight"`
	Values      []string `json:"values"`
}

var app = tview.NewApplication()
var columns [6]*tview.List
var logView *tview.TextView
var flex *tview.Flex
var currentColumn int

func main() {
	logView = tview.NewTextView().SetDynamicColors(true).SetWrap(true).SetChangedFunc(func() {
		app.Draw()
	})
	logView.SetBorder(true).SetTitle("Data")

	columnFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	for i := range columns {
		list := tview.NewList().SetWrapAround(false)
		list.SetBorder(true)
		list.ShowSecondaryText(false)
		columns[i] = list
		columnFlex.AddItem(columns[i], 0, 1, true)
	}

	flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(columnFlex, 0, 1, true).
		AddItem(logView, 0, 2, false)

	app.SetRoot(flex, true)

	// Initialize the first column
	initializeColumn("published", 0)
	currentColumn = 0
	app.SetFocus(columns[currentColumn])

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

func initializeColumn(path string, level int) int {
	children, err := fetchChildren(path)
	if len(children) == 0 || err != nil {
		// Fetch data from the alternative endpoint and log it
		data, err := fetchData(path)
		if err != nil {
			logMessage(fmt.Sprintf("[red]Error fetching data: %v", err))
		} else {
			logMessage(fmt.Sprintf("[green]%s", data))
		}
		app.SetFocus(columns[level-1])
		return 0
	}

	columns[level].Clear()
	for _, child := range children {
		childPath := path + "." + child
		levelCopy := level
		childPathCopy := childPath
		columns[level].AddItem(child, "", 0, func() {
			if levelCopy+1 < len(columns) {
				nextLevel := initializeColumn(childPathCopy, levelCopy+1)
				currentColumn = levelCopy + nextLevel
				app.SetFocus(columns[currentColumn])
			}
		})
	}

	// Set input capture for navigation
	columns[level].SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			if currentColumn > 0 {
				currentColumn--
				app.SetFocus(columns[currentColumn])
			}
			return nil
		case tcell.KeyRight:
			if currentColumn < len(columns)-1 {
				currentColumn++
				app.SetFocus(columns[currentColumn])
			}
			return nil
		}
		return event
	})

	return 1
}

func fetchChildren(path string) ([]string, error) {
	var url string
	if path == "published" {
		url = "https://main.api.agoric.net:443/agoric/vstorage/children/published"
	} else {
		url = fmt.Sprintf("https://main.api.agoric.net:443/agoric/vstorage/children/%s", path)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return response.Children, nil
}

func fetchData(path string) (string, error) {
	url := fmt.Sprintf("https://main.api.agoric.net:443/agoric/vstorage/data/%s", path)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	// Clean up the encoded JSON data

	str := string(body)
	cleanedValue := cleanJSON(str)
	// bytesValue := []byte(cleanedValue)
	// var encodedResponse EncodedResponse
	//if err := json.Unmarshal(bytesValue, &encodedResponse); err != nil {
	//		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	//	}

	// Pretty-print the cleaned JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(cleanedValue), "", "  "); err != nil {
		return "", fmt.Errorf("failed to pretty-print JSON: %v", err)
	}

	return prettyJSON.String(), nil
}

func cleanJSON(input string) string {
	replacer := strings.NewReplacer(
		`\\`, ``,
		`"#{`, `{`,
		`"{`, `{`,
		`}"`, `}`,
		`"#[`, `[`,
		`]"`, `]`,
		`"{`, `{`,
		`}"`, `}`,
		`\"`, `"`,
	)
	return replacer.Replace(replacer.Replace(input))
}

func logMessage(message string) {
	logView.Clear()
	logView.Write([]byte(message + "\n"))
}
