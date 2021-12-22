package deps

type Deps interface {
	FetchDependencies(directory string) ([]string, error)
}
