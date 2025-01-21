package selectors

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/rabbitmq/amqp091-go"
)

type FilterExprSelector struct {
	program *vm.Program
}

func NewFilterExprSelector(filterExpr string) (*FilterExprSelector, error) {
	program, err := expr.Compile(filterExpr, expr.Env(DeliverySubset{}))
	if err != nil {
		return nil, err
	}
	return &FilterExprSelector{program: program}, nil
}

func (s *FilterExprSelector) IsSelected(msg amqp091.Delivery) (bool, error) {
	output, err := expr.Run(s.program, SubsetFromDelivery(msg))
	if err != nil {
		return false, err
	}
	isSelected, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("unknown output type: %T", output)
	}
	return isSelected, nil
}
