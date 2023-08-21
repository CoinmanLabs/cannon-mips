package main

import (
	"encoding/binary"
	"fmt"
	"strings"

	//"database/sql"
	"encoding/json"

	_ "github.com/lib/pq"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type State struct {
	Memory *Memory `json:"memory"`

	PreimageKey    common.Hash `json:"preimageKey"`
	PreimageOffset uint32      `json:"preimageOffset"` // note that the offset includes the 8-byte length prefix

	PC     uint32 `json:"pc"`
	NextPC uint32 `json:"nextPC"`
	LO     uint32 `json:"lo"`
	HI     uint32 `json:"hi"`
	Heap   uint32 `json:"heap"` // to handle mmap growth

	ExitCode uint8 `json:"exit"`
	Exited   bool  `json:"exited"`

	Step uint64 `json:"step"`

	Registers [32]uint32 `json:"registers"`

	// LastHint is optional metadata, and not part of the VM state itself.
	// It is used to remember the last pre-image hint,
	// so a VM can start from any state without fetching prior pre-images,
	// and instead just repeat the last hint on setup,
	// to make sure pre-image requests can be served.
	// The first 4 bytes are a uin32 length prefix.
	// Warning: the hint MAY NOT BE COMPLETE. I.e. this is buffered,
	// and should only be read when len(LastHint) > 4 && uint32(LastHint[:4]) >= len(LastHint[4:])
	LastHint hexutil.Bytes `json:"lastHint,omitempty"`
}

type traceState struct {
	Step   uint64 `json:"cycle"`
	PC     uint32 `json:"pc"`
	NextPC uint32 `json:"nextPC"`

	LO uint32 `json:"lo"`
	HI uint32 `json:"hi"`

	Registers [32]uint32 `json:"regs"`

	//PreimageKey   [32]byte `json:"preimageKey"`
	//PreimageOffset uint32      `json:"preimageOffset"` // note that the offset includes the 8-byte length prefix

	Heap uint32 `json:"heap"` // to handle mmap growth

	ExitCode uint8     `json:"exitCode"`
	Exited   bool      `json:"exited"`
	MemRoot  [32]uint8 `json:"memRoot"`

	Insn_proof   [28 * 32]uint8 `json:"insn_proof"`
	Memory_proof [28 * 32]uint8 `json:"mem_proof"`
}

func (s *traceState) insertToDB() {

	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	str := string(b)

	str = strings.ReplaceAll(str, ":", ":\"")
	str = strings.ReplaceAll(str, ",", "\",\"")
	str = strings.ReplaceAll(str, "]\"", "\"]")
	str = strings.ReplaceAll(str, "\"[", "[\"")
	str = strings.ReplaceAll(str, "\"\"", "\"")
	str = strings.ReplaceAll(str, "\"false\"", "false")
	str = strings.ReplaceAll(str, "\"true\"", "true")
	str = strings.ReplaceAll(str, "]}", "\"]}")

	fmt.Println(str)
}

func (s *State) EncodeWitness() []byte {
	out := make([]byte, 0)
	memRoot := s.Memory.MerkleRoot()
	out = append(out, memRoot[:]...)
	out = append(out, s.PreimageKey[:]...)
	out = binary.BigEndian.AppendUint32(out, s.PreimageOffset)
	out = binary.BigEndian.AppendUint32(out, s.PC)
	out = binary.BigEndian.AppendUint32(out, s.NextPC)
	out = binary.BigEndian.AppendUint32(out, s.LO)
	out = binary.BigEndian.AppendUint32(out, s.HI)
	out = binary.BigEndian.AppendUint32(out, s.Heap)
	out = append(out, s.ExitCode)
	if s.Exited {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	out = binary.BigEndian.AppendUint64(out, s.Step)
	for _, r := range s.Registers {
		out = binary.BigEndian.AppendUint32(out, r)
	}
	return out
}