package test

type Util struct {
	Mock *mock
}

func NewUtil() *Util {
	return &Util{
		Mock: &mock{},
	}
}
