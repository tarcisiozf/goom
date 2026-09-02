package goom

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	onPause = make(chan struct{}, 1)
)

type keyEvent struct {
	key     byte
	pressed bool
}

type Game struct {
	width  int
	height int

	img       *ebiten.Image
	pixels    []byte
	keyEvents chan keyEvent

	finished bool
}

var _ ebiten.Game = (*Game)(nil)

const (
	keyRightArrow = 0xae
	keyLeftArrow  = 0xac
	keyUpArrow    = 0xad
	keyDownArrow  = 0xaf
	keyStrafeL    = 0xa0
	keyStrafeR    = 0xa1
	keyUse        = 0xa2
	keyFire       = 0xa3
	keyEscape     = 27
	keyEnter      = 13
	keyTab        = 9
)

var keyMap = map[byte]ebiten.Key{
	keyRightArrow: ebiten.KeyRight,
	keyLeftArrow:  ebiten.KeyLeft,
	keyUpArrow:    ebiten.KeyUp,
	keyDownArrow:  ebiten.KeyDown,
	keyStrafeL:    ebiten.KeyA,
	keyStrafeR:    ebiten.KeyD,
	keyUse:        ebiten.KeyE,
	keyFire:       ebiten.KeySpace,
	keyEscape:     ebiten.KeyEscape,
	keyEnter:      ebiten.KeyEnter,
	keyTab:        ebiten.KeyTab,
}

func (g *Game) Update() error {
	if g.finished {
		return ebiten.Termination
	}
	for code, key := range keyMap {
		var pressed bool
		if inpututil.IsKeyJustPressed(key) {
			pressed = true
		} else if inpututil.IsKeyJustReleased(key) {
			pressed = false
		} else {
			continue
		}
		g.keyEvents <- keyEvent{code, pressed}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		onPause <- struct{}{}
		close(onPause)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.img == nil {
		return
	}
	g.img.WritePixels(g.pixels)
	screen.Fill(color.Black)
	screen.DrawImage(g.img, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}

func (g *Game) SetResolution(width, height int) {
	g.width = width
	g.height = height
	g.pixels = make([]byte, width*height*4) // RGBA format
	g.img = ebiten.NewImage(width, height)
}

func newGame() *Game {
	return &Game{
		width:  16,
		height: 16,

		keyEvents: make(chan keyEvent, 32),
	}
}
