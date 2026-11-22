package engine

import (
	"fmt"
	"strconv"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func evaluate(config node.Values) (bool, error) {
	left, _ := config["left"].(string)
	op, _ := config["op"].(string)
	right, _ := config["right"].(string)

	switch op {
	case "equal":
		return left == right, nil
	case "not_equal":
		return left != right, nil
	case "greater_than", "less_than":
		l, errL := strconv.ParseFloat(left, 64)
		r, errR := strconv.ParseFloat(right, 64)
		if errL != nil || errR != nil {
			return false, fmt.Errorf("%s needs two numbers, got %q and %q", op, left, right)
		}
		if op == "greater_than" {
			return l > r, nil
		}
		return l < r, nil
	default:
		return false, fmt.Errorf("unknown operator %q", op)
	}
}
