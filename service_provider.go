package wlgo_svc_lifecycle_mgr

// It just a interface. You can build your own service provider in your service package.
type ServiceProvider interface {
	// Let you know which service. Just a tag for management.
	GetName() string

	// Open Log file, Connect to DB, Start listen socket, Prepare variables(Memory).... Do everything you want
	Begin() error

	// Close Log file, Disconnect to DB, Stop listen socket, Release variables(Memory)..... Do everything you want
	End() error
}

type serviceProvider struct {
	ServiceProvider
	name  string
	begin func() error
	end   func() error
}

// It is just a convenience method for you to build service provider in quick way.
//
// If you pass nil to the `begin` and `end` parameter.
//
// It will asign default empty func for you.
//
// default empty func: func() error { return nil }
//
// Example:
//
//	my_service := wlgo_svc_lifecycle_mgr.NewServiceProvider(
//	    "my_service",
//	    func() error {
//	        fmt.Println("my_service is started")
//			return nil
//	    },
//	    func() error {
//	        fmt.Println("my_service is ended")
//			return nil
//	    },
//	)
func NewServiceProvider(name string, begin func() error, end func() error) ServiceProvider {
	if begin == nil {
		begin = func() error { return nil }
	}
	if end == nil {
		end = func() error { return nil }
	}
	return &serviceProvider{
		name:  name,
		begin: begin,
		end:   end,
	}
}

func (sp *serviceProvider) GetName() string {
	return sp.name
}

func (sp *serviceProvider) Begin() error {
	return sp.begin()
}

func (sp *serviceProvider) End() error {
	return sp.end()
}
