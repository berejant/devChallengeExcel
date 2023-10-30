package contracts

type CellSerializer interface {
	Marshal(key string, value string) []byte
	Unmarshal([]byte) (key string, value string, err error)
}
