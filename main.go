package main

import (
	"context"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/sashabaranov/go-openai"

	"github.com/joho/godotenv"
)

var (
	apiKey string

	labelText = "Hello, I'm ChatGPT. Let's start a conversation!"
	theme     = material.NewTheme()
	title     = "Go GPT"

	// app dimensions and spacing
	appWidth  = unit.Dp(400)
	appHeight = unit.Dp(600)
	appMargin = unit.Dp(16)

	// widgets
	button      = new(widget.Clickable)
	promptInput = new(widget.Editor)

	requestProcessing = false
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// todo: add a separate initial screen where the user can set their own API key.
	// todo: additionally, figure out where and how to store the API key securely within the app.
	apiKey = os.Getenv("OPENAI_API_KEY")

	// create our window
	go func() {
		w := app.NewWindow(
			app.Title(title),
			app.Size(appWidth, appHeight),
			app.MinSize(appWidth, appHeight),
		)
		err := run(w)

		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}

func run(w *app.Window) error {
	var ops op.Ops

	// listen for app events
	for {
		switch e := w.NextEvent().(type) {

		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// component spacing
			margins := layout.Inset{
				Top:    appMargin,
				Bottom: appMargin,
				Left:   appMargin,
				Right:  appMargin,
			}

			// handle button click events
			if button.Clicked(gtx) {
				if !requestProcessing {
					requestProcessing = true

					go func() {
						client := openai.NewClient(apiKey)
						query := strings.TrimSpace(promptInput.Text())
						res, err := generateChatResponse(client, query)

						if err != nil {
							log.Printf("GPT Error: %s\n", err)
						}

						labelText = res
						requestProcessing = false

						w.Invalidate()
					}()
				}
			}

			// app layout
			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,

				// response display
				layout.Flexed(1,
					func(gtx layout.Context) layout.Dimensions {
						return margins.Layout(gtx, func(gtx C) D {
							return material.Label(theme, unit.Sp(12), labelText).Layout(gtx)
						})
					},
				),

				// prompt input
				layout.Rigid(
					func(gtx C) D {
						return margins.Layout(gtx, func(gtx C) D {
							return material.Editor(theme, promptInput, "Type your message here...").Layout(gtx)
						})
					},
				),

				// submit button
				layout.Rigid(
					func(gtx C) D {
						responseText := "Generate Response"

						if requestProcessing {
							responseText = "Generating response..."
						}

						return margins.Layout(gtx, func(gtx C) D {
							return material.Button(theme, button, responseText).Layout(gtx)
						})
					},
				),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func generateChatResponse(client *openai.Client, prompt string) (string, error) {
	// isLoading := true

	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}
