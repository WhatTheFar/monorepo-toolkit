package core

// Hash SHA1 hashed content
type Hash string

type GitGateway interface {
	DiffNameOnly(from, to Hash) ([]string, error)
}
