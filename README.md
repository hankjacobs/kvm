# kvm

A pure go implemention of the [KVM API](https://www.kernel.org/doc/Documentation/virtual/kvm/api.txt). 

## THIS IS A WORK IN PROGRESS

What does that mean?
- It probably doesn't work
- If it does, you probably shouldn't use it (yet)
- If you do, the API is subject to change and things will most likely break 
- I would love feedback around the design of the API 

The first goal of this project is to be able to run an extremely trivial VM. See this [blog post](https://david942j.blogspot.com/2018/10/note-learning-kvm-implement-your-own.html) for what I am using as a reference implementation and as a first goal. The `barevm` command uses the `kvm` package to build the VM outlined in the above blog post. 

Run using 

```bash
go run cmd/barevm/main.go
```