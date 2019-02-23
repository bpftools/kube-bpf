package loader

import (
	"reflect"

	"github.com/davecgh/go-spew/spew"
	elf "github.com/iovisor/gobpf/elf"
)

// getMapAttribute implements an ugly hack to grab map definition attributes.
func getMapAttribute(m *elf.Map, attr string) uint64 {
	v := reflect.ValueOf(*m)
	y := v.FieldByName("m")
	t := reflect.Indirect(y).FieldByName("def").FieldByName(attr)
	return t.Uint()
}

type Loader struct {
	module   *elf.Module
	mapsMeta map[uint64][]string
}

func NewLoader(filepath string) (*Loader, error) {
	elf := elf.NewModule(filepath)
	err := elf.Load(nil)
	if err != nil {
		return nil, err
	}

	mapsMeta := make(map[uint64][]string, 0)
	for m := range elf.IterMaps() {
		t := getMapAttribute(m, "_type")
		if _, ok := mapsMeta[t]; !ok {
			mapsMeta[t] = make([]string, 0)
		}
		mapsMeta[t] = append(mapsMeta[t], m.Name)
	}

	for t := range elf.IterTracepointProgram() {
		err := elf.EnableTracepoint(t.Name)
		if err != nil {
			spew.Dump("exist: tracepoint error: %v\n", err)
			return nil, err
		}
	}

	return &Loader{
		module:   elf,
		mapsMeta: mapsMeta,
	}, nil
}

func (s *Loader) Module() *elf.Module {
	return s.module
}

func (s *Loader) MapsMeta() map[uint64][]string {
	return s.mapsMeta
}
