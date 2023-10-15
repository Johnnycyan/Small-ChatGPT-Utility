//go:generate fyne bundle -o bundled.go icon.png

// This is a simple program for quickly interacting with ChatGPT
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"time"

	//"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	//"fyne.io/fyne/v2/data/binding"
	//"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/ztrue/tracerr"

	//"fyne.io/fyne/v2/theme"

	"strings"

	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"github.com/micmonay/keybd_event"
	openai "github.com/sashabaranov/go-openai"
	"golang.design/x/clipboard"
	"golang.design/x/hotkey"
)

var (
	a = app.NewWithID("com.johnnycyan.chatgpt")
	// get the key from the OPEN_API_KEY environment variable
	key    = os.Getenv("OPENAI_API_KEY")
	client = openai.NewClient(key)
)

type myEntry struct {
	widget.Entry
	OnKeyDown func(*fyne.KeyEvent)
}

func (m *myEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name != "Return" && key.Name != "BackSpace" && key.Name != "Delete" {
		m.Entry.TypedKey(key)
		return
	}
	if m.OnKeyDown != nil {
		m.Entry.TypedKey(key)
		m.OnKeyDown(key)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	icon := resourceIconPng
	a.SetIcon(icon)

	w := a.NewWindow("ChatGPT")
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	entryWindow(w)

	askMenu := func() {
		log.Println("Tapped ask")
		a.Preferences().SetString("type", "ask")
		w.SetTitle("Enter Question")
		w.Show()
	}

	translatetoMenu := func() {
		kb, err := keybd_event.NewKeyBonding()
		if err != nil {
			panic(err)
		}
		log.Println("Tapped translate to")
		a.Preferences().SetString("type", "translate to")
		kb.SetKeys(keybd_event.VK_A, keybd_event.VK_X)
		kb.HasCTRL(true)
		kb.Launching()
		time.Sleep(100 * time.Millisecond)

		// kb.SetKeys(keybd_event.VK_X)
		// kb.HasCTRL(true)
		// kb.Launching()
		w.SetTitle("Enter Language")
		w.Show()
	}

	translateMenu := func() {
		log.Println("Tapped translate")
		a.Preferences().SetString("type", "translate")
		translate()
	}

	grammarMenu := func() {
		kb, err := keybd_event.NewKeyBonding()
		if err != nil {
			panic(err)
		}
		log.Println("Tapped grammar")
		a.Preferences().SetString("type", "grammar")
		kb.SetKeys(keybd_event.VK_A, keybd_event.VK_X)
		kb.HasCTRL(true)
		kb.Launching()
		time.Sleep(100 * time.Millisecond)
		grammar()
	}

	askItem := fyne.NewMenuItem("Ask", askMenu)
	translateToItem := fyne.NewMenuItem("Translate To", translatetoMenu)
	translateItem := fyne.NewMenuItem("Translate", translateMenu)
	grammarItem := fyne.NewMenuItem("Grammar", grammarMenu)

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			askItem,
			translateToItem,
			translateItem,
			grammarItem,
		)
		desk.SetSystemTrayMenu(m)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				tracerr.PrintSourceColor(err.(error))
			}
		}()
		askShortcut := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF19)
		err := askShortcut.Register()
		if err != nil {
			log.Fatalf("hotkey: failed to register hotkey: %v", err)
		}
		for range askShortcut.Keydown() {
			askMenu()
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				tracerr.PrintSourceColor(err.(error))
			}
		}()
		translatetoShortcut := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF17)
		err := translatetoShortcut.Register()
		if err != nil {
			log.Fatalf("hotkey: failed to register hotkey: %v", err)
		}
		for range translatetoShortcut.Keydown() {
			translatetoMenu()
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				tracerr.PrintSourceColor(err.(error))
			}
		}()
		translateShortcut := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF18)
		err := translateShortcut.Register()
		if err != nil {
			log.Fatalf("hotkey: failed to register hotkey: %v", err)
		}
		for range translateShortcut.Keydown() {
			translateMenu()
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				tracerr.PrintSourceColor(err.(error))
			}
		}()
		grammarShortcut := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyF18)
		err := grammarShortcut.Register()
		if err != nil {
			log.Fatalf("hotkey: failed to register hotkey: %v", err)
		}
		for range grammarShortcut.Keydown() {
			grammarMenu()
		}
	}()

	w.Resize(fyne.NewSize(800, w.Content().MinSize().Height))
	w.CenterOnScreen()
	w.FixedSize()
	a.Run()
}

func entryWindow(w fyne.Window) *fyne.Container {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	entry := &myEntry{}
	entry.ExtendBaseWidget(entry)
	entry.SetPlaceHolder("Type here...")

	entry.OnKeyDown = func(key *fyne.KeyEvent) {
		if key.Name == "Return" {
			if a.Preferences().StringWithFallback("type", "ask") == "ask" {
				question(entry)
				entry.Text = ""
				w.Hide()
			} else if a.Preferences().StringWithFallback("type", "ask") == "translate to" {
				translateto(entry)
				entry.Text = ""
				w.Hide()
			}
		}
	}

	entryBox := container.NewVBox(entry)
	w.SetContent(entryBox)
	w.Canvas().Focus(entry)

	return entryBox
}

func question(entry *myEntry) {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	question := entry.Text
	systemMessage := "Answer the prompt very concisely, in no more than a few sentences. If asked to convert something like celsius to fahrenheit or do math just output the answer, don't explain the formula"
	log.Println("question " + question)
	go chatgpt(systemMessage, question, openai.GPT3Dot5Turbo, 10*time.Millisecond)
}

func translateto(entry *myEntry) {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	language := entry.Text
	text := string(clipboard.Read(clipboard.FmtText))
	systemMessage := "Take the prompt given and translate it to " + language + " using casual wording, don't correct punctuation and don't add commas."
	log.Println("translate " + text + " to " + language)
	go chatgpt(systemMessage, text, openai.GPT4, 30*time.Millisecond)
}

func translate() {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	text := string(clipboard.Read(clipboard.FmtText))
	systemMessage := "You will take the text given in the message from the user and translate it to English using casual wording, don't correct punctuation and don't add commas. Output with the following format: Translation: translated message | Language: language of the original message before translation if there are multiple languages or the language is already English then just return Message is already in English"
	log.Println("translate " + text)
	go chatgpt(systemMessage, text, openai.GPT4, 30*time.Millisecond)
}

func grammar() {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	text := string(clipboard.Read(clipboard.FmtText))
	systemMessage := "Take the prompt given and fix the grammar and spelling. Do not explain your changes. Just fix them."
	log.Println("grammar " + text)
	go chatgpt(systemMessage, text, openai.GPT3Dot5Turbo, 10*time.Millisecond)
}

func chatgpt(systemMessage string, message string, model string, delay time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			tracerr.PrintSourceColor(err.(error))
		}
	}()
	var content string
	resp := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
		Stream: true,
	}
	stream, err := client.CreateChatCompletionStream(context.Background(), resp)
	if err != nil {
		robotgo.TypeStr(err.Error())
		return
	}
	defer stream.Close()
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("\nStream finished")
			stream.Close()
			return
		}

		if err != nil {
			log.Printf("\nStream error: %v\n", err)
			robotgo.TypeStr(err.Error())
			return
		}

		content = strings.ReplaceAll(response.Choices[0].Delta.Content, "\n", " ")

		// for each character in content
		for _, c := range content {
			robotgo.TypeStr(string(c))
			time.Sleep(delay)
		}
	}
}
