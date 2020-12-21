package iter

type doubleUint8Iterator struct {
	lhs, rhs Uint8Iterator
	inRHS    bool
}

func (it *doubleUint8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint8Iterator) Next() uint8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint8Iterator combines all iterators to one.
func SuperUint8Iterator(itemList ...Uint8Iterator) Uint8Iterator {
	var super Uint8Iterator = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
