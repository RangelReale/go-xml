package xsdruntime

type InfoDecl interface {
	Namespace() string
	NewInstance(name string) (any, error)
}

type InstanceFactory struct {
	factories map[string]InfoDecl
}

func NewInstanceFactory() *InstanceFactory {
	return &InstanceFactory{
		factories: make(map[string]InfoDecl),
	}
}

func (f *InstanceFactory) Register(infoDecl ...InfoDecl) {
	for _, i := range infoDecl {
		f.factories[i.Namespace()] = i
	}
}

func (f *InstanceFactory) Get(ns string) (nf InfoDecl, ok bool) {
	nf, ok = f.factories[ns]
	return
}
