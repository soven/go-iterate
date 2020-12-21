package iter

type doubleUint16Iterator struct {
	lhs, rhs Uint16Iterator
	inRHS    bool
}

func (it *doubleUint16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint16Iterator) Next() uint16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint16Iterator combines all iterators to one.
func SuperUint16Iterator(itemList ...Uint16Iterator) Uint16Iterator {
	var super Uint16Iterator = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
