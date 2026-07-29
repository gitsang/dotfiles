package repository

// Collection holds repository implementations owned by the petstore application.
type Collection struct {
	Echo *EchoRepository
}

func NewCollection() *Collection {
	return &Collection{
		Echo: NewEchoRepository(),
	}
}
