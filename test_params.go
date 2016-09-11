package komblobulate

// Some implementations of KCodecParams:

type TestParams interface {
	KCodecParams
	GetResistType() byte
	GetCipherType() byte
}

type TestNullNullParams struct {
}

func (p *TestNullNullParams) GetRsParams() (int, int, int) {
	panic("Called GetRsParams without rs")
}

func (p *TestNullNullParams) GetAeadChunkSize() int {
	panic("Called GetAeadChunkSize without aead")
}

func (p *TestNullNullParams) GetAeadPassword() string {
	panic("Called GetAeadPassword() without aead")
}

func (p *TestNullNullParams) GetResistType() byte {
	return ResistType_None
}

func (p *TestNullNullParams) GetCipherType() byte {
	return CipherType_None
}

type TestNullAeadParams struct {
	ChunkSize int
	Password  string
}

func (p *TestNullAeadParams) GetRsParams() (int, int, int) {
	panic("Called GetRsParams without rs")
}

func (p *TestNullAeadParams) GetAeadChunkSize() int {
	return p.ChunkSize
}

func (p *TestNullAeadParams) GetAeadPassword() string {
	return p.Password
}

func (p *TestNullAeadParams) GetResistType() byte {
	return ResistType_None
}

func (p *TestNullAeadParams) GetCipherType() byte {
	return CipherType_Aead
}

type TestRsNullParams struct {
	DataPieceSize    int
	DataPieceCount   int
	ParityPieceCount int
}

func (p *TestRsNullParams) GetRsParams() (int, int, int) {
	return p.DataPieceSize, p.DataPieceCount, p.ParityPieceCount
}

func (p *TestRsNullParams) GetAeadChunkSize() int {
	panic("Called GetAeadChunkSize without aead")
}

func (p *TestRsNullParams) GetAeadPassword() string {
	panic("Called GetAeadPassword() without aead")
}

func (p *TestRsNullParams) GetResistType() byte {
	return ResistType_Rs
}

func (p *TestRsNullParams) GetCipherType() byte {
	return CipherType_None
}

type TestRsAeadParams struct {
	DataPieceSize    int
	DataPieceCount   int
	ParityPieceCount int
	ChunkSize        int
	Password         string
}

func (p *TestRsAeadParams) GetRsParams() (int, int, int) {
	return p.DataPieceSize, p.DataPieceCount, p.ParityPieceCount
}

func (p *TestRsAeadParams) GetAeadChunkSize() int {
	return p.ChunkSize
}

func (p *TestRsAeadParams) GetAeadPassword() string {
	return p.Password
}

func (p *TestRsAeadParams) GetResistType() byte {
	return ResistType_Rs
}

func (p *TestRsAeadParams) GetCipherType() byte {
	return CipherType_Aead
}
