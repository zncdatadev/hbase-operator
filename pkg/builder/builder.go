package builder

type ResourceBuilder[T any] interface {
	Build() *T
	Name() string
	Namespace() string
}

type T any

var _ ResourceBuilder[T] = &Resource{}

type Resource struct {
}

func (r *Resource) Build() *T {
	panic("implement me")
}

func (r *Resource) Name() string {
	panic("implement me")
}

func (r *Resource) Namespace() string {
	panic("implement me")
}
