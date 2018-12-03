package main

func main() {

	// driver.Main(backend.Run)

	// c := backend.NewCPU()
	// d := backend.NewDebugHarness()

	// for c.PC != 0x55 {
	// 	c.DecodeAndExecuteNext()
	// }

	// // r := bufio.NewReader(os.Stdin)

	// for i := 0; i < 150; i++ {
	// 	c.DecodeAndExecuteNext()
	// 	d.PrintDebug(c)

	// 	// r.ReadBytes('\n')
	// }

	// c := backend.NewTestCPU()
	// c.PC = 0x100
	// c.SP = 0xFFFE

	// dat, err := ioutil.ReadFile("rom/cpu_instrs/individual/09-op r,r.gb")
	// if err != nil {
	// 	panic(err)
	// }
	// c.LoadToRAM(dat)

	// ram := c.GetRAM()
	// fmt.Printf("%X", ram[0x0100:0x104])

	// for i := 0; i < 3; i++ {
	// 	d.PrintDebug(c)
	// 	c.DecodeAndExecuteNext()
	// }

	// for i := 0; i < 1000; i++ {
	// 	c.DecodeAndExecuteNext()
	// 	d.PrintDebug(c)
	// }

	// p := backend.NewPPU(c)
	// p.DrawFrame()

}
