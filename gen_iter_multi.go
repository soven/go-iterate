// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type doubleIterator struct {
	lhs, rhs Iterator
	inRHS    bool
}

func (it *doubleIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIterator) Next() interface{} {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

func SuperIterator(itemList ...Iterator) Iterator {
	var super Iterator = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
