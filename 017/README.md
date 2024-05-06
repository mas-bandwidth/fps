# 017

Go back to bare metal and see if there is a bottleneck in the golang processing of inputs with goroutines.

# Results

Everything looks fine!

I did notice that it is necessary to call runtime.Gosched() at the end of processing each input, as well as inside the loop that was generating player inputs.

This might be the thing that unlocks performance when run inside ring buffer processing?
