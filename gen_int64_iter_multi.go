package iter

type doubleInt64Iterator struct {
	lhs, rhs Int64Iterator
	inRHS    bool
}

func (it *doubleInt64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt64Iterator) Next() int64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt64Iterator combines all iterators to one.
func SuperInt64Iterator(itemList ...Int64Iterator) Int64Iterator {
	var super Int64Iterator = EmptyInt64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}
