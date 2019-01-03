package backend

type Bus struct {
	allowanceChannel chan int      // used by PPU to signal to CPU how long to run
	cpuDoneChannel   chan struct{} // used by CPU to tell PPU it is done
}

func NewBus() *Bus {
	b := new(Bus)
	b.allowanceChannel = make(chan int)
	b.cpuDoneChannel = make(chan struct{})
	return b
}
