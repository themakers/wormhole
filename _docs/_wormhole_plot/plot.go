package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/fogleman/ln/ln"
	"github.com/skratchdot/open-golang/open"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

// go get -t -x -u ./...
// go run plot.go

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigs
		log.Println(">>> caught", sig)
		cancel()
	}()

	shape1 := ln.NewFunction(func(x, y float64) float64 {
		return 1/(x*x+y*y) - 2
	}, ln.Box{
		Min: ln.Vector{X: -2, Y: -2, Z: -2},
		Max: ln.Vector{X: 2, Y: 2, Z: 0},
	}, ln.Below)

	shape1 = ln.NewTransformedShape(shape1,
		ln.Translate(ln.Vector{X: 0, Y: 0, Z: 0.01}),
	)

	shape2 := ln.NewTransformedShape(shape1,
		ln.Rotate(ln.Vector{X: 0, Y: 1, Z: 0}, ln.Radians(180)).
			Rotate(ln.Vector{X: 0, Y: 0, Z: 1}, ln.Radians(0.2)),
	)

	marks := []ln.Shape{
		ln.NewSphere(ln.Vector{X: 0, Y: 0, Z: 0}, 0.15),
		//ln.NewSphere(ln.Vector{X: 1, Y: 0, Z: 0}, 0.15),
		//ln.NewSphere(ln.Vector{X: 0, Y: 1, Z: 0}, 0.15),
		//ln.NewSphere(ln.Vector{X: 0, Y: 0, Z: 1}, 0.15),
		//ln.NewSphere(ln.Vector{X: 0, Y: 0, Z: -1}, 0.15),
	}

	shape1.Compile()
	shape2.Compile()

	if err := renderMovie(ctx, RenderOptions{
		Eye:    ln.Vector{X: 0, Y: 6, Z: 1.6},
		Center: ln.Vector{X: 0, Y: 0, Z: 0},
		Up:     ln.Vector{X: 0, Y: 0, Z: 1.2},

		Fovy: 45,
		Near: 0.1,
		Far:  100,
		Step: 0.01,
		//Step: 0.001,

		Width:  320,
		Height: 320,

		Frames: 10,
		Angle:  5,
	}, "wormhole.gif", append(marks, shape1, shape2)...); err != nil {
		panic(err)
	}
}

type RenderOptions struct {
	Eye    ln.Vector
	Center ln.Vector
	Up     ln.Vector

	Width  float64
	Height float64
	Fovy   float64
	Near   float64
	Far    float64
	Step   float64

	Frames int
	Angle  float64
}

func renderMovie(ctx context.Context, ops RenderOptions, outFile string, shapes ...ln.Shape) error {
	tmpDir := fmt.Sprintf("tmp-%s", randomString(16))

	if err := os.MkdirAll(tmpDir, os.ModeDir|0777); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			panic(err)
		}
	}()

	shiftAngle := ops.Angle / float64(ops.Frames)

	var wg sync.WaitGroup
	for i := 0; i < ops.Frames; i += 1 {
		fName := filepath.Join(tmpDir, fmt.Sprintf("frame-%03d.png", i))

		wg.Add(1)
		go func(i int, fName string) {
			defer wg.Done()
			defer log.Println("done", i)

			log.Println("start", i)

			m := ln.Rotate(ln.Vector{X: 0, Y: 0, Z: 1}, ln.Radians(float64(i)*shiftAngle))

			scene := ln.Scene{}
			for _, shape := range shapes {
				scene.Add(ln.NewTransformedShape(shape, m))
			}
			paths := scene.Render(ops.Eye, ops.Center, ops.Up, ops.Width, ops.Height, ops.Fovy, ops.Near, ops.Far, ops.Step)
			paths.WriteToPNG(fName, ops.Width, ops.Height)
		}(i, fName)
	}
	wg.Wait()

	log.Println(">>> Building movie")

	return convert(ctx, filepath.Join(tmpDir, "*.png"), outFile)
}

func convert(ctx context.Context, inFile, outFile string) error {
	//out, err := exec.Command("ffmpeg", "-i", "out%03d.png", "-y", "wormhole.gif").CombinedOutput()

	out, err := exec.CommandContext(ctx, "convert", "-delay", "2", "-loop", "0", inFile, outFile).CombinedOutput()

	log.Println(string(out))

	if err != nil {
		return err
	}

	if err := open.Run(outFile); err != nil {
		return err
	}

	return nil
}

func randomString(l int) (code string) {
	b := [1]byte{}
	for i := 0; i < l; i++ {
		if _, err := rand.Read(b[:]); err != nil {
			panic(err)
		}
		code += fmt.Sprint(int(b[0] % 10))
	}
	return
}
