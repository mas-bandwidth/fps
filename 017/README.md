# 017

Go back to bare metal and see if there is a bottleneck in the golang processing of inputs with goroutines.

# Results

Everything looks fine!

```
2024/05/06 15:00:55 update 8: 225000 inputs processed on cpu #12
2024/05/06 15:00:55 update 8: 225000 inputs processed on cpu #7
2024/05/06 15:00:55 update 8: 225000 inputs processed on cpu #13
2024/05/06 15:00:55 update 8: 225000 inputs processed on cpu #1
2024/05/06 15:00:55 update 8: 225000 inputs processed on cpu #15
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #15
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #2
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #8
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #1
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #4
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #10
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #0
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #13
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #12
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #5
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #14
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #3
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #7
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #6
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #9
2024/05/06 15:00:56 update 9: 250000 inputs processed on cpu #11
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #1
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #7
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #12
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #0
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #2
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #8
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #6
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #3
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #11
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #13
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #14
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #5
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #10
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #15
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #4
2024/05/06 15:00:57 update 10: 275000 inputs processed on cpu #9
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #10
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #13
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #2
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #6
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #7
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #4
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #15
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #0
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #5
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #1
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #12
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #14
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #8
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #3
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #9
2024/05/06 15:00:58 update 11: 300000 inputs processed on cpu #11
```

I did notice that it is necessary to call runtime.Gosched() at the end of processing each input, as well as inside the loop that was generating player inputs.

This might be the thing that unlocks performance when run inside ring buffer processing?
