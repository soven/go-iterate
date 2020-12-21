package iter

type doubleUintIterator struct {
	lhs, rhs UintIterator
	inRHS    bool
}

func (it *doubleUintIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUintIterator) Next() uint {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUintIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUintIterator combines all iterators to one.
func SuperUintIterator(itemList ...UintIterator) UintIterator {
	var super UintIterator = EmptyUintIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUintIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
