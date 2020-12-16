package resembled

type doublePrefixIterator struct {
	lhs, rhs PrefixIterator
	inRHS    bool
}

func (it *doublePrefixIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doublePrefixIterator) Next() Type {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doublePrefixIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperPrefixIterator combines all iterators to one.
func SuperPrefixIterator(itemList ...PrefixIterator) PrefixIterator {
	var super PrefixIterator = EmptyPrefixIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doublePrefixIterator{lhs: itemList[i], rhs: super}
	}
	return super
}
