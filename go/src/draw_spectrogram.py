import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd

SamplingRate = 48000
FrameSize    = 2048
FrameShift   = FrameSize // 4

def get_wave(filename):
    with open(filename) as f:
        l = f.readlines()
    return np.array([float(d) for d in l])

def calc_spectrogram(wave):
    spectrum_amps = []
    for start in range(0, len(wave)-FrameSize+1, FrameShift):
        target_wave = wave[start:start + FrameSize] * np.hamming(FrameSize)
        spectrum = np.fft.fft(target_wave)[:FrameSize//2]
        spectrum_amp = np.abs(spectrum)
        spectrum_amps.append(spectrum_amp)
    return np.array(spectrum_amps)

def draw_spectrogram(spectrogram):
    frames = (len(wave)-FrameSize+1) / FrameShift
    times = np.arange(frames) * FrameShift / SamplingRate
    freqs = np.arange(FrameSize // 2) * SamplingRate / FrameSize
    print(times)
    print(freqs)
    spectrogram_df = pd.DataFrame(data=spectrogram)

    plt.figure(figsize=(20, 6))
    sns.heatmap(data=np.log(spectrogram_df[list(range(len(spectrogram_df.columns)//2-1, -1, -1))].T),
                xticklabels=100, 
                yticklabels=100, 
                cmap=plt.cm.gist_rainbow_r,
                )
    plt.show()

if __name__ == "__main__":
    filename = input()
    print(filename)

    wave = get_wave(filename)
    spectrogram = calc_spectrogram(wave)
    draw_spectrogram(spectrogram)
