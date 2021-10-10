package compiler

import (
	"fmt"
	"sort"
)

type registerTable struct {
	general, floating map[string]bool
}

func riscvTable() *registerTable {
	return &registerTable{
		general: map[string]bool{
			"t0": false,
			"t1": false,
			"t2": false,
			"t3": false,
			"t4": false,
			"t5": false,
			"t6": false,
		},
		floating: map[string]bool{
			"ft0":  false,
			"ft1":  false,
			"ft2":  false,
			"ft3":  false,
			"ft4":  false,
			"ft5":  false,
			"ft6":  false,
			"ft7":  false,
			"ft8":  false,
			"ft9":  false,
			"ft10": false,
			"ft11": false,
		},
	}
}

func (rt *registerTable) allocGeneral() (string, error) {
	var generals []string
	for k := range rt.general {
		generals = append(generals, k)
	}
	sort.Strings(generals)

	for _, v := range generals {
		if !rt.general[v] {
			rt.general[v] = true
			return v, nil
		}
	}

	return "", fmt.Errorf("no more general register available: %v", rt.general)
}

func (rt *registerTable) allocFloating() (string, error) {
	var floats []string
	for k := range rt.floating {
		floats = append(floats, k)
	}
	sort.Strings(floats)

	for _, v := range floats {
		if !rt.floating[v] {
			rt.floating[v] = true
			return v, nil
		}
	}

	return "", fmt.Errorf("no more floating register available: %v", rt.floating)
}

func (rt *registerTable) dealloc(reg string) {
	if _, ok := rt.general[reg]; ok {
		rt.general[reg] = false
	} else if _, ok := rt.floating[reg]; ok {
		rt.floating[reg] = false
	}
}
