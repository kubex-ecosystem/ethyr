package genkit

// Vector representa a identidade e o estado atômico de uma tarefa de IA.
// Layout de 64 bits:
// [ 8 bits: Status/Flags ] [ 24 bits: Job Context/Type ] [ 32 bits: Registry ID ]
type Vector uint64

const (
	// Máscaras de Extração
	MaskRegistryID Vector = 0xFFFFFFFF         // Bits 0-31
	MaskJobContext Vector = 0xFFFFFF00000000   // Bits 32-55
	MaskFlags      Vector = 0xFF00000000000000 // Bits 56-63

	// Bit Flags (Máquinas de Estado / Status)
	FlagPending Vector = 1 << 56
	FlagSuccess Vector = 1 << 57
	FlagTimeout Vector = 1 << 58
	FlagError   Vector = 1 << 59
)

// RegistryID extrai o ID do Data Plane
func (v Vector) RegistryID() uint32 { return uint32(v & MaskRegistryID) }

// JobContext extrai o tipo/contexto do Job
func (v Vector) JobContext() uint32 { return uint32((v & MaskJobContext) >> 32) }

// HasFlag verifica se o status está ativo
func (v Vector) HasFlag(flag Vector) bool { return (v & flag) != 0 }

// WithFlag limpa as flags antigas e define a nova
func (v Vector) WithFlag(flag Vector) Vector { return (v &^ MaskFlags) | flag }
