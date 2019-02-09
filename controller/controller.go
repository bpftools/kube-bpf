package controller

import (
	"sync"
)

type Controller interface {
	Run(threadiness int, stopCh <-chan struct{}, wg *sync.WaitGroup)
}
