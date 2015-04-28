package cqrs

// Helper struct for aggregates to better express sets of links

var EmptyStruct = struct{}{}

func NewInt64Set(d ...int64) *Int64Set {
	s := make(map[int64]struct{})
	for _, i := range d {
		s[i] = EmptyStruct
	}
	return &Int64Set{
		Set: s,
	}
}

type Int64Set struct {
	Set map[int64]struct{}
}

func (s *Int64Set) EqualsSet(t *Int64Set) bool {
	for k, _ := range t.Set {
		if _, ok := s.Set[k]; !ok {
			return false
		}
	}
	return true
}

func (s *Int64Set) Equals(t ...int64) bool {
	for _, i := range t {
		if _, ok := s.Set[i]; !ok {
			return false
		}
	}
	return true
}

func (s *Int64Set) Add(i int64) bool {
	if _, ok := s.Set[i]; !ok {
		s.Set[i] = struct{}{}
		return true
	}
	return false // False means it existed already
}

func (s *Int64Set) Remove(i int64) bool {
	if _, ok := s.Set[i]; ok {
		delete(s.Set, i)
		return true
	}
	return false // False means it wasn't removed
}

func (s *Int64Set) ToSlice() []int64 {
	r := make([]int64, 0, len(s.Set))
	for k, _ := range s.Set {
		r = append(r, k)
	}
	return r
}

func (s *Int64Set) DiffSet(t ...int64) (l *Int64Set, c *Int64Set, r *Int64Set) {
	l = NewInt64Set()
	c = NewInt64Set()
	r = NewInt64Set()
	for _, v := range t {
		if _, ok := s.Set[v]; !ok {
			r.Add(v) // Not found in t so add to left
		} else {
			c.Add(v) // Found in both so add to common
		}
	}
	for k, _ := range s.Set {
		if _, ok := c.Set[k]; !ok { // Not found in s so add to right
			l.Add(k)
		}
	}
	return
}

func (s *Int64Set) Diff(t ...int64) (l []int64, c []int64, r []int64) {
	t_l, t_c, t_r := s.DiffSet(t...)
	return t_l.ToSlice(), t_c.ToSlice(), t_r.ToSlice()
}
