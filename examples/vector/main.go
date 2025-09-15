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
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	elapsed := float32(time.Since(g.startTime).Seconds())
	t := math.Mod(elapsed, 6.0)
	x, y := g.ae.Get2(t, 0)
	ebitenutil.DebugPrintAt(screen, "t: "+fmt.Sprintf("%.2f", t), 50, 50)
	vector.DrawFilledRect(screen, x-400, y-100, 50, 50, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	var g Game
	g.startTime = time.Now()
	if err := g.ae.LoadJsonFile(path.Join("examples", "vector", "test2.json")); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(600, 600)
	ebiten.SetWindowTitle("AEEasingLoader Vector Example")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
