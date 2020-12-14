// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type doubleUint64Iterator struct {
	lhs, rhs Uint64Iterator
	inRHS    bool
}

func (it *doubleUint64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint64Iterator) Next() uint64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

func SuperUint64Iterator(itemList ...Uint64Iterator) Uint64Iterator {
	var super Uint64Iterator = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}