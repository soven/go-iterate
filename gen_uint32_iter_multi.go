package iter

type doubleUint32Iterator struct {
	lhs, rhs Uint32Iterator
	inRHS    bool
}

func (it *doubleUint32Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint32Iterator) Next() uint32 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint32Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint32Iterator combines all iterators to one.
func SuperUint32Iterator(itemList ...Uint32Iterator) Uint32Iterator {
	var super Uint32Iterator = EmptyUint32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint32Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
