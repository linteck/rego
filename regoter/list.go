package regoter

type IList[T any] interface {
	Add(e T) IList[T]
	Remove(e T) IList[T]
}
