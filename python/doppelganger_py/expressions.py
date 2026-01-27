"""Expression evaluation system for matching request parameters."""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Callable
import re


@dataclass
class EvaluationFetchers:
    """Container for functions that fetch values from the request."""
    body_fetcher: dict[str, Any]
    query_fetcher: Callable[[str], str]
    query_array_fetcher: Callable[[str], list[str]]
    param_fetcher: Callable[[str], str]


class Expression(ABC):
    """Base class for all expressions."""
    
    @abstractmethod
    def evaluate(self, fetchers: EvaluationFetchers) -> Any:
        """Evaluate the expression and return a result."""
        pass
    
    @abstractmethod
    def return_type(self) -> type:
        """Return the type of the evaluation result."""
        pass


class AndExpression(Expression):
    """Logical AND of multiple boolean expressions."""
    
    def __init__(self, expressions: list[Expression]):
        self.expressions = expressions
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        for expr in self.expressions:
            if not expr.evaluate(fetchers):
                return False
        return True
    
    def return_type(self) -> type:
        return bool


class OrExpression(Expression):
    """Logical OR of multiple boolean expressions."""
    
    def __init__(self, expressions: list[Expression]):
        self.expressions = expressions
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        for expr in self.expressions:
            if expr.evaluate(fetchers):
                return True
        return False
    
    def return_type(self) -> type:
        return bool


class NotExpression(Expression):
    """Logical NOT of a boolean expression."""
    
    def __init__(self, expression: Expression):
        self.expression = expression
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        return not self.expression.evaluate(fetchers)
    
    def return_type(self) -> type:
        return bool


class ContainsExpression(Expression):
    """Check if a list contains all specified values."""
    
    def __init__(self, list_expr: Expression, values: list[Expression]):
        self.list_expr = list_expr
        self.values = values
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        list_values = self.list_expr.evaluate(fetchers)
        for value_expr in self.values:
            if value_expr.evaluate(fetchers) not in list_values:
                return False
        return True
    
    def return_type(self) -> type:
        return bool


class EqualsExpression(Expression):
    """Check if two expressions are equal."""
    
    def __init__(self, left: Expression, right: Expression):
        self.left = left
        self.right = right
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        left_val = self.left.evaluate(fetchers)
        right_val = self.right.evaluate(fetchers)
        return left_val == right_val
    
    def return_type(self) -> type:
        return bool


class RegexExpression(Expression):
    """Check if a value matches a regex pattern."""
    
    def __init__(self, value: Expression, pattern: str):
        self.value = value
        self.pattern = pattern
        self._compiled = re.compile(pattern)
    
    def evaluate(self, fetchers: EvaluationFetchers) -> bool:
        value = self.value.evaluate(fetchers)
        return bool(self._compiled.search(value))
    
    def return_type(self) -> type:
        return bool


class BodyValueExpression(Expression):
    """Get a value from the request body."""
    
    def __init__(self, id: str):
        self.id = id
    
    def evaluate(self, fetchers: EvaluationFetchers) -> str:
        value = fetchers.body_fetcher.get(self.id)
        return str(value) if value is not None else ""
    
    def return_type(self) -> type:
        return str


class QueryValueExpression(Expression):
    """Get a value from query parameters."""
    
    def __init__(self, id: str):
        self.id = id
    
    def evaluate(self, fetchers: EvaluationFetchers) -> str:
        return fetchers.query_fetcher(self.id)
    
    def return_type(self) -> type:
        return str


class QueryArrayValueExpression(Expression):
    """Get an array value from query parameters."""
    
    def __init__(self, id: str):
        self.id = id
    
    def evaluate(self, fetchers: EvaluationFetchers) -> list[str]:
        value = fetchers.query_fetcher(self.id)
        if "," in value:
            return value.split(",")
        return fetchers.query_array_fetcher(self.id)
    
    def return_type(self) -> type:
        return list


class PathValueExpression(Expression):
    """Get a value from path parameters."""
    
    def __init__(self, id: str):
        self.id = id
    
    def evaluate(self, fetchers: EvaluationFetchers) -> str:
        return fetchers.param_fetcher(self.id)
    
    def return_type(self) -> type:
        return str


class StringValueExpression(Expression):
    """A literal string value."""
    
    def __init__(self, value: str):
        self.value = value
    
    def evaluate(self, fetchers: EvaluationFetchers) -> str:
        return self.value
    
    def return_type(self) -> type:
        return str


# Expression factory registry
EXPRESSION_REGISTRY: dict[str, Callable[[dict], Expression]] = {}


def register_expression(type_name: str):
    """Decorator to register an expression factory."""
    def decorator(factory: Callable[[dict], Expression]):
        EXPRESSION_REGISTRY[type_name] = factory
        return factory
    return decorator


@register_expression("AND")
def and_factory(data: dict) -> Expression:
    expressions = [build_expression(expr) for expr in data["expressions"]]
    for expr in expressions:
        if expr.return_type() != bool:
            raise ValueError("AND values must be bool")
    return AndExpression(expressions)


@register_expression("OR")
def or_factory(data: dict) -> Expression:
    expressions = [build_expression(expr) for expr in data["expressions"]]
    for expr in expressions:
        if expr.return_type() != bool:
            raise ValueError("OR values must be bool")
    return OrExpression(expressions)


@register_expression("NOT")
def not_factory(data: dict) -> Expression:
    expression = build_expression(data["expression"])
    if expression.return_type() != bool:
        raise ValueError("NOT value must be bool")
    return NotExpression(expression)


@register_expression("CONTAINS")
def contains_factory(data: dict) -> Expression:
    if "values" not in data:
        raise ValueError("CONTAINS must have values attribute")
    if "list" not in data:
        raise ValueError("CONTAINS must have list attribute")
    
    values = [build_expression(v) for v in data["values"]]
    for v in values:
        if v.return_type() != str:
            raise ValueError("CONTAINS values must be string")
    
    list_expr = build_expression(data["list"])
    if list_expr.return_type() != list:
        raise ValueError("CONTAINS list must be list")
    
    return ContainsExpression(list_expr, values)


@register_expression("EQUALS")
def equals_factory(data: dict) -> Expression:
    right = build_expression(data["right"])
    left = build_expression(data["left"])
    
    if right.return_type() != left.return_type():
        raise ValueError("EQUALS right and left must be the same type")
    
    return EqualsExpression(left, right)


@register_expression("REGEX")
def regex_factory(data: dict) -> Expression:
    value = build_expression(data["value"])
    pattern = data["pattern"]
    
    if value.return_type() != str:
        raise ValueError("REGEX value must be string")
    
    return RegexExpression(value, pattern)


@register_expression("BODY")
def body_factory(data: dict) -> Expression:
    return BodyValueExpression(data["id"])


@register_expression("QUERY")
def query_factory(data: dict) -> Expression:
    return QueryValueExpression(data["id"])


@register_expression("QUERY_ARRAY")
def query_array_factory(data: dict) -> Expression:
    return QueryArrayValueExpression(data["id"])


@register_expression("PATH")
def path_factory(data: dict) -> Expression:
    return PathValueExpression(data["id"])


@register_expression("STRING")
def string_factory(data: dict) -> Expression:
    return StringValueExpression(data["value"])


def build_expression(data: dict) -> Expression:
    """Build an expression from a dictionary."""
    type_name = data.get("type")
    if type_name not in EXPRESSION_REGISTRY:
        raise ValueError(f"Unknown expression type: {type_name}")
    
    factory = EXPRESSION_REGISTRY[type_name]
    return factory(data)
