package contracts

type Canonicalizer interface {
	Canonicalize(s string) string
}
