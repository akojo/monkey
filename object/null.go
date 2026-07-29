package object

type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL
}

func (n *Null) Inspect() string {
	return "null"
}

func (n *Null) Equals(other Object) bool {
	return n == other
}
