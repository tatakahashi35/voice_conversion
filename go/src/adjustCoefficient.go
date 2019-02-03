package main

var (
	F0Coef    = 1.0
	mcepCoefs = make([]float64, 1+MCEP_degree)
)

func init() {
	F0Coef = 1.0 //1.0594630943592953
	for i, _ := range mcepCoefs {
		mcepCoefs[i] = 1.0
	}
}

func adjustF0andMCEP() {
	for {
		frame := <-waiting4adjustF0andMCEP
		if frame.F0 > 0 {
			frame.F0 *= F0Coef
		}
		for i, _ := range mcepCoefs {
			frame.mcep[i] *= mcepCoefs[i]
		}
		waiting4MLSAfilter <- frame
	}
}
