import numpy as np
import matplotlib.pyplot as plt

filename = "/Users/tata/src/voice_convertion/wav/fujitou_normal_signal/fujitou_normal_001.dat" # input()
filename = 'fujitou_normal_001_converted.dat'
with open(filename) as f:
    l = f.readlines()
y = [float(f) for f in l]
y_max = max(y)
y_min = min(y)
y = [(f-y_min)/(y_max-y_min) * 2-1 for f in y]
x = np.arange(0, len(y), 1)

x=x[:360000]
y=y[:360000]

plt.plot(x, y)
plt.show()