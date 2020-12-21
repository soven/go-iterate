package iter

type doubleInt32Iterator struct {
	lhs, rhs Int32Iterator
	inRHS    bool
}

func (it *doubleInt32Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt32Iterator) Next() int32 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt32Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt32Iterator combines all iterators to one.
func SuperInt32Iterator(itemList ...Int32Iterator) Int32Iterator {
	var super Int32Iterator = EmptyInt32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt32Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
