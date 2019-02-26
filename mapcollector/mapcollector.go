package mapcollector

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/leodido/bpf-operator/loader"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func New(path string, logger *zap.Logger) *MapCollector {
	return &MapCollector{
		l:    logger,
		path: path,
	}
}

type MapCollector struct {
	l            *zap.Logger
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
		c.l.Info("New map found", zap.String("name", mapname), zap.String("type", "hash"))

		desc := prometheus.NewDesc(
			prometheus.BuildFQName("test", "", mapname),
			mapname,
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

			key := fmt.Sprintf("%05d", k) // todos > read number of digits of map keys (from map definition) and pad them accordingly
			c.l.Debug("New map value", zap.String("map", mapname), zap.String("key", key), zap.Uint64("val", v), zap.String("node", nodeName))
			ch <- prometheus.MustNewConstMetric(
				c.descriptions[mapname],
				prometheus.CounterValue,
				float64(v),
				key,
				nodeName,
			)
		}
	}
}
