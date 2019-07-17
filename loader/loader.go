package loader

import (
	"github.com/davecgh/go-spew/spew"
	elf "github.com/iovisor/gobpf/elf"
	"github.com/skydive-project/skydive/common"
	"github.com/vishvananda/netlink"
	"reflect"
	"syscall"
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
	m := elf.NewModule(filepath)
	err := m.Load(nil)
	if err != nil {
		return nil, err
	}

	mapsMeta := make(map[uint64][]string, 0)
	for m := range m.IterMaps() {
		t := getMapAttribute(m, "_type")
		if _, ok := mapsMeta[t]; !ok {
			mapsMeta[t] = make([]string, 0)
		}
		mapsMeta[t] = append(mapsMeta[t], m.Name)
	}

	// program type: BPF_PROG_TYPE_TRACEPOINT
	for t := range m.IterTracepointProgram() {
		err := m.EnableTracepoint(t.Name)
		if err != nil {
			spew.Dump("exist: tracepoint error: %v\n", err) // todos > log with adequate logger
			return nil, err
		}
	}

	// program type: BPF_PROG_TYPE_SOCKET_FILTER
	for s := range m.IterSocketFilter() {
		// todos > it currently attaches to all the interfaces, maybe make them selectable?
		links, err := netlink.LinkList()
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			rs, err := common.NewRawSocketInNs("/proc/1/ns/net", link.Attrs().Name, syscall.ETH_P_ALL)
			if err != nil {
				return nil, err
			}
			fd := rs.GetFd()
			if err := elf.AttachSocketFilter(s, fd); err != nil {
				return nil, err
			}
		}
	}

	return &Loader{
		module:   m,
		mapsMeta: mapsMeta,
	}, nil
}

func (s *Loader) Module() *elf.Module {
	return s.module
}

func (s *Loader) MapsMeta() map[uint64][]string {
	return s.mapsMeta
}
