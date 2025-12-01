package bind

import (
	"errors"
	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func Bind(data *data.ASTData, spec *specir.SpecModule) (*data.ASTData, error) {
	return data, errors.New("Bind(data *data.ASTData, spec specir.SpecModule) not implemented")
}
