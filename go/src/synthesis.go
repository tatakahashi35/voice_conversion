package main

import (
	"math"
	"math/rand"
)

const (
	Alpha = 0.554
	/*
		16000 Hz: 0.41
		22050 Hz: 0.455
		44100 Hz: 0.544
		48000 Hz: 0.554
	*/
	PADE          = 4
	Interpolation = 1

	RandSeed = 1
)

var (
	PADE_Coefs      = []float64{0.4999273, 0.1067005, 0.01170221, 0.0005656279}
	pre_wave_output = []float64{}
)

type Excite struct {
	mcep_degree     int64
	p1, p2, pc, inc float64
	exciteWave      []float64
}

type BaseFilter struct {
	mcep_degree int64
	delay       []float64
	tmp         []float64
}

type ExpBaseFilter struct {
	baseFilters []BaseFilter
	mcep_degree int64
}

type MLSAFilter struct {
	expBaseFilterFirst  ExpBaseFilter
	expBaseFilterSecond ExpBaseFilter
	mcep1               []float64
	b1                  []float64
	mcep2               []float64
	b2                  []float64
	inc                 []float64
	output              []float64
}

// Excite
func newExcite() *Excite {
	excite := Excite{}
	excite.p1 = 0.0
	excite.p2 = 0.0
	excite.pc = 0.0
	excite.inc = 0.0
	excite.exciteWave = make([]float64, FrameShift)
	return &excite
}

func (excite *Excite) exciteInit(input float64) {
	excite.p1 = input
	excite.pc = excite.p1
}

func (excite *Excite) output(input float64) {
	excite.p2 = input
	if excite.p1 != 0.0 && excite.p2 != 0.0 {
		excite.inc = (excite.p2 - excite.p1) * Interpolation * 1.0 / FrameShift
	} else {
		excite.inc = 0.0
		excite.pc = excite.p2
		excite.p1 = 0.0
	}

	i := (Interpolation + 1) / 2.0
	for j := 0; j < FrameShift; j++ {
		if excite.p1 == 0.0 {
			excite.exciteWave[j] = float64(rand.Int()%3 - 1) // -1, 0, 1
		} else {
			excite.pc += 1.0
			if excite.pc >= excite.p1 {
				excite.exciteWave[j] = math.Sqrt(excite.p1)
				excite.pc -= excite.p1
			} else {
				excite.exciteWave[j] = 0.0
			}
		}

		if i == 1 {
			excite.p1 += excite.inc
			i = Interpolation
		} else {
			i--
		}
	}
	excite.p1 = excite.p2
}

// BaseFilter
func newBaseFilter(mcep_degree int64) BaseFilter {
	baseFilter := BaseFilter{}
	baseFilter.mcep_degree = mcep_degree
	baseFilter.delay = make([]float64, baseFilter.mcep_degree+1)
	baseFilter.tmp = make([]float64, baseFilter.mcep_degree+1)
	return baseFilter
}

func (baseFilter *BaseFilter) setFirstDelay(x float64) {
	baseFilter.delay[0] = (1-Alpha*Alpha)*x + Alpha*baseFilter.delay[0]
}

func (baseFilter *BaseFilter) getValue(fisrt bool, b []float64) float64 {
	output := 0.0

	baseFilter.tmp[1] = baseFilter.delay[0]
	for i := 2; i <= int(baseFilter.mcep_degree); i++ {
		baseFilter.tmp[i] = baseFilter.delay[i-1] + Alpha*(baseFilter.delay[i]-baseFilter.tmp[i-1])
	}
	for i := 1; i <= int(baseFilter.mcep_degree); i++ {
		baseFilter.delay[i] = baseFilter.tmp[i]
	}

	if fisrt == true {
		output += b[1] * baseFilter.delay[1]
	}
	for i := 2; i <= int(baseFilter.mcep_degree); i++ {
		output += b[i] * baseFilter.delay[i]
	}
	return output
}

// ExpBaseFilter
func newExpBaseFilter(mcep_degree int64) ExpBaseFilter {
	expBaseFilter := ExpBaseFilter{}
	expBaseFilter.mcep_degree = mcep_degree
	expBaseFilter.baseFilters = make([]BaseFilter, PADE)
	for i := 0; i < PADE; i++ {
		expBaseFilter.baseFilters[i] = newBaseFilter(expBaseFilter.mcep_degree)
	}
	return expBaseFilter
}

func (expBaseFilter *ExpBaseFilter) getValue(first bool, x float64, b []float64) float64 {
	output := 0.0
	nextFirstDelay := 0.0
	for i := PADE - 1; i >= 0; i-- {
		v := expBaseFilter.baseFilters[i].getValue(first, b)
		if i+1 < PADE {
			expBaseFilter.baseFilters[i+1].setFirstDelay(v)
		}
		v = PADE_Coefs[i] * v
		output += v

		if i%2 == 1 {
			v = -v
		}
		nextFirstDelay += v
	}
	expBaseFilter.baseFilters[0].setFirstDelay(nextFirstDelay + x)
	return x + output + nextFirstDelay
}

//MLSAFilter
func newMLSAFilter() *MLSAFilter {
	MLSAFilter := MLSAFilter{}
	MLSAFilter.expBaseFilterFirst = newExpBaseFilter(1)
	MLSAFilter.expBaseFilterSecond = newExpBaseFilter(MCEP_degree)
	MLSAFilter.mcep1 = make([]float64, MCEP_degree+1)
	MLSAFilter.b1 = make([]float64, MCEP_degree+1)
	MLSAFilter.mcep2 = make([]float64, MCEP_degree+1)
	MLSAFilter.b2 = make([]float64, MCEP_degree+1)
	MLSAFilter.inc = make([]float64, MCEP_degree+1)
	MLSAFilter.output = make([]float64, FrameShift)
	return &MLSAFilter
}

func (mlsaFilter *MLSAFilter) mlsaFilterInit(mcep_input []float64) {
	for i := 0; i < MCEP_degree+1; i++ {
		mlsaFilter.mcep1[i] = mcep_input[i]
	}
	mlsaFilter.calc_b1()
}

func (mlsaFilter *MLSAFilter) setMCEP(mcep_input []float64) {
	for i := 0; i < MCEP_degree+1; i++ {
		mlsaFilter.mcep2[i] = mcep_input[i]
	}
	mlsaFilter.calc_b2()
}

func (mlsaFilter *MLSAFilter) calc_b1() {
	mlsaFilter.b1[MCEP_degree+1-1] = mlsaFilter.mcep1[MCEP_degree+1-1]
	for i := MCEP_degree - 1; i >= 0; i-- {
		mlsaFilter.b1[i] = mlsaFilter.mcep1[i] - Alpha*mlsaFilter.b1[i+1]
	}
}

func (mlsaFilter *MLSAFilter) calc_b2() {
	mlsaFilter.b2[MCEP_degree+1-1] = mlsaFilter.mcep2[MCEP_degree+1-1]
	for i := MCEP_degree - 1; i >= 0; i-- {
		mlsaFilter.b2[i] = mlsaFilter.mcep2[i] - Alpha*mlsaFilter.b2[i+1]
	}
}

func (mlsaFilter *MLSAFilter) mlsaOutput(input, mcep []float64) {
	mlsaFilter.setMCEP(mcep)

	for i := 0; i < MCEP_degree+1; i++ {
		mlsaFilter.inc[i] = (mlsaFilter.b2[i] - mlsaFilter.b1[i]) * Interpolation * 1.0 / FrameShift
	}
	i := (Interpolation + 1) / 2
	for j := 0; j < FrameShift; j++ {
		x := input[j] * math.Exp(mlsaFilter.b1[0])
		x = mlsaFilter.expBaseFilterFirst.getValue(true, x, mlsaFilter.b1)
		x = mlsaFilter.expBaseFilterSecond.getValue(false, x, mlsaFilter.b1)
		mlsaFilter.output[j] = x
		i--
		if i == 0 {
			for i := 0; i <= MCEP_degree; i++ {
				mlsaFilter.b1[i] += mlsaFilter.inc[i]
			}
			i = Interpolation
		}
	}
}

func init() {
	rand.Seed(RandSeed)
}

func (frame *Frame) F02Pitch() float64 {
	if frame.F0 == 0 {
		return 0.0
	}
	return SamplingRate / frame.F0
}

func (frame *Frame) mlsaFilteredWave(excite *Excite, mlsaFilter *MLSAFilter) {
	pitch := frame.F02Pitch()
	excite.output(pitch)
	mlsaFilter.mlsaOutput(excite.exciteWave, frame.mcep)
	///
	// a := mlsaFilter.output
	/*
		if frame.F0 != 0 {
			for i := 0; i < len(mlsaFilter.output); i++ {
				if i-int(SamplingRate/frame.F0/(2*math.Pi)-1) >= 0 {
					mlsaFilter.output[i] += -0.8 * mlsaFilter.output[i-int(SamplingRate/frame.F0/(2*math.Pi)-1)]
				} else if FrameShift+i-int(SamplingRate/frame.F0/(2*math.Pi)-1) >= 0 {
					mlsaFilter.output[i] += -0.8 * pre_wave_output[FrameShift+i-int(SamplingRate/frame.F0/(2*math.Pi)-1)]
				}
			}
		}
		pre_wave_output = mlsaFilter.output
	*/
	///
	frame.synthesizedWave = mlsaFilter.output
}

func MLSAfilter() {
	excite := newExcite()
	mlsaFilter := newMLSAFilter()

	frame := <-waiting4MLSAfilter
	pitch := frame.F02Pitch()
	excite.exciteInit(pitch)
	mlsaFilter.mlsaFilterInit(frame.mcep)
	frame.mlsaFilteredWave(excite, mlsaFilter)

	for {
		frame := <-waiting4MLSAfilter
		frame.mlsaFilteredWave(excite, mlsaFilter)
		waiting4outputSignal <- frame
	}
}
