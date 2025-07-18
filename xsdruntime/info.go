package xsdruntime

type InfoDecl interface {
	Namespace() string
	NewInstance(name string) (any, error)
}
