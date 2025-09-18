package thecatapi

type Breed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Client interface {
	ListBreeds() ([]Breed, error)
	ValidateBreed(nameOrID string) (bool, error)
}
