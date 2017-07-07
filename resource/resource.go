package resource



type Reader interface {
	Read (id string) []byte
}

