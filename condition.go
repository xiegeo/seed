package seed

// Condition represents boolean operations on branch nodes and value to boolean operations on leaves.
//
// The root condition can use any operator except PushUp. PushUp sudo condition exist only in the physical
// representation of tree, not in the logical.
type Condition struct {
	Op         Op // condition operator
	Children   []Condition
	FieldPaths []Path
	Literal    any // if null, the literal is skipped
}

// Path is a list of code names to walk through references and fields
type Path []CodeName

type _nextFieldState int

const (
	_literalState _nextFieldState = iota
	_fieldPathState
	_childrenState
)

// MakeDirectedCondition create a Condition tree from a list of operands for a directional operator.
// If types match, MakeDirectedCondition sees the operand as a child condition or field path,
// otherwise the operand is a literal. Operands should not contain any nil pointers.
// MakeDirectedCondition encapsulate the order of operands so Condition builders don't need to
// worry about it.
//
// Both []Condition and Condition matches Children. Both []Path and Path matches FieldPaths
func MakeDirectedCondition(op Op, operands ...any) (Condition, error) {
	root := Condition{
		Op: op,
	}
	var state _nextFieldState
	for i := len(operands) - 1; i >= 0; i-- { // consume operands in reverse
		switch vt := operands[i].(type) {
		default: // literal
			// future: check for unsupported literals
			if state == _literalState {
				root.Literal = vt
				state = _fieldPathState // only one literal can be contained
			} else {
				root.Children = append(root.Children, Condition{Op: PushUp, Literal: vt})
				state = _childrenState // all operands must now go in to child nodes
			}
		case Path: // one field path
			if state == _childrenState {
				root.Children = append(root.Children, Condition{Op: PushUp, FieldPaths: []Path{vt}})
			} else {
				root.FieldPaths = append(root.FieldPaths, vt)
				state = _fieldPathState
			}
		case []Path: // many field paths
			if state == _childrenState {
				root.Children = append(root.Children, Condition{Op: PushUp, FieldPaths: vt})
			} else {
				root.FieldPaths = append(root.FieldPaths, vt...)
				state = _fieldPathState
			}
		case Condition: // one child
			root.Children = append(root.Children, vt)
			state = _childrenState
		case []Condition: // many children
			root.Children = append(root.Children, vt...)
			state = _childrenState
		}
	}
	// operands of the same type where appended in reverse order
	reverseSlice(root.Children)
	reverseSlice(root.FieldPaths)
	return root, nil
}

func reverseSlice[T any](s []T) {
	for i := len(s) / 2; i > 0; {
		mirror := len(s) - i
		i--
		s[i], s[mirror] = s[mirror], s[i]
	}
}

// ForEach loops through each operand to call a function to handle each case.
// `PushUp` sudo conditions are recursed so the condition handler function will never see a `PushUp`,
// while the invariant Op of c applies to all handler function calls.
// The literal handler function will never be called if Literal is the unset nil.
func (c Condition) ForEach(cf func(Condition), pf func([]CodeName), lf func(any)) {
	for _, child := range c.Children {
		if child.Op == PushUp {
			child.ForEach(cf, pf, lf)
		} else {
			cf(child)
		}
	}
	for _, path := range c.FieldPaths {
		pf(path)
	}
	if c.Literal != nil {
		lf(c.Literal)
	}
}

// Op defines condition operators. All Op, except PushUp, has an inverse.
// A large number of operators are defined for ease of writing and manipulating conditions.
// When realizing conditions, conditions should be simplified to match closely with underling implementation.
type Op uint8

const (
	// Push up elements to the parent Condition, useful for custom ordering of operands.
	// When evaluated, PushUp can not be the top condition's operator.
	PushUp Op = iota

	// Unidirectional boolean operators.
	//
	// Care is taken to give sensible results when there are less operands than common definition to decrease
	// special case handling code. (|E| stands for the number of operands)
	//
	// Conceptually, removing an operand from `And` should only move the statement in the true direction.
	// Well removing an operand from `Or` should only move the statement in the false direction.
	//
	// When used to concatenate variable conditions, `And` is often used to concatenate filters,
	// without any filter defined, all value should be returned. So filter should be true for all values.
	// `Or` is often used to concatenate authorisation rules from roles. When no role is found,
	// user should not have access to any data.
	//
	// Implantation wise, `And` can be check by returning false on first false operand encountered,
	// true on exhaustion (unless null is also encountered).
	// `Or` can be check by returning true on first true operand encountered, false on exhaustion.
	And  // all operands are true, if |E| = 0, always true
	Nand // if |E| = 0, always false
	Or   // one or more operands is true, if |E| = 0, always false
	Nor  // if |E| = 0, always true
	Eq   // all operands are equal, if |E| < 2, always true
	Neq  // if |E| < 2, always false

	// A nil Literal can't be used, so use *Null for comparison against nil values.
	// When other operators encounter nil values, it is viewed as unknown.
	// ie: Eq(nil, nil) = nil; And(true, nil) = true; And(false, nil) = nil.
	AndIsNull  // true iff all values are nil.    If |E| = 0, always true.  Inverse of OrNotNull
	OrIsNull   // true iff any values are nil.    If |E| = 0, always false. Inverse of AndNotNull
	AndNotNull // true iff no  values are nil.    If |E| = 0, always true.  Inverse of OrIsNull
	OrNotNull  // true iff any values is not nil. If |E| = 0, always false. Inverse of AndIsNull

	// Set operators work on sets, `In` does intersects and return true iff intersect is not empty.
	// Operands that are not sets are automatically promoted to a set containing it self.
	In    // if |E| = 0, always true. But if any operand is empty, always false.
	NotIn // inverse of In

	// Directional operators apply Children first, then FieldPaths, and Literal last.
	// All operands must be of the same orderable type to be comparable.
	// If |E| > 2, such as {a<b<c}, it is considered equivalent to {(a<b)&(b<c)}.
	// Even though {a<b} is inversed of {a>=b}, this is not true for {a<b<c} vs {a>=b>=c}.
	// For example: {1<3<2} and {1>=3>=2} are both false. So directional operators get their own
	// inverse N* operators instead of using the more common conversions.
	//
	// For Lt,Lte,Gt,and Gte; if |E| < 2, always true. Meaning operand list is total or partial ordered.
	// For their inverses (N*); if |E| < 2, always false.
	Lt
	Nlt
	Lte
	Nlte
	Gt
	Ngt
	Gte
	Ngte

	OpMax = Ngte
)

var _opString = [OpMax + 1]string{
	"PushUp", "And", "Nand", "Or", "Nor", "Eq", "Neq", "AndIsNull", "OrIsNull", "AndNotNull", "OrNotNull",
	"In", "NotIn", "Lt", "Nlt", "Lte", "NLte", "Gt", "Ngt", "Gte", "Ngte",
}

func (op Op) String() string {
	return _opString[op]
}

// Operator to inverse index, left shifted by one. Panic on PushUp.
var _opInverse = [OpMax]Op{
	Nand, And, Nor, Or, Neq, Eq, OrNotNull, AndNotNull, OrIsNull, AndIsNull,
	NotIn, In, Nlt, Lt, Nlte, Lte, Ngt, Gt, Ngte, Gte,
}

func (op Op) Inverse() Op {
	return _opInverse[op-1]
}
