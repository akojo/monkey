package object

type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType {
	return RETURN
}

func (r *Return) Inspect() string {
	return r.Value.Inspect()
}

func (r *Return) Equals(other Object) bool {
	return r == other
}
