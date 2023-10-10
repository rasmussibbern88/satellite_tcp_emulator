package plotting

import (
	"image/color"
	"log"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var (
	BLACK   = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
	ORANGE  = color.RGBA{R: 0xE6, G: 0x9F, B: 0x00, A: 0xFF}
	CYAN    = color.RGBA{R: 0x56, G: 0xB4, B: 0xE9, A: 0xFF}
	GREEN   = color.RGBA{R: 0x00, G: 0x9E, B: 0x73, A: 0xFF}
	YELLOW  = color.RGBA{R: 0xF0, G: 0xE4, B: 0x42, A: 0xFF}
	BLUE    = color.RGBA{R: 0x00, G: 0x72, B: 0xB2, A: 0xFF}
	RED     = color.RGBA{R: 0xD5, G: 0x5E, B: 0x00, A: 0xFF}
	MAGENTA = color.RGBA{R: 0xCC, G: 0x79, B: 0xA7, A: 0xFF}
)

var pallete = []color.RGBA{BLACK, ORANGE, CYAN, GREEN, YELLOW, BLUE, RED, MAGENTA}

func TimeSeriesLinePlot(X []time.Time, Y []float64, title, xlabel, ylabel, fname string) {
	newX := make([]float64, len(X))
	for i, X := range X {
		newX[i] = float64(X.Unix())
	}
	LinePlot(newX, Y, title, xlabel, ylabel, fname)
}

func LinePlot(X, Y []float64, title, xlabel, ylabel, fname string) {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel
	p.Add(plotter.NewGrid())

	data := make(plotter.XYs, len(X))
	for i := 0; i < len(X); i++ {
		data[i].X = X[i]
		data[i].Y = Y[i]
	}
	line, err := plotter.NewLine(data)
	if err != nil {
		log.Panic(err)
	}
	line.Color = pallete[3]
	p.Add(line)

	err = p.Save(20*vg.Centimeter, 10*vg.Centimeter, fname+".pdf")
	if err != nil {
		log.Panic(err)
	}
	// cbbPalette <- c("#000000", "#E69F00", "#56B4E9", "#009E73", "#F0E442", "#0072B2", "#D55E00", "#CC79A7")

}
