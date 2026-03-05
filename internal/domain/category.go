package domain

// Category is a lookup entity for requests.
//
// In DB it's a simple table with an integer ID (SERIAL).
type Category struct {
	ID   int
	Name string
	Slug string
}

