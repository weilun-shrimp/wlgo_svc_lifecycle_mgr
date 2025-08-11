package wlgo_svc_lifecycle_mgr

type Manager struct {
	services        []ServiceProvider
	startedServices []ServiceProvider
}

func NewManager() *Manager {
	return &Manager{
		services:        make([]ServiceProvider, 0),
		startedServices: make([]ServiceProvider, 0),
	}
}

func (m *Manager) AddService(services ...ServiceProvider) {
	m.services = append(m.services, services...)
}

func (m *Manager) Start() Result {
	m.startedServices = make([]ServiceProvider, 0, len(m.services))
	var err error
	var errServiceProvider ServiceProvider

	for _, s := range m.services {
		if err = s.Begin(); err != nil {
			errServiceProvider = s
			break
		}
		m.startedServices = append(m.startedServices, s)
	}

	return &result{
		err:                err,
		errServiceProvider: errServiceProvider,
	}
}

// Must be called after `Start` func.
//
// Otherwise it will no nothing. And you will get a nothing result (nil error) for sure.
func (m *Manager) Rollback() Result {
	var err error
	var errServiceProvider ServiceProvider

	for i := len(m.startedServices) - 1; i >= 0; i-- {
		if err = m.startedServices[i].End(); err != nil {
			errServiceProvider = m.startedServices[i]
			break
		}
	}

	return &result{
		err:                err,
		errServiceProvider: errServiceProvider,
	}
}
