package cluster

type meta struct {
	data map[string][]byte
}

func NewMeta() *meta {
	return &meta{
		data: make(map[string][]byte),
	}
}

func (m *meta) GetNodeMeta(node string) []byte {
	if v, ok := m.data[node]; ok {
		return v
	}
	return nil
}

func (m *meta) SetNodeMeta(node string, v []byte) {
	m.data[node] = v
}
