package logrec

type Field struct {
	Key   string
	Value interface{}
}

func WithField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}
