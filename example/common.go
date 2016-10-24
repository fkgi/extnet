package example

type IPList string

func (l *IPList) String() string {
	return string(*l)
}

func (l *IPList) Set(s string) error {
	*l = IPList(string(*l) + "/" + s)
	return nil
}
