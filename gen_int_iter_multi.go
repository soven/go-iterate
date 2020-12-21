package iter

type doubleIntIterator struct {
	lhs, rhs IntIterator
	inRHS    bool
}

func (it *doubleIntIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIntIterator) Next() int {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIntIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperIntIterator combines all iterators to one.
func SuperIntIterator(itemList ...IntIterator) IntIterator {
	var super IntIterator = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIntIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
