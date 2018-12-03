# GoGB

GoGB is a work-in-progress of a Gameboy emulator written in Go. The main objectives for this project are:
- build a medium-large sized project in Go to become familiar with the language and the design patterns associated with it.
- write an emulator that is performant enough to run [Super Mario Land](https://en.wikipedia.org/wiki/Super_Mario_Land) at a playable frame-rate.
- I am not planning on implementing sound for now.

# Todo
- [ ] create unit test suite for backend
    - [ ] instructions: arithmetic
    - [ ] instructions: memory
    - [ ] instructions: rotates
    - [ ] instructions: shifts
    - [ ] instructions: miscellaneous
    - [ ] CPU base
- [ ] implement basic graphics to be able to run test suite
    - [ ] learn how to draw graphics in Go + basic implementation
    - [ ] implement draw background
    - [ ] implement draw sprites
    - [ ] implement draw Window
- [ ] Advanced tests: get the backend to pass correctness and timing tests
- [ ] Robust graphics implementation
