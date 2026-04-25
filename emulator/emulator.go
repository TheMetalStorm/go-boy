// Aims to be M-Cycle Accurate

package emulator

import (
	"go-boy/internal"
	"os"
	"strings"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	backend "github.com/micahke/glfw_imgui_backend"
	"github.com/micahke/imgui-go"
)

type Rom = internal.Rom
type Cpu = internal.Cpu

type Ppu = internal.Ppu

var MAX_CYCLES_PER_FRAME uint64 = 17556
var GB_CLOCK_SPEED_HZ uint64 = 4194304
var DIV_REG_INCREMENT_HZ = 16384

var ScreenSizeMultiplier int = 5

var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

var vertices = []float32{
	// Positions    // Texture Coords
	-1.0, 1.0, 0.0, 0.0, 0.0,
	-1.0, -1.0, 0.0, 0.0, 1.0,
	1.0, -1.0, 0.0, 1.0, 1.0,
	-1.0, 1.0, 0.0, 0.0, 0.0,
	1.0, -1.0, 0.0, 1.0, 1.0,
	1.0, 1.0, 0.0, 1.0, 0.0,
}

var vertexShaderSource = `
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

void main() {
    gl_Position = vec4(aPos, 1.0);
    TexCoord = aTexCoord;
}
` + "\x00"

var fragmentShaderSource = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D texture1;

void main() {
    FragColor = texture(texture1, TexCoord);
}
` + "\x00"

type Emulator struct {
	Cpu         *Cpu
	Ppu         *Ppu
	currentGame *Rom

	DoRender            bool
	ranMCyclesThisFrame uint64

	Window   *glfw.Window
	context  *imgui.Context
	Impl     *backend.ImguiGlfw3
	Vao, vbo uint32
	Program  uint32

	Io imgui.IO
}

func NewEmulator() *Emulator {
	emu := &Emulator{}
	emu.DoRender = true
	emu.Cpu = internal.NewCpu()
	emu.Ppu = internal.NewPpu(ScreenSizeMultiplier)

	emu.Ppu.HandleGLUpdate = true

	emu.Ppu.Cpu = emu.Cpu
	emu.Cpu.Ppu = emu.Ppu

	emu.Window = initOpenGl()

	emu.context, emu.Impl, emu.Io = initImgui(emu.Window)

	emu.SetupGL()
	emu.SetupDebugTextures()
	emu.Restart()

	return emu
}

func (e *Emulator) Delete() {
	e.Impl.Shutdown()
	e.context.Destroy()
	glfw.Terminate()
}

func initOpenGl() *glfw.Window {
	// Initialize GLFW through go-gl/glfw
	if err := glfw.Init(); err != nil {
		panic("Error initializing GLFW")
	}

	// GLFW setup
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Initialize window through go-gl/glfw

	//window, win_err := glfw.CreateWindow(GB_WINDOW_WIDTH*ScreenSizeMultiplier, GB_WINDOW_HEIGHT*ScreenSizeMultiplier, "Hello, world!", nil, nil)

	//For now we draw the BG only
	window, win_err := glfw.CreateWindow(internal.BG_WINDOW_X_Y*ScreenSizeMultiplier, internal.BG_WINDOW_X_Y*ScreenSizeMultiplier, "Background Layer", nil, nil)

	if win_err != nil {
		panic("Error creating window")
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		panic("Error initializing OpenGL")
	}
	return window

}

func initImgui(window *glfw.Window) (*imgui.Context, *backend.ImguiGlfw3, imgui.IO) {
	// Initialize imgui
	context := imgui.CreateContext(nil)

	io := imgui.CurrentIO()
	// io.AddFocusEvent(false)

	// KEY: link imgui context with GLFW window context
	impl := backend.ImguiGlfw3Init(window, io)
	return context, impl, io
}

func createShaderProgram(vSource, fSource string) uint32 {
	// Compile individual shaders
	vertexShader := compileShader(vSource+"\x00", gl.VERTEX_SHADER)
	fragmentShader := compileShader(fSource+"\x00", gl.FRAGMENT_SHADER)

	// Link shaders into a program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check for linking errors
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		panic("failed to link program: " + log)
	}

	// Shaders are now linked; we can delete the individual objects
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

func compileShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	// Check for compilation errors
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		panic("failed to compile shader: " + log)
	}

	return shader
}

func (e *Emulator) SetupGL() {
	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &e.Ppu.ViewPortTex)
	gl.BindTexture(gl.TEXTURE_2D, e.Ppu.ViewPortTex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// 2. Setup VAO and VBO
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coordinate attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	e.Vao = vao
	e.vbo = vbo

	e.Program = createShaderProgram(vertexShaderSource, fragmentShaderSource)

}

func (e *Emulator) SetupDebugTextures() {

	gl.GenTextures(1, &e.Ppu.BackgroundTex)
	gl.BindTexture(gl.TEXTURE_2D, e.Ppu.BackgroundTex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.GenTextures(1, &e.Ppu.WindowTex)
	gl.BindTexture(gl.TEXTURE_2D, e.Ppu.WindowTex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.GenTextures(1, &e.Ppu.TileViewerTex)
	gl.BindTexture(gl.TEXTURE_2D, e.Ppu.TileViewerTex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	e.Window.MakeContextCurrent()
}

func (e *Emulator) SetWindow(window *glfw.Window) {
	e.Window = window
}

func (e *Emulator) Restart() {
	e.Cpu.Restart()
	e.Ppu.Restart(ScreenSizeMultiplier)
	e.ranMCyclesThisFrame = 0

	e.Ppu.Cpu = e.Cpu
	e.Cpu.Ppu = e.Ppu
	e.Cpu.Memory.Ppu = e.Ppu

	//e.currentGame = internal.NewRom("./games/Dr.M.gb")
	e.currentGame = internal.NewRom("./games/Tetris.gb")

	e.LoadRom(e.currentGame)
}

func (e *Emulator) LoadRom(r *Rom) {
	for i := 0x0; i <= 0x7fff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		e.Cpu.Memory.SetValueForRom(uint16(i), newVal)
	}
}

func (e *Emulator) RunTests(tests []string) {

	var startNext bool = false
	go changeBool(&startNext)
	for _, test := range tests {
		startNext = false
		e.Restart()
		e.currentGame = internal.NewRom(test)
		e.LoadRom(e.currentGame)
		for {
			if startNext {
				println()
				break
			}
			e.SerialOut()
			e.Step()
		}
	}
	os.Exit(0)

}

func changeBool(startNextTest *bool) {
	for range time.Tick(time.Second * 2) {
		println("")
		*startNextTest = true
	}
}

func (e *Emulator) Run() {
	for !e.Window.ShouldClose() {
		if e.Ppu.HandleGLUpdate {
			glfw.PollEvents()
		}
		e.SerialOut()
		e.Step()
	}
}

func (e *Emulator) Render() {
	if e.Ppu.HandleGLUpdate {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		e.Ppu.Render()

		gl.UseProgram(e.Program)
		gl.BindVertexArray(e.Vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		e.Window.SwapBuffers()
	} else {
		//delegate rendering to Debugger to handle ALL Graphics Operation in a single place
		e.DoRender = true

	}
}

func (e *Emulator) Step() {

	ranMCyclesThisStep := uint64(1)
	ranMCyclesThisStep += e.handleInterrupts()

	if !e.Cpu.Halt {
		ranMCyclesThisStep += e.Cpu.Step()
	}
	e.Cpu.UpdateTimers(ranMCyclesThisStep)

	e.Ppu.Step(ranMCyclesThisStep)

	e.ranMCyclesThisFrame += ranMCyclesThisStep

	if e.ShouldRender() {
		e.Render()
		e.FinishFrame()
	}

}

func (e *Emulator) ShouldRender() bool {
	return e.ranMCyclesThisFrame >= MAX_CYCLES_PER_FRAME
}

func (e *Emulator) FinishFrame() {
	e.ranMCyclesThisFrame = 0
	time.Sleep(time.Second / 60)
}

func (e *Emulator) handleInterrupts() uint64 {
	requestedInterrupts := e.Cpu.Memory.Io.GetIF()
	enabledInterrupts := e.Cpu.Memory.GetIe()
	activeInterrupts := requestedInterrupts & enabledInterrupts & 0x1f
	if activeInterrupts != 0 {
		e.Cpu.Halt = false
	}

	if e.Cpu.IME {
		if activeInterrupts != 0 {
			e.Cpu.SP--
			e.Cpu.Memory.SetValue(e.Cpu.SP, internal.GetHigher8(e.Cpu.PC))
			e.Cpu.SP--
			e.Cpu.Memory.SetValue(e.Cpu.SP, internal.GetLower8(e.Cpu.PC))

			if e.Cpu.Memory.GetInterruptEnabledBit(internal.VBLANK) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.VBLANK) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.VBLANK, false)
				e.Cpu.PC = 0x0040
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.LCD) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.LCD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.LCD, false)
				e.Cpu.PC = 0x0048
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.TIMER) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.TIMER) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.TIMER, false)
				e.Cpu.PC = 0x0050
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.SERIAL) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.SERIAL) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.SERIAL, false)
				e.Cpu.PC = 0x0058
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.JOYPAD) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.JOYPAD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.JOYPAD, false)
				e.Cpu.PC = 0x0060
			}
			e.Cpu.IME = false
			return 5
		}
	}
	return 0
}

func (e *Emulator) SerialOut() {
	read, _ := e.Cpu.Memory.ReadByteAt(0xff02)
	if read == 0x81 {
		ch, _ := e.Cpu.Memory.ReadByteAt(0xff01)
		print(string(ch))
		e.Cpu.Memory.SetValue(0xff02, 0x00)

	}
}

func (e *Emulator) GetCurrentGame() []byte {
	return e.currentGame.GetData()
}
