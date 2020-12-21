package iter

type doubleRuneIterator struct {
	lhs, rhs RuneIterator
	inRHS    bool
}

func (it *doubleRuneIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleRuneIterator) Next() rune {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleRuneIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperRuneIterator combines all iterators to one.
func SuperRuneIterator(itemList ...RuneIterator) RuneIterator {
	var super RuneIterator = EmptyRuneIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleRuneIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
