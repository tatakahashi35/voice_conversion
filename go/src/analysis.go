package main

import (
	"log"
	"math"

	world "github.com/r9y9/go-world"
	"github.com/r9y9/gossp/f0"
)

const (
	ScaledHammingWindowCoefA = 0.54
	ScaledHammingWindowCoefB = -0.46

	MCEP_degree = 400

	INF = math.MaxFloat64 / 10
)

var (
	scaledHammingWindow = make([]float64, FrameSize)

	f0_cnt    = 0
	f0_buffer = 0.0
	f0_pre    = 0.0

	yin           = f0.NewYIN(SamplingRate)
	f0_estimation = world.New(SamplingRate, 5)

	a_memo      = make([][]float64, 1+MCEP_degree)
	a_memo_flag = make([][]bool, 1+MCEP_degree)
	a_tilda     = make([]float64, 1+MCEP_degree)

	autocorrelations = make([]float64, 1+MCEP_degree)
	lpc              = make([]float64, 1+MCEP_degree)
	lpc_tmp          = make([]float64, 1+MCEP_degree)
	K                = 0.0
)

func init() {
	for i, _ := range scaledHammingWindow {
		scaledHammingWindow[i] = ScaledHammingWindowCoefA + ScaledHammingWindowCoefB*math.Cos(2*math.Pi*(float64(i)+1)/FrameSize)
	}

	for i := 0; i < 1+MCEP_degree; i++ {
		a_memo[i] = make([]float64, 1+MCEP_degree)
		a_memo_flag[i] = make([]bool, 1+MCEP_degree)
	}
}

func calcF0andMCEP() {
	chF0 := make(chan float64)
	chMCEP := make(chan []float64)
	for {
		frame := <-waiting4calcF0andMCEP
		go frame.calcF0(chF0)
		go frame.calcMCEP(chMCEP)
		frame.F0, frame.mcep = <-chF0, <-chMCEP
		waiting4adjustF0andMCEP <- frame
	}
}

func (frame *Frame) calcF0(ch chan float64) {
	if f0_cnt%4 != 0 {
		f0_cnt++
		ch <- f0_buffer
		return
	}
	f0_cnt = 1

	wave := make([]float64, FrameSize*2)
	for i, w := range frame.wave {
		wave[i] = w * 32768 * scaledHammingWindow[int(i/2)]
	}

	//f0, _ := yin.ComputeF0(w)
	f0s := f0.SWIPE(wave, SamplingRate, FrameSize*2/16, 50, 600)
	// _, f0 := f0_estimation.Dio(w, f0_estimation.NewDioOption())
	//log.Print(f0s)
	f0_sum := 0.0
	f0_count := 0
	for _, f0 := range f0s {
		if f0 > 0 {
			f0_sum += f0
			f0_count++
		}
	}
	if f0_count == 0 {
		f0_buffer = 0.0
	} else {
		f0_buffer = f0_sum / float64(f0_count)
	}

	log.Print(f0_buffer)
	ch <- f0_buffer
}

func (frame *Frame) calcMCEP(ch chan []float64) {
	wave := make([]float64, FrameSize)
	for i, w := range frame.wave[FrameSize:] {
		wave[i] = w * scaledHammingWindow[i]
	}
	frame.mel_cepstrum_from_LPC(&wave)
	ch <- frame.mcep
}

func (frame *Frame) mel_cepstrum_from_LPC(wave *[]float64) {
	Linear_Prediction_coefficient(wave)

	for i := 0; i < 1+MCEP_degree; i++ {
		for j := 0; j < 1+MCEP_degree; j++ {
			a_memo[i][j] = -INF
			a_memo_flag[i][j] = false
		}
	}
	for i := 0; i < 1+MCEP_degree; i++ {
		a_tilda[i] = a_i(i, 0)
	}
	if a_tilda[0] == 0.0 {
		return
	}

	K_tilda := K / a_tilda[0]
	for i := 1; i < 1+MCEP_degree; i++ {
		a_tilda[i] /= a_tilda[0]
	}

	frame.mcep[0] = math.Log(K_tilda)
	for i := 1; i < 1+MCEP_degree; i++ {
		frame.mcep[i] = -a_tilda[i]
		for j := 1; j < i; j++ {
			frame.mcep[i] -= frame.mcep[j] * a_tilda[i-j] * float64(j) / float64(i)
		}
	}

}

func Linear_Prediction_coefficient(wave *[]float64) {
	for i := 0; i < 1+MCEP_degree; i++ {
		autocorrelations[i] = autocorrelation(wave, i)
	}
	if autocorrelations[0] == 0.0 {
		return
	}

	lpc[0] = 1.0
	lpc[1] = -autocorrelations[1] / autocorrelations[0]
	lpc_err := autocorrelations[0] + lpc[1]*autocorrelations[1]
	lpc_lambda := 0.0
	for i := 1; i < MCEP_degree; i++ {
		lpc_lambda = 0.0
		for j := 0; j <= i; j++ {
			lpc_lambda += lpc[j] * autocorrelations[i+1-j]
		}
		lpc_lambda /= -lpc_err

		for j := 1; j <= i; j++ {
			lpc_tmp[j] = lpc[j] + lpc_lambda*lpc[i+1-j]
		}
		for j := 0; j < 1+MCEP_degree; j++ {
			lpc[j] = lpc_tmp[j]
		}

		lpc[0] = 1.0
		lpc[i+1] = lpc_lambda

		lpc_err = (1 - lpc_lambda*lpc_lambda) * lpc_err
	}
	K = math.Sqrt(lpc_err)
}

func autocorrelation(wave *[]float64, l int) float64 {
	value := 0.0
	for i, _ := range *wave {
		if i+l < FrameSize {
			value += (*wave)[i] * (*wave)[i+l]
		}
	}
	return value
}

func a_i(k, n int) float64 {
	if k > MCEP_degree || -n > MCEP_degree {
		return 0.0
	} else if a_memo_flag[k][-n] {
		return a_memo[k][-n]
	}

	if k > 1 {
		a_memo[k][-n] = a_i(k-1, n-1) + Alpha*(a_i(k, n-1)-a_i(k-1, n))
	} else if k == 1 {
		a_memo[1][-n] = (1-Alpha*Alpha)*a_i(0, n-1) + Alpha*a_i(1, n-1)
	} else if k == 0 {
		a_memo[0][-n] = lpc[-n] + Alpha*a_i(0, n-1)
	}

	a_memo_flag[k][-n] = true
	return a_memo[k][-n]
}
