package server

type Watchlist struct {
	List []string `json:"list"`
	Name string   `json:"name"`
}

func (wl *Watchlist) IsValid() bool {
	return len(wl.List) > 0
}
