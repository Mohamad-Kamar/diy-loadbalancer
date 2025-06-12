package admin

import (
	"round-robin-api/internal/metrics"
	"testing"
)

type dummyLB struct{}

func (d *dummyLB) AddBackend(url string) {}
func (d *dummyLB) RemoveBackend(url string) {}
func (d *dummyLB) GetBackends() []string { return []string{"http://localhost:8081"} }

func TestNewAdminServer(t *testing.T) {
	m := metrics.NewMetrics()
	lb := &dummyLB{}
	admin := NewAdminServer(m, lb)
	if admin == nil {
		t.Fatal("AdminServer should not be nil")
	}
}
