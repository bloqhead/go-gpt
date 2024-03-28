package main

import (
	"context"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	// "gioui.org/text"
	"github.com/sashabaranov/go-openai"

	"github.com/joho/godotenv"
)

var (
	apiKey string

	labelText = "Hello, I'm ChatGPT. Let's start a conversation!"
	theme = material.NewTheme()
	title = "Go GPT"

	// app dimensions and spacing
	appWidth = unit.Dp(400)
	appHeight = unit.Dp(600)
	appPadding = unit.Dp(16)

	// widgets
	button = new(widget.Clickable)
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

				layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						return layout.Inset{
							Top: appPadding,
							Bottom: appPadding,
							Left: appPadding,
							Right: appPadding,
						}.Layout(gtx, func(gtx C) D {
							// content area
							return layout.Center.Layout(gtx, func(gtx C) D {
								// gtx.Constraints.Min = gtx.Constraints.Max

								return layout.Flex{
									Axis: layout.Vertical,
									Spacing: layout.SpaceBetween,
								}.Layout(gtx,
									// label
									layout.Flexed(1, func(gtx C) D {
										return layout.Center.Layout(gtx, func(gtx C) D {
											label := material.Label(theme, unit.Sp(12), labelText)

											return label.Layout(gtx)
										})
									}),

									// input and button
									layout.Flexed(5, func(gtx C) D {
										return layout.Center.Layout(gtx, func(gtx C) D {
											return layout.Flex{
												Alignment: layout.Middle,
											}.Layout(gtx,
												// prompt input
												layout.Flexed(4, func(gtx C) D {
													input := material.Editor(theme, promptInput, "Enter your query...")

													return input.Layout(gtx)
												}),

												// submit button
												layout.Rigid(func(gtx C) D {
													return layout.Center.Layout(gtx, func(gtx C) D {
														btn := material.Button(theme, button, "Submit")

														return btn.Layout(gtx)
													})
												}),
											)
										})
									}),
								)
							})
						})
					}),

					layout.Stacked(func(gtx C) D {
						if requestProcessing {
							bgColor := color.NRGBA{R: 0, G: 0, B: 0, A: 150}

							return layout.Center.Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min = gtx.Constraints.Max

								return layout.Stack{
									Alignment: layout.Center,
								}.Layout(gtx,
									layout.Expanded(func(gtx C) D {
										// semi-transparent background overlay
										return layout.Inset{
											Top: unit.Dp(0),
											Bottom: unit.Dp(0),
											Left: unit.Dp(0),
											Right: unit.Dp(0),
										}.Layout(gtx, func(gtx C) D {
											paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
											paint.PaintOp{}.Add(gtx.Ops)
											return D{}
										})
									}),

									// loader
									layout.Expanded(func(gtx C) D {
										loader := material.Loader(theme)

										return loader.Layout(gtx)
									}),
								)
							})
						}

						return D{}
					}),
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
