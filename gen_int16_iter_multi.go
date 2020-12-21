package iter

type doubleInt16Iterator struct {
	lhs, rhs Int16Iterator
	inRHS    bool
}

func (it *doubleInt16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt16Iterator) Next() int16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt16Iterator combines all iterators to one.
func SuperInt16Iterator(itemList ...Int16Iterator) Int16Iterator {
	var super Int16Iterator = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
