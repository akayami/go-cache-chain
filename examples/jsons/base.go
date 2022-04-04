package jsons

type BaseModelInterface interface {
	toJSON() ([]byte, error)
	fromJson([]byte) *BaseModel
}

type BaseModel struct {
}

func (b BaseModel) toJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b BaseModel) fromJson(bytes []byte) *BaseModel {
	//TODO implement me
	panic("implement me")
}
