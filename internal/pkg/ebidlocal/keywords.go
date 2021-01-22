package ebidlocal

//Keywords list of keywords to search open auctions for.
type Keywords []string

//Search search all open auctions for the list of keywords.
func (kw Keywords) Search() <-chan string {
	var results = make(chan string)
	go SearchAuctions(kw, results)
	return results
}
