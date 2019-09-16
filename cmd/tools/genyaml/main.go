package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"text/template"
)

type data struct {
	Namespace, Resource string
	WithEndingDashes bool
}

func main() {
	d := &data{}

	pflag.StringVarP(&d.Namespace, "namespace", "n", "", "Namespace of the BPF resource (optional)")
	pflag.StringVarP(&d.Resource, "resource", "r", "", "Name of the BPF resource")
	pflag.BoolVar(&d.WithEndingDashes, "ending-dashes", false, "Whether to append three dashes at the end or not")
	pflag.Parse()

	if(d.Resource == "") {
    fmt.Fprintf(os.Stderr, "missing resource name\n")
    os.Exit(1)
  }

	const yaml = `{{if .Namespace}}---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}
{{end}}---
apiVersion: bpf.sh/v1alpha1
kind: BPF
metadata:
  name: {{.Resource}}-bpf{{if .Namespace}}
  namespace: {{.Namespace}}{{end}}
spec:
  program:
    valueFrom:
      configMapKeyRef:
        name: {{.Resource}}-config
		key: {{.Resource}}.o{{if .WithEndingDashes}}
---
{{end}}`

	t := template.Must(template.New("yaml").Parse(yaml))
	t.Execute(os.Stdout, d)
}
