package main

import (
	"fmt"
	"image/color"
	"log"
	"path"
	"time"

	math "github.com/chewxy/math32"

	"github.com/funatsufumiya/go_ae_easing_loader/easing_loader"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	ae easing_loader.AEEasingLoader
	startTime time.Time
	propertyIndex1 int
	propertyIndex2 int
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	elapsed := float32(time.Since(g.startTime).Seconds())
	t := math.Mod(elapsed, 6.0)
	x, y := g.ae.Get2(t, g.propertyIndex1)
	deg := g.ae.Get(t, g.propertyIndex2)

	ebitenutil.DebugPrintAt(screen, "t: "+fmt.Sprintf("%.2f", t), 50, 50)
	vector.DrawFilledRect(screen, x-400, y-100, 50, 50, color.RGBA{255, 255, 255, 255}, false)

	// Red rotated rectangle
	vector.DrawFilledRect(screen, 0, 0, 0, 0, color.RGBA{0, 0, 0, 0}, false) // dummy to keep API usage
	op := ebiten.GeoM{}
	op.Translate(200, 50)
	op.Rotate(float64(deg * math.Pi / 180))
	img := ebiten.NewImage(200, 50)
	img.Fill(color.RGBA{255, 0, 0, 255})
	screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: op})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	var g Game
	g.startTime = time.Now()
	if err := g.ae.LoadJsonFile(path.Join("examples", "advanced", "test3.json")); err != nil {
		log.Fatal(err)
	}
	g.ae.DumpTracks()
	g.propertyIndex1, _ = g.ae.GetPropertyIndex("Position", "A", "")
	g.propertyIndex2, _ = g.ae.GetPropertyIndex("Rotation", "B", "")
	
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("AEEasingLoader Advanced Example")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
