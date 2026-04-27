package app

type Necessity int

const (
	NecessityNone Necessity = iota
	NecessityRequired
	NecessityRecommended
	NecessityOptional
)
