package cache

type Backend struct {
	name    string
	marshal bool
}

func (b Backend) GetName() string {
	return b.name
}

func (b Backend) IsMarshaled() bool {
	return b.marshal
}
