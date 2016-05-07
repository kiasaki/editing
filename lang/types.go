package lang

type SValue interface {
	String() string
	Type() string
}

type SInt int64

type SNum float64

type SStr string

type SList struct {
	Head   *SListElement
	length int
}

func NewSList() *SList {
	return new(SList)
}

func (this *SList) Length() int {
	return this.length
}

type SListElement struct {
	Value SValue
	next  *SListElement
}

func (this *SListElement) Next() *SListElement {
	if this == nil {
		return nil
	}
	return this.next
}
