.code16
.globl _start

_start:
mov $0x61, %al
mov $0x217, %dx
out %al, (%dx)
mov $10, %al
out %al, (%dx)
hlt
