package authcontext

import "maps"

import "slices"

type Scope string

type ScopeSet map[Scope]struct{}

func NewScopeSet(scopes ...Scope) ScopeSet {
	s := make(ScopeSet, len(scopes))
	for _, sc := range scopes {
		s[sc] = struct{}{}
	}
	return s
}

func (s ScopeSet) Has(scope Scope) bool {
	if s == nil {
		return false
	}
	_, ok := s[scope]
	return ok
}

func (s ScopeSet) HasAll(scopes ...Scope) bool {
	if s == nil {
		return len(scopes) == 0
	}
	for _, sc := range scopes {
		if !s.Has(sc) {
			return false
		}
	}
	return true
}

func (s ScopeSet) HasAny(scopes ...Scope) bool {
	if s == nil {
		return false
	}
	for _, sc := range scopes {
		if s.Has(sc) {
			return true
		}
	}
	return false
}

func (s ScopeSet) Clone() ScopeSet {
	if s == nil {
		return nil
	}
	cp := make(ScopeSet, len(s))
	maps.Copy(cp, s)
	return cp
}

func (s ScopeSet) Sorted() []Scope {
	if s == nil {
		return nil
	}
	result := make([]Scope, 0, len(s))
	for sc := range s {
		result = append(result, sc)
	}
	slices.Sort(result)
	return result
}
