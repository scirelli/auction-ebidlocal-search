package ebidlocal

//Searchable interface for searchable keywords.
type Searchable interface {
	Search() <-chan string
}

//SearchFunc func that implements the Search interface.
type SearchFunc func() <-chan string

func (s SearchFunc) Search() <-chan string {
	return s()
}
