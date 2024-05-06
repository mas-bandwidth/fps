# 017

Go back to bare metal and see if there is a bottleneck in the golang processing of inputs with goroutines.

# Results

Everything looks fine!

```
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #18
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #17
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #25
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #3
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #22
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #19
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #9
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #5
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #7
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #10
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #8
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #6
2024/05/06 15:29:44 update 134: 3375000 inputs processed on cpu #27
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #3
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #17
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #1
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #0
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #2
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #4
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #15
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #28
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #11
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #12
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #19
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #22
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #25
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #14
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #30
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #13
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #31
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #26
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #16
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #23
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #20
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #18
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #24
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #29
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #21
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #10
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #9
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #7
2024/05/06 15:29:45 update 135: 3400000 inputs processed on cpu #5
```

That's 8000 players processing player inputs at 100HZ across 32 worker CPUs.

<img width="802" alt="image" src="https://github.com/mas-bandwidth/fps/assets/696656/d41cc5ac-7d90-415d-972e-f8c803ea7220">

I did notice that it is necessary to call runtime.Gosched() at the end of processing each input, as well as inside the loop that was generating player inputs.

This might be the thing that unlocks performance when run inside ring buffer processing?
