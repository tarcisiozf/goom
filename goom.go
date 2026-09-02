package goom

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tarcisiozf/wasp"
	"github.com/tarcisiozf/wasp/wasi"
)

const (
	dumpFileName = "dump.bin"
)

var (
	linker = wasp.NewLinker()

	options = []wasp.InstanceOption{
		wasp.WithLinker(linker),
		wasp.IgnoreUnreachable(), // Allow DOOM to continue despite UBSan panics
	}
)

type Goom struct {
	store    *wasp.Store
	instance *wasp.Instance
	game     *Game

	paused bool
}

func New(wadFile string) (*Goom, error) {
	wasm, err := os.ReadFile("dg.wasm")
	if err != nil {
		return nil, fmt.Errorf("error reading WASM file: %w", err)
	}

	module, err := wasp.NewModule(wasm)
	if err != nil {
		return nil, fmt.Errorf("error creating module: %w", err)
	}

	game := newGame()

	doomWasi := newDoomWasi(game)
	if err := doomWasi.Register(linker); err != nil {
		return nil, fmt.Errorf("error registering DOOM WASI: %w", err)
	}

	sp := wasi.NewWasiSnapshotPreview1()
	sp.SetArgs([]string{wadFile}) // Pass remaining args to WASI
	sp.AddPreopen(3, ".")         // Preopen current directory as fd 3
	if err := sp.Register(linker); err != nil {
		return nil, fmt.Errorf("error registering WASI snapshot preview 1: %w", err)
	}

	store, instance, err := initializeInstance(module)
	if err != nil {
		return nil, fmt.Errorf("error initializing instance: %w", err)
	}

	sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
	doomWasi.SetMemory(store)

	return &Goom{
		store:    store,
		instance: instance,
		game:     game,
	}, nil
}

func (g *Goom) Run() error {
	errCh := make(chan error, 1)
	defer close(errCh)

	go (func() {
		defer func() {
			g.game.finished = true
		}()

		if err := g.instance.Run(); err != nil {
			errCh <- fmt.Errorf("error running instance: %w", err)
			return
		}

		if g.paused {
			serializeState(g.store, g.instance)
		}

		errCh <- nil
	})()

	go func() {
		<-onPause
		g.paused = true
		g.instance.Pause()
	}()

	if err := ebiten.RunGame(g.game); err != nil {
		return fmt.Errorf("error running game: %w", err)
	}

	return <-errCh
}

func (g *Goom) Pause() {
	g.paused = true
}

func initializeInstance(module *wasp.Module) (*wasp.Store, *wasp.Instance, error) {
	//dump, err := os.Open(dumpFileName)
	//if err == nil {
	//	defer dump.Close()
	//	return loadStateFromDump(module, dump)
	//}
	//if os.IsNotExist(err) {
	return createNewInstance(module)
	//}
	//return nil, nil, fmt.Errorf("error opening dump file: %w", err)
}

func loadStateFromDump(module *wasp.Module, dump *os.File) (*wasp.Store, *wasp.Instance, error) {
	start := time.Now()
	store, instance, err := wasp.DeserializeState(dump, module, linker)
	if err != nil {
		return nil, nil, fmt.Errorf("error deserializing state: %w", err)
	}
	fmt.Printf("Deserialization took: %v\n", time.Since(start))
	return store, instance, nil
}

func createNewInstance(module *wasp.Module) (*wasp.Store, *wasp.Instance, error) {
	store := wasp.NewStore(module)

	instance, err := wasp.NewInstance(
		module,
		store,
		options...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating instance of module: %w", err)
	}

	fn, err := module.GetExportedFunction("_start")
	if err != nil {
		return nil, nil, fmt.Errorf("error getting start function: %w", err)
	}

	if _, err := instance.Call(fn); err != nil {
		return nil, nil, fmt.Errorf("error calling start function: %w", err)
	}

	return store, instance, nil
}

func serializeState(store *wasp.Store, instance *wasp.Instance) {
	file, err := os.Create(dumpFileName)
	if err != nil {
		println("Error creating dump file:", err.Error())
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	start := time.Now()
	err = wasp.SerializeState(writer, store, instance)
	if err != nil {
		panic(err)
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}

	fmt.Printf("Serialization took: %v\n", time.Since(start))
}
