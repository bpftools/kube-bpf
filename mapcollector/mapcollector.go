package mapcollector

import (
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/leodido/bpf-operator/loader"
	"github.com/prometheus/client_golang/prometheus"
)

func New(path string) *MapCollector {
	return &MapCollector{
		path: path,
	}
}

type MapCollector struct {
	path         string
	str          *loader.Loader
	descriptions map[string]*prometheus.Desc
}

func (c *MapCollector) Setup() error {
	str, err := loader.NewLoader(c.path)
	if err != nil {
		return err
	}
	c.str = str
	c.descriptions = make(map[string]*prometheus.Desc, 0)
	return nil
}

func (c *MapCollector) Describe(ch chan<- *prometheus.Desc) {
	labels := []string{
		"key",
		"node",
	}
	// Handling only type 1 for now // BPF_MAP_TYPE_HASH
	for _, mapname := range c.str.MapsMeta()[1] {
		log.Printf("Map: %q.\n", mapname) // todos > log with zap (injecting it?)
		desc := prometheus.NewDesc(
			prometheus.BuildFQName("test", "", mapname),
			fmt.Sprintf("Data coming from %s BPF map", mapname),
			labels,
			nil,
		)
		c.descriptions[mapname] = desc
		ch <- desc
	}
}

func (c *MapCollector) Collect(ch chan<- prometheus.Metric) {
	nodeName := os.Getenv("NODENAME")
	if nodeName == "" {
		nodeName = "unknown"
	}
	module := c.str.Module()
	// Handling only type 1 for now // BPF_MAP_TYPE_HASH
	for _, mapname := range c.str.MapsMeta()[1] {
		m := module.Map(mapname)
		// key, next key, and value
		var k, n, v uint64
		var err error
		still := true
		// continue until there are other values
		for still {
			still, err = module.LookupNextElement(m, unsafe.Pointer(&k), unsafe.Pointer(&n), unsafe.Pointer(&v))
			if err != nil {
				break
			}
			k = n

			ch <- prometheus.MustNewConstMetric(
				c.descriptions[mapname],
				prometheus.CounterValue,
				float64(v),
				fmt.Sprintf("%05d", k),
				nodeName,
			)
		}
	}
}
