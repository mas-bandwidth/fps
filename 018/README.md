# 018

Go back to google cloud, make sure to call runtime.Gosched() frequently to yield, and call taskset when launching the worker thread, instead of trying to pin to CPU after launch.

I believe this will fix the pathological performance on google cloud, which was caused by lack of yielding in cooperative scheduling on the one CPU, and contention because workers were not being pinned to the correct CPU.

# Results

...