package iter

type doubleByteIterator struct {
	lhs, rhs ByteIterator
	inRHS    bool
}

func (it *doubleByteIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleByteIterator) Next() byte {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleByteIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperByteIterator combines all iterators to one.
func SuperByteIterator(itemList ...ByteIterator) ByteIterator {
	var super ByteIterator = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleByteIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
