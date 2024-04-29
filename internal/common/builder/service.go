package builder

type IServiceBuilder interface {
	Builder
}

type ServiceBuilder struct {
	BaseResourceBuilder
}
