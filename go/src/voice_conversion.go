package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	SamplingRate = 48000
	FrameSize    = 2048 //+ 128
	FrameShift   = FrameSize / 4

	ChannelSize = 200

	Volume = 0.01
)

var (
	input = bufio.NewScanner(os.Stdin)

	waiting4calcF0andMCEP   = make(chan *Frame, ChannelSize)
	waiting4adjustF0andMCEP = make(chan *Frame, ChannelSize)
	waiting4MLSAfilter      = make(chan *Frame, ChannelSize)
	waiting4outputSignal    = make(chan *Frame, ChannelSize)

	skippedFrame = 0
)

type Frame struct {
	wave            []float64 // 2フレーム分保存、前のフレームはF0用
	F0              float64
	mcep            []float64
	synthesizedWave []float64
}

func newFrame() *Frame {
	frame := Frame{}
	frame.wave = make([]float64, FrameSize*2)
	frame.F0 = 0.0
	frame.mcep = make([]float64, 1+MCEP_degree)
	frame.synthesizedWave = make([]float64, FrameShift)
	return &frame
}

func inputSignal(number int) []float64 {
	// 指定した数だけ入力波形を受け取る
	wave := make([]float64, number)
	for i := 0; i < number; i++ {
		if !input.Scan() {
			return wave
		}
		float_value, err := strconv.ParseFloat(input.Text(), 64)
		if err != nil {
			log.Print(err)
			return wave
		}
		wave[i] = float_value
	}
	return wave
}

func inputFrame() {
	// 入力波形からフレームを作成
	//入力の頭には余計な文字が含まれるため取り除く
	input.Scan() // Sampling_Frequency: 22050
	input.Scan() // press any key for finish

	// 入力波形を一時的に入れておく
	var waveBuffer_0 = make([]float64, FrameShift)
	var waveBuffer_1 = make([]float64, FrameShift)
	var waveBuffer_2 = make([]float64, FrameShift)
	var waveBuffer_3 = make([]float64, FrameShift)
	var waveBuffer_4 = make([]float64, FrameShift)
	var waveBuffer_5 = make([]float64, FrameShift)
	var waveBuffer_6 = make([]float64, FrameShift)
	var waveBuffer_7 = make([]float64, FrameShift)

	// フレーム
	var frame *Frame

	for {
		waveBuffer_0, waveBuffer_1, waveBuffer_2, waveBuffer_3, waveBuffer_4, waveBuffer_5, waveBuffer_6 = waveBuffer_1, waveBuffer_2, waveBuffer_3, waveBuffer_4, waveBuffer_5, waveBuffer_6, waveBuffer_7
		waveBuffer_7 = inputSignal(FrameShift)
		frame = newFrame()
		for i := 0; i < FrameShift; i++ {
			frame.wave[i] = waveBuffer_0[i]
			frame.wave[i+FrameShift] = waveBuffer_1[i]
			frame.wave[i+FrameShift*2] = waveBuffer_2[i]
			frame.wave[i+FrameShift*3] = waveBuffer_3[i]
			frame.wave[i+FrameShift*4] = waveBuffer_4[i]
			frame.wave[i+FrameShift*5] = waveBuffer_5[i]
			frame.wave[i+FrameShift*6] = waveBuffer_6[i]
			frame.wave[i+FrameShift*7] = waveBuffer_7[i]
		}
		waiting4calcF0andMCEP <- frame
	}
}

func outputSignal() {
	for {
		if len(waiting4outputSignal) > 1 {
			skippedFrame += len(waiting4outputSignal) / 2
		}
		for i := len(waiting4outputSignal) / 2; i > 0; i-- {
			_ = <-waiting4outputSignal
		}
		frame := <-waiting4outputSignal
		for i := 0; i < FrameShift; i += 1 {
			fmt.Println(frame.synthesizedWave[i] * Volume)
			//log.Println(frame.synthesizedWave[i] * Volume)
		}
	}
}

func checkChannelLen() {
	for {
		time.Sleep(1 * time.Second)
		log.Println(
			"waiting4calcF0andMCEP: ", len(waiting4calcF0andMCEP),
			", waiting4adjustF0andMCEP: ", len(waiting4adjustF0andMCEP),
			", waiting4MLSAfilter: ", len(waiting4MLSAfilter),
			", waiting4outputSignal: ", len(waiting4outputSignal),
			", skippedFrame:", skippedFrame,
			", skippedFrameTime:", float64(skippedFrame*FrameShift)/SamplingRate)
		skippedFrame = 0
	}
}

func api() {
	/*
		- 元のスペクトログラム
		- 変換グラフ
		- 修正したスペクトログラム

		- 元のF0
		- 係数
		- 変換後のF0

		- 元のmcep
		- 係数
		- 変換後のmcep

		- チャンネルの中の個数 確認用
	*/

	/*
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})

		e.Logger.Fatal(e.Start(":8080"))
	*/
	for {
		time.Sleep(10000 * time.Second)
	}
}

func main() {
	go inputFrame()
	go calcF0andMCEP()
	go adjustF0andMCEP()
	go MLSAfilter()
	go outputSignal()

	go checkChannelLen()

	api()
}
