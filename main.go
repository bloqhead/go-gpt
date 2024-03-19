package main

import (
	"context"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
 	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/unit"
	"gioui.org/text"
	"github.com/sashabaranov/go-openai"
)

var (
	client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	labelText = "Hello, I'm ChatGPT. I can tell jokes."
	theme = material.NewTheme()
	button = new(widget.Clickable)
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func main() {

	// create our window
	go func() {
		w := app.NewWindow()
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

				// handle button click events
				if button.Clicked(gtx) {
					go func() {
						res, err := generateChatResponse(client, "Tell me a joke about Go developers.")

						if err != nil {
							log.Printf("GPT Error: %s\n", err)
						}

						labelText = res

						w.Invalidate()
					}()
				}

				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// label
					layout.Flexed(1, func(gtx C) D {
						return layout.Center.Layout(gtx, func(gtx C) D {
							label := material.Label(theme, unit.Sp(20), labelText)
							label.Alignment = text.Middle

							return label.Layout(gtx)
						})
					}),

					// button
					layout.Flexed(1, func(gtx C) D {
						return layout.Center.Layout(gtx, func(gtx C) D {
							btn := material.Button(theme, button, "Tell me a joke")

							return btn.Layout(gtx)
						})
					}),
				)

				e.Frame(gtx.Ops)
		}
	}
}

func generateChatResponse(client *openai.Client, prompt string) (string, error) {
	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
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
