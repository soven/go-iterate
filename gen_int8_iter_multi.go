package iter

type doubleInt8Iterator struct {
	lhs, rhs Int8Iterator
	inRHS    bool
}

func (it *doubleInt8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt8Iterator) Next() int8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt8Iterator combines all iterators to one.
func SuperInt8Iterator(itemList ...Int8Iterator) Int8Iterator {
	var super Int8Iterator = EmptyInt8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
