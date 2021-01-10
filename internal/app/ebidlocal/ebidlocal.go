package ebidlocal

func New(config Config) *Ebidlocal {
	return &Ebidlocal{
		Config: config,
	}
}

type Ebidlocal struct {
	Config Config
}

func (e *Ebidlocal) CreateUser(username string) *Ebidlocal {
	return e
}
