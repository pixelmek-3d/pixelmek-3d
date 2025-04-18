package game

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/common"
)

type BenchmarkHandler struct {
	file      *os.File
	writer    *bufio.Writer
	stopwatch *common.Stopwatch
	tick      uint64
}

func NewBenchmarkHandler() *BenchmarkHandler {
	file, _ := os.Create("benchmark_" + strconv.Itoa(os.Getpid()) + ".csv")
	writer := bufio.NewWriter(file)

	b := &BenchmarkHandler{
		file:      file,
		writer:    writer,
		stopwatch: &common.Stopwatch{},
		tick:      0,
	}

	// write header
	b.write("Tick,Duration(ns),TPS,FPS\n")

	return b
}

func (b *BenchmarkHandler) write(s string) {
	if b.writer == nil {
		return
	}
	b.writer.WriteString(s)
}

func (b *BenchmarkHandler) Close() {
	if b.writer != nil {
		b.writer.Flush()
		b.writer = nil
	}
	if b.file != nil {
		b.file.Close()
	}
}

func (b *BenchmarkHandler) UpdateStart() {
	b.tick++
	b.stopwatch.Start()
}

func (b *BenchmarkHandler) UpdateStop() {
	// record update duration and current FPS/TPS
	duration := b.stopwatch.Stop()
	fps, tps := ebiten.ActualFPS(), ebiten.ActualTPS()

	line := fmt.Sprintf("%d,%d,%0.1f,%0.1f\n", b.tick, duration.Nanoseconds(), tps, fps)
	b.write(line)
}
