package builder

type IServiceAccountBuilder interface {
	Builder
}

type BaseServiceAccountBuilder struct {
	BaseResourceBuilder
}
