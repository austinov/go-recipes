package loader

type Loader interface {
	Start() error
	Stop()
}
