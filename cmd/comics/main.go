package main

import (
	"log"
	"os"

	"fmt"

	"github.com/jirwin/comics/src/comics"
	"gopkg.in/urfave/cli.v1"
)

const (
	W = 1024
	H = 512
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "template",
			Usage: "The url for a template to render",
		},
		cli.StringFlag{
			Name:  "imgur-client-id",
			Usage: "The client-id for your imgur app",
		},
		cli.StringSliceFlag{
			Name:  "text",
			Usage: "The text to use to fill comic bubbles",
		},
	}

	app.Action = func(c *cli.Context) error {
		if !c.IsSet("template") {
			fmt.Println("--template is a required argument.")
			return cli.ShowAppHelp(c)
		}
		tPath := c.String("template")

		t, err := comics.NewTemplate(tPath)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		outBytes, err := t.Render(c.StringSlice("text"))
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		imgUrl, err := comics.ImgurUpload(outBytes, c.String("imgur-client-id"))
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		fmt.Println(imgUrl)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	//dc := gg.NewContext(W, H)
	//
	//// draw text
	//dc.SetRGB(0, 0, 0)
	//dc.LoadFontFace("/Library/Fonts/Impact.ttf", 128)
	//dc.DrawStringAnchored("Gradient Text", W/2, H/2, 0.5, 0.5)
	//
	//// get the context as an alpha mask
	//mask := dc.AsMask()
	//
	//// clear the context
	//dc.SetRGB(1, 1, 1)
	//dc.Clear()
	//
	//// set a gradient
	//g := gg.NewLinearGradient(0, 0, W, H)
	//g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	//g.AddColorStop(1, color.RGBA{0, 0, 255, 255})
	//dc.SetFillStyle(g)
	//
	//// using the mask, fill the context with the gradient
	//dc.SetMask(mask)
	//dc.DrawRectangle(0, 0, W, H)
	//dc.Fill()
	//
	//dc.SavePNG("out.png")
}
