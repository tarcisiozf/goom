package goom

import (
	"fmt"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tarcisiozf/wasp"
	"github.com/tarcisiozf/wasp/memory"
)

type DoomWasi struct {
	startTime time.Time
	initOnce  sync.Once

	width, height int
	defaultMemory memory.Memory

	game       *Game
	frameCount int
}

func newDoomWasi(game *Game) *DoomWasi {
	return &DoomWasi{
		game: game,
	}
}

func (dw *DoomWasi) init(resx, resy int32) {
	dw.initOnce.Do(func() {
		dw.width = int(resx)
		dw.height = int(resy)
		ebiten.SetWindowSize(dw.width, dw.height)
		ebiten.SetWindowTitle("GOOM")

		dw.game.SetResolution(dw.width, dw.height)
	})
}

func (dw *DoomWasi) drawFrame(ptr int32) {
	if dw.game == nil {
		return
	}

	numPixels := dw.width * dw.height
	size := numPixels * 4
	data := dw.defaultMemory.Load(int(ptr), size)

	pixels := make([]byte, size)
	for i := 0; i < numPixels; i++ {
		offset := i * 4
		pixels[offset] = data[offset+2]   // R
		pixels[offset+1] = data[offset+1] // G
		pixels[offset+2] = data[offset]   // B
		pixels[offset+3] = 0xFF           // A
	}
	copy(dw.game.pixels, pixels)

	dw.frameCount++
	elapsedTime := time.Since(dw.startTime).Seconds()
	if elapsedTime > 0 {
		fps := float64(dw.frameCount) / elapsedTime
		fmt.Printf("Frame Count: %d, FPS: %.2f\n", dw.frameCount, fps)
	}
}

func (dw *DoomWasi) getTickMs() int32 {
	elapsed := time.Since(dw.startTime)
	ms := int32(elapsed.Milliseconds())
	return ms
}

func (dw *DoomWasi) sleepMs(ms int32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (dw *DoomWasi) getKey() int32 {
	if dw.game == nil {
		return -1
	}

	var event keyEvent
	select {
	case event = <-dw.game.keyEvents:
	default:
		return -1
	}

	key := int32(event.key)
	if event.pressed {
		key |= 1 << 9
	}

	return key
}

func (dw *DoomWasi) SetMemory(store *wasp.Store) {
	dw.defaultMemory = store.Memories[0]
}

func (dw *DoomWasi) Register(linker *wasp.Linker) error {
	dw.startTime = time.Now()

	errors := []error{
		linker.Define("dg", "init", dw.init),

		linker.Define("dg", "draw_frame", dw.drawFrame),

		linker.Define("dg", "sleep_ms", dw.sleepMs),

		linker.Define("dg", "get_ticks_ms", dw.getTickMs),

		linker.Define("dg", "get_key", dw.getKey),

		linker.Define("dg", "set_window_title", func(ptr, len int32) {
			fmt.Printf("DB set_window_title called with ptr=%d, len=%d\n", ptr, len)
		}),

		// UBSan stubs - allow undefined behavior to continue without panicking
		linker.Define("env", "__ubsan_handle_shift_out_of_bounds", func(dataPtr, lhs, rhs int32) {
			// Log but don't panic - this allows DOOM to continue despite UB
			fmt.Printf("[UBSAN] shift out of bounds: lhs=%d, rhs=%d (continuing)\n", lhs, rhs)
		}),
	}
	for _, err := range errors {
		if err != nil {
			return fmt.Errorf("error defining function: %w", err)
		}
	}
	return nil
}
