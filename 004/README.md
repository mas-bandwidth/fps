# 004

Minimal repro for what seems to be a stack smash in libxdp with the ring buffer:

https://github.com/xdp-project/xdp-tools/issues/422

UPDATE: Appears to be fixed if I install the latest libxdp and libbpf from source.