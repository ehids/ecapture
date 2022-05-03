package user

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

type bashEvent struct {
	module IModule
	Pid    uint32
	Line   [80]uint8
	Retval uint32
	Comm   [16]byte
}

func (this *bashEvent) Decode(payload []byte) (err error) {
	buf := bytes.NewBuffer(payload)
	if err = binary.Read(buf, binary.LittleEndian, &this.Pid); err != nil {
		return
	}
	if err = binary.Read(buf, binary.LittleEndian, &this.Line); err != nil {
		return
	}
	if err = binary.Read(buf, binary.LittleEndian, &this.Retval); err != nil {
		return
	}
	if err = binary.Read(buf, binary.LittleEndian, &this.Comm); err != nil {
		return
	}

	return nil
}

func (this *bashEvent) String() string {
	var s string
	errno :=  GetModuleByName(MODULE_NAME_BASH).GetConf().(*BashConfig).ErrNo
	if(errno != -1){
		if errno >= 0 && errno < 128 &&uint32(errno) == this.Retval{
			s = fmt.Sprintf(fmt.Sprintf(" PID:%d, \tComm:%s, \tRetvalue:%d, \tLine:\n%s", this.Pid, this.Comm, this.Retval, unix.ByteSliceToString((this.Line[:]))))
		}
	}else{
		s = fmt.Sprintf(fmt.Sprintf(" PID:%d, \tComm:%s, \tRetvalue:%d, \tLine:\n%s", this.Pid, this.Comm, this.Retval, unix.ByteSliceToString((this.Line[:]))))
	}
	return s
}

func (this *bashEvent) StringHex() string {
	s := fmt.Sprintf(fmt.Sprintf(" PID:%d, \tComm:%s, \tRetvalue:%d, \tLine:\n%s,", this.Pid, this.Comm, this.Retval, dumpByteSlice([]byte(unix.ByteSliceToString((this.Line[:]))), "")))
	return s
}

func (this *bashEvent) SetModule(module IModule) {
	this.module = module
}

func (this *bashEvent) Module() IModule {
	return this.module
}

func (this *bashEvent) Clone() IEventStruct {
	return new(bashEvent)
}
