package iter

type doubleStringIterator struct {
	lhs, rhs StringIterator
	inRHS    bool
}

func (it *doubleStringIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleStringIterator) Next() string {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleStringIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperStringIterator combines all iterators to one.
func SuperStringIterator(itemList ...StringIterator) StringIterator {
	var super StringIterator = EmptyStringIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleStringIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
