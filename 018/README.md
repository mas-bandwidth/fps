# 018

Run with XDP ring buffer again, this time regularly yielding with runtime.Gosched() and making sure the go workers are actually pinned to the right CPU.

Goal is to be able to handle 8k players per-bare metal machine with a 10G NIC.

# Results

...