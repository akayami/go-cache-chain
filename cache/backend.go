package cache

type Backend struct {
	name string
}

func (b Backend) GetName() string {
	return b.name
}
