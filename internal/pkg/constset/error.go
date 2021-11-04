package constset

type BeeError struct {
	Message string
	Code    string
}

func (b *BeeError) Error() string {
	return b.Message
}
