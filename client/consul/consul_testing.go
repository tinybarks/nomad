package consul

import (
	"fmt"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/command/agent/consul"
	"github.com/hashicorp/nomad/helper/testlog"
	"github.com/mitchellh/go-testing-interface"
)

// MockConsulOp represents the register/deregister operations.
type MockConsulOp struct {
	Op      string // add, remove, or update
	AllocID string
	Task    string
}

func NewMockConsulOp(op, allocID, task string) MockConsulOp {
	if op != "add" && op != "remove" && op != "update" && op != "alloc_registrations" {
		panic(fmt.Errorf("invalid consul op: %s", op))
	}
	return MockConsulOp{
		Op:      op,
		AllocID: allocID,
		Task:    task,
	}
}

// MockConsulServiceClient implements the ConsulServiceAPI interface to record
// and log task registration/deregistration.
type MockConsulServiceClient struct {
	ops []MockConsulOp
	mu  sync.Mutex

	logger hclog.Logger

	// AllocRegistrationsFn allows injecting return values for the
	// AllocRegistrations function.
	AllocRegistrationsFn func(allocID string) (*consul.AllocRegistration, error)
}

func NewMockConsulServiceClient(t testing.T) *MockConsulServiceClient {
	m := MockConsulServiceClient{
		ops:    make([]MockConsulOp, 0, 20),
		logger: testlog.HCLogger(t).Named("mock_consul"),
	}
	return &m
}

func (m *MockConsulServiceClient) UpdateTask(old, new *consul.TaskServices) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("UpdateTask", "alloc_id", new.AllocID[:6], "task_name", new.Name)
	m.ops = append(m.ops, NewMockConsulOp("update", new.AllocID, new.Name))
	return nil
}

func (m *MockConsulServiceClient) RegisterTask(task *consul.TaskServices) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("RegisterTask", "alloc_id", task.AllocID, "task_name", task.Name)
	m.ops = append(m.ops, NewMockConsulOp("add", task.AllocID, task.Name))
	return nil
}

func (m *MockConsulServiceClient) RemoveTask(task *consul.TaskServices) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("RemoveTask", "alloc_id", task.AllocID, "task_name", task.Name)
	m.ops = append(m.ops, NewMockConsulOp("remove", task.AllocID, task.Name))
}

func (m *MockConsulServiceClient) AllocRegistrations(allocID string) (*consul.AllocRegistration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("AllocRegistrations", "alloc_id", allocID)
	m.ops = append(m.ops, NewMockConsulOp("alloc_registrations", allocID, ""))

	if m.AllocRegistrationsFn != nil {
		return m.AllocRegistrationsFn(allocID)
	}

	return nil, nil
}

func (m *MockConsulServiceClient) GetOps() []MockConsulOp {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ops
}
