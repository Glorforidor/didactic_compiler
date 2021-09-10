package compiler

import "fmt"

type register struct {
	name  string
	inuse bool
}

type registerTable []*register

func (rt registerTable) alloc() (int, error) {
	for i, v := range rt {
		if !v.inuse {
			v.inuse = true
			return i, nil
		}
	}

	return 0, fmt.Errorf("no more register available")
}

func (rt registerTable) dealloc(reg int) {
	rt[reg].inuse = false
}

func (rt registerTable) name(reg int) string {
	return rt[reg].name
}
