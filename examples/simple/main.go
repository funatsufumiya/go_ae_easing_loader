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
    t := math.Mod(elapsed, 14.0)
    y := g.ae.Get(t, 0)
    ebitenutil.DebugPrintAt(screen, "t: "+fmt.Sprintf("%.2f", t), 50, 50)
    vector.DrawFilledRect(screen, 100, y-600, 50, 50, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 300, 300
}

func main() {
    var g Game
    g.startTime = time.Now()
    if err := g.ae.LoadJsonFile(path.Join("examples", "simple", "test.json")); err != nil {
        log.Fatal(err)
    }
    ebiten.SetWindowSize(300, 300)
    ebiten.SetWindowTitle("AEEasingLoader Example")
    if err := ebiten.RunGame(&g); err != nil {
        log.Fatal(err)
    }
}