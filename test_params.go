package komblobulate

// Some implementations of KCodecParams:

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
