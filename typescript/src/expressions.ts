/**
 * Expression evaluation system for matching request parameters.
 */

// Type for fetcher functions
export interface EvaluationFetchers {
  bodyFetcher: Record<string, unknown>;
  queryFetcher: (key: string) => string;
  queryArrayFetcher: (key: string) => string[];
  paramFetcher: (key: string) => string;
}

// Return type enum
export type ReturnType = 'boolean' | 'string' | 'array';

// Base expression interface
export interface Expression {
  evaluate(fetchers: EvaluationFetchers): unknown;
  returnType(): ReturnType;
}

// Expression data from JSON
export interface ExpressionData {
  type: string;
  [key: string]: unknown;
}

// Expression factory type
type ExpressionFactory = (data: ExpressionData) => Expression;

// Expression registry
const expressionRegistry: Map<string, ExpressionFactory> = new Map();

/**
 * Register an expression factory
 */
function registerExpression(typeName: string, factory: ExpressionFactory): void {
  expressionRegistry.set(typeName, factory);
}

/**
 * AND Expression - All expressions must be true
 */
class AndExpression implements Expression {
  constructor(private readonly expressions: Expression[]) {}

  evaluate(fetchers: EvaluationFetchers): boolean {
    for (const expr of this.expressions) {
      if (!expr.evaluate(fetchers)) {
        return false;
      }
    }
    return true;
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * OR Expression - At least one expression must be true
 */
class OrExpression implements Expression {
  constructor(private readonly expressions: Expression[]) {}

  evaluate(fetchers: EvaluationFetchers): boolean {
    for (const expr of this.expressions) {
      if (expr.evaluate(fetchers)) {
        return true;
      }
    }
    return false;
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * NOT Expression - Negates a boolean expression
 */
class NotExpression implements Expression {
  constructor(private readonly expression: Expression) {}

  evaluate(fetchers: EvaluationFetchers): boolean {
    return !this.expression.evaluate(fetchers);
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * CONTAINS Expression - Check if list contains all specified values
 */
class ContainsExpression implements Expression {
  constructor(
    private readonly listExpr: Expression,
    private readonly values: Expression[]
  ) {}

  evaluate(fetchers: EvaluationFetchers): boolean {
    const listValues = this.listExpr.evaluate(fetchers) as string[];
    for (const valueExpr of this.values) {
      const value = valueExpr.evaluate(fetchers) as string;
      if (!listValues.includes(value)) {
        return false;
      }
    }
    return true;
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * EQUALS Expression - Compare two values for equality
 */
class EqualsExpression implements Expression {
  constructor(
    private readonly left: Expression,
    private readonly right: Expression
  ) {}

  evaluate(fetchers: EvaluationFetchers): boolean {
    const leftVal = this.left.evaluate(fetchers);
    const rightVal = this.right.evaluate(fetchers);

    if (Array.isArray(leftVal) && Array.isArray(rightVal)) {
      if (leftVal.length !== rightVal.length) {
        return false;
      }
      return leftVal.every((val, idx) => val === rightVal[idx]);
    }

    return leftVal === rightVal;
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * REGEX Expression - Match a value against a regex pattern
 */
class RegexExpression implements Expression {
  private readonly regex: RegExp;

  constructor(
    private readonly value: Expression,
    pattern: string
  ) {
    this.regex = new RegExp(pattern);
  }

  evaluate(fetchers: EvaluationFetchers): boolean {
    const value = this.value.evaluate(fetchers) as string;
    return this.regex.test(value);
  }

  returnType(): ReturnType {
    return 'boolean';
  }
}

/**
 * BODY Expression - Get a value from the request body
 */
class BodyValueExpression implements Expression {
  constructor(private readonly id: string) {}

  evaluate(fetchers: EvaluationFetchers): string {
    const value = fetchers.bodyFetcher[this.id];
    return value !== undefined && value !== null ? String(value) : '';
  }

  returnType(): ReturnType {
    return 'string';
  }
}

/**
 * QUERY Expression - Get a value from query parameters
 */
class QueryValueExpression implements Expression {
  constructor(private readonly id: string) {}

  evaluate(fetchers: EvaluationFetchers): string {
    return fetchers.queryFetcher(this.id);
  }

  returnType(): ReturnType {
    return 'string';
  }
}

/**
 * QUERY_ARRAY Expression - Get an array value from query parameters
 */
class QueryArrayValueExpression implements Expression {
  constructor(private readonly id: string) {}

  evaluate(fetchers: EvaluationFetchers): string[] {
    const value = fetchers.queryFetcher(this.id);
    if (value.includes(',')) {
      return value.split(',');
    }
    return fetchers.queryArrayFetcher(this.id);
  }

  returnType(): ReturnType {
    return 'array';
  }
}

/**
 * PATH Expression - Get a value from path parameters
 */
class PathValueExpression implements Expression {
  constructor(private readonly id: string) {}

  evaluate(fetchers: EvaluationFetchers): string {
    return fetchers.paramFetcher(this.id);
  }

  returnType(): ReturnType {
    return 'string';
  }
}

/**
 * STRING Expression - A literal string value
 */
class StringValueExpression implements Expression {
  constructor(private readonly value: string) {}

  evaluate(_fetchers: EvaluationFetchers): string {
    return this.value;
  }

  returnType(): ReturnType {
    return 'string';
  }
}

// Register all expression factories
registerExpression('AND', (data: ExpressionData): Expression => {
  const expressionsData = data.expressions as ExpressionData[];
  const expressions = expressionsData.map((expr) => buildExpression(expr));
  for (const expr of expressions) {
    if (expr.returnType() !== 'boolean') {
      throw new Error('AND values must be boolean');
    }
  }
  return new AndExpression(expressions);
});

registerExpression('OR', (data: ExpressionData): Expression => {
  const expressionsData = data.expressions as ExpressionData[];
  const expressions = expressionsData.map((expr) => buildExpression(expr));
  for (const expr of expressions) {
    if (expr.returnType() !== 'boolean') {
      throw new Error('OR values must be boolean');
    }
  }
  return new OrExpression(expressions);
});

registerExpression('NOT', (data: ExpressionData): Expression => {
  const expressionData = data.expression as ExpressionData;
  const expression = buildExpression(expressionData);
  if (expression.returnType() !== 'boolean') {
    throw new Error('NOT value must be boolean');
  }
  return new NotExpression(expression);
});

registerExpression('CONTAINS', (data: ExpressionData): Expression => {
  if (!data.values) {
    throw new Error('CONTAINS must have values attribute');
  }
  if (!data.list) {
    throw new Error('CONTAINS must have list attribute');
  }

  const valuesData = data.values as ExpressionData[];
  const values = valuesData.map((v) => buildExpression(v));
  for (const v of values) {
    if (v.returnType() !== 'string') {
      throw new Error('CONTAINS values must be string');
    }
  }

  const listExpr = buildExpression(data.list as ExpressionData);
  if (listExpr.returnType() !== 'array') {
    throw new Error('CONTAINS list must be array');
  }

  return new ContainsExpression(listExpr, values);
});

registerExpression('EQUALS', (data: ExpressionData): Expression => {
  const right = buildExpression(data.right as ExpressionData);
  const left = buildExpression(data.left as ExpressionData);

  if (right.returnType() !== left.returnType()) {
    throw new Error('EQUALS right and left must be the same type');
  }

  return new EqualsExpression(left, right);
});

registerExpression('REGEX', (data: ExpressionData): Expression => {
  const value = buildExpression(data.value as ExpressionData);
  const pattern = data.pattern as string;

  if (value.returnType() !== 'string') {
    throw new Error('REGEX value must be string');
  }

  return new RegexExpression(value, pattern);
});

registerExpression('BODY', (data: ExpressionData): Expression => {
  return new BodyValueExpression(data.id as string);
});

registerExpression('QUERY', (data: ExpressionData): Expression => {
  return new QueryValueExpression(data.id as string);
});

registerExpression('QUERY_ARRAY', (data: ExpressionData): Expression => {
  return new QueryArrayValueExpression(data.id as string);
});

registerExpression('PATH', (data: ExpressionData): Expression => {
  return new PathValueExpression(data.id as string);
});

registerExpression('STRING', (data: ExpressionData): Expression => {
  return new StringValueExpression(data.value as string);
});

/**
 * Build an expression from JSON data
 */
export function buildExpression(data: ExpressionData): Expression {
  const typeName = data.type;
  const factory = expressionRegistry.get(typeName);

  if (!factory) {
    throw new Error(`Unknown expression type: ${typeName}`);
  }

  return factory(data);
}
