package expressions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type ExpressionFactory func([]byte) (Expression, error)

type EvaluationFetchers struct {
	BodyFetcher       map[string]any
	QueryFetcher      func(string) string
	QueryArrayFetcher func(string) []string
	ParamFetcher      func(string) string
}

type Expression interface {
	Evaluate(fetchers EvaluationFetchers) any
	ReturnType() reflect.Kind
}

var ExpressionRegistry map[string]ExpressionFactory

func init() {
	ExpressionRegistry = map[string]ExpressionFactory{
		"AND":         andFactory,
		"OR":          orFactory,
		"NOT":         notFactory,
		"BODY":        bodyValueFactory,
		"QUERY":       queryValueFactory,
		"QUERY_ARRAY": queryArrayValueFactory,
		"PATH":        pathValueFactory,
		"STRING":      stringValueFactory,
		"EQUALS":      equalsFactory,
		"CONTAINS":    containsFactory,
	}
}

type AndExpression struct {
	expressions []Expression
}

func (e AndExpression) Evaluate(fetchers EvaluationFetchers) any {
	for _, expression := range e.expressions {
		if !expression.Evaluate(fetchers).(bool) {
			return false
		}
	}
	return true
}

func (e AndExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(true).Kind()
}

func andFactory(data []byte) (Expression, error) {
	body := parseJson(data)

	rawExpressions := body["expressions"]
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(rawExpressions, &rawMessages); err != nil {
		panic(err)
	}

	expressions := make([]Expression, len(rawMessages))

	for i, item := range rawMessages {
		expression, err := BuildExpression(item)
		if err != nil {
			return nil, err
		}
		if expression.ReturnType() != reflect.Bool {
			panic("invalid block: AND values must be bool")
		}
		expressions[i] = expression
	}

	return AndExpression{expressions: expressions}, nil
}

type OrExpression struct {
	expressions []Expression
}

func (e OrExpression) Evaluate(fetchers EvaluationFetchers) any {
	for _, expression := range e.expressions {
		if expression.Evaluate(fetchers).(bool) {
			return true
		}
	}
	return false
}

func (e OrExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(true).Kind()
}

func orFactory(data []byte) (Expression, error) {
	body := parseJson(data)

	rawExpressions := body["expressions"]
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(rawExpressions, &rawMessages); err != nil {
		panic(err)
	}

	expressions := make([]Expression, len(rawMessages))

	for i, item := range rawMessages {
		expression, err := BuildExpression(item)
		if err != nil {
			return nil, err
		}
		if expression.ReturnType() != reflect.Bool {
			panic("invalid block: OR values must be bool")
		}
		expressions[i] = expression
	}

	return OrExpression{expressions: expressions}, nil
}

type NotExpression struct {
	expression Expression
}

func (e NotExpression) Evaluate(fetchers EvaluationFetchers) any {
	result := e.expression.Evaluate(fetchers).(bool)
	return !result
}

func (e NotExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(true).Kind()
}

func notFactory(data []byte) (Expression, error) {
	body := parseJson(data)

	expression, err := BuildExpression(body["expression"])
	if err != nil {
		return nil, err
	}

	if expression.ReturnType() != reflect.Bool {
		panic("invalid block: NOT value must be bool")
	}

	return NotExpression{expression: expression}, nil
}

type ContainsExpression struct {
	list   Expression
	values []Expression
}

func (e ContainsExpression) Evaluate(fetchers EvaluationFetchers) any {
	listValues := e.list.Evaluate(fetchers).([]string)

	for _, value := range e.values {
		if !slices.Contains(listValues, value.Evaluate(fetchers).(string)) {
			return false
		}
	}
	return true
}

func (e ContainsExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(true).Kind()
}

func containsFactory(data []byte) (Expression, error) {
	body := parseJson(data)

	rawExpressions := body["values"]
	if rawExpressions == nil {
		panic("invalid block: CONTAINS must have values attribute")
	}
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(rawExpressions, &rawMessages); err != nil {
		panic(err)
	}

	expressions := make([]Expression, len(rawMessages))

	for i, item := range rawMessages {
		expression, err := BuildExpression(item)
		if err != nil {
			return nil, err
		}
		if expression.ReturnType() != reflect.String {
			panic("invalid block. CONTAINS values must be string")
		}
		expressions[i] = expression
	}

	rawList := body["list"]
	if rawList == nil {
		panic("invalid block: CONTAINS must have list attribute")
	}
	list, err := BuildExpression(rawList)

	if err != nil {
		return nil, err
	}

	if list.ReturnType() != reflect.Slice {
		panic("invalid block: CONTAINS list must be slice")
	}

	return ContainsExpression{list: list, values: expressions}, nil
}

type EqualsExpression struct {
	right Expression
	left  Expression
}

func (e EqualsExpression) Evaluate(fetchers EvaluationFetchers) any {
	switch e.right.ReturnType() {
	case reflect.String:
		{
			right := e.right.Evaluate(fetchers).(string)
			left := e.left.Evaluate(fetchers).(string)
			return right == left
		}
	case reflect.Slice:
		{
			right := e.right.Evaluate(fetchers).([]string)
			left := e.left.Evaluate(fetchers).([]string)
			return reflect.DeepEqual(right, left)
		}
	case reflect.Bool:
		{
			right := e.right.Evaluate(fetchers).(bool)
			left := e.left.Evaluate(fetchers).(bool)
			return right == left
		}
	default:
		panic("")
	}
}

func (e EqualsExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(true).Kind()
}

func equalsFactory(data []byte) (Expression, error) {
	body := parseJson(data)

	right, err := BuildExpression(body["right"])
	if err != nil {
		return nil, err
	}
	left, err := BuildExpression(body["left"])

	if err != nil {
		return nil, err
	}

	if right.ReturnType() != left.ReturnType() {
		panic("invalid blocks: EQUALS right and left must be the same kind")
	}

	return EqualsExpression{left: left, right: right}, nil
}

type BodyValueExpression struct {
	id string
}

func (e BodyValueExpression) Evaluate(fetchers EvaluationFetchers) any {
	value := fmt.Sprintf("%v", fetchers.BodyFetcher[e.id])
	return value

}

func (e BodyValueExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf("").Kind()
}

func bodyValueFactory(data []byte) (Expression, error) {
	body := parseJson(data)
	id := parseJsonString(body["id"])
	return BodyValueExpression{id: id}, nil
}

type QueryValueExpression struct {
	id string
}

func (e QueryValueExpression) Evaluate(fetchers EvaluationFetchers) any {
	value := fetchers.QueryFetcher(e.id)
	fmt.Println(value)
	return value

}

func (e QueryValueExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf("").Kind()
}

func queryValueFactory(data []byte) (Expression, error) {
	body := parseJson(data)
	id := parseJsonString(body["id"])
	return QueryValueExpression{id: id}, nil
}

type QueryArrayValueExpression struct {
	id string
}

func (e QueryArrayValueExpression) Evaluate(fetchers EvaluationFetchers) any {
	value := fetchers.QueryFetcher(e.id)
	if strings.Contains(value, ",") {
		return strings.Split(value, ",")
	}
	return fetchers.QueryArrayFetcher(e.id)
}

func (e QueryArrayValueExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf(make([]string, 0)).Kind()
}

func queryArrayValueFactory(data []byte) (Expression, error) {
	body := parseJson(data)
	id := parseJsonString(body["id"])
	return QueryArrayValueExpression{id: id}, nil
}

type PathValueExpression struct {
	id string
}

func (e PathValueExpression) Evaluate(fetchers EvaluationFetchers) any {
	value := fetchers.ParamFetcher(e.id)
	return value

}

func (e PathValueExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf("").Kind()
}

func pathValueFactory(data []byte) (Expression, error) {
	body := parseJson(data)
	id := parseJsonString(body["id"])
	return PathValueExpression{id: id}, nil
}

type StringValueExpression struct {
	value string
}

func (e StringValueExpression) Evaluate(fetchers EvaluationFetchers) any {
	return e.value
}

func (e StringValueExpression) ReturnType() reflect.Kind {
	return reflect.TypeOf("").Kind()
}

func stringValueFactory(data []byte) (Expression, error) {
	body := parseJson(data)
	value := parseJsonString(body["value"])

	return StringValueExpression{value: value}, nil
}

func BuildExpression(data []byte) (Expression, error) {
	var bodyRaw any
	if err := json.Unmarshal(data, &bodyRaw); err != nil {
		return nil, err
	}
	body := bodyRaw.(map[string]any)

	typ := fmt.Sprintf("%v", body["type"])
	factory := ExpressionRegistry[typ]
	expr, err := factory(data)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func parseJson(data []byte) map[string][]byte {
	var bodyRaw map[string]json.RawMessage
	if err := json.Unmarshal(data, &bodyRaw); err != nil {
		panic(err)
	}

	body := make(map[string][]byte, len(bodyRaw))
	for k, v := range bodyRaw {
		body[k] = []byte(v) // json.RawMessage is []byte
	}
	return body
}

func parseJsonString(data []byte) string {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		panic(err)
	}
	return s

}
