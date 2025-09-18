package thecatapi

type Mock struct {
	Breeds []Breed
	Err    error
}

func (m *Mock) ListBreeds() ([]Breed, error) { return m.Breeds, m.Err }
func (m *Mock) ValidateBreed(s string) (bool, error) {
	if m.Err != nil {
		return false, m.Err
	}
	for _, b := range m.Breeds {
		if b.Name == s || b.ID == s {
			return true, nil
		}
	}
	return false, nil
}
