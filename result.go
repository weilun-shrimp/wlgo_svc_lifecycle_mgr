package wlgo_svc_lifecycle_mgr

type result struct {
	err                error
	errServiceProvider ServiceProvider
}

// You can know what kind of error be returned.
// Also can know which `ServiceProvider` return the error.
// You will get it from `Manager.Start()` and `Manager.Rollback()`.
//
// Example:
//
//	 var r wlgo_svc_lifecycle_mgr.Result = Manager.Start() // or Rollback()
//	 if r.GetError() == nil {
//		return
//	 }
//	 // You can check which service return the error. And do everything you want. eg. log the error.
//	 log.Printf(
//		"Error happend in process. Msg: %s | Service name: %s",
//		r.GetError().Error(),
//		r.GetErrServiceProvider().GetName(),
//	 )
//	 switch r.GetErrServiceProvider().GetName() {
//		case "s1":
//			// Do something
//		case "s2":
//			// Do something
//		default:
//			// Do something
//	 }
type Result interface {
	GetError() error
	GetErrServiceProvider() ServiceProvider
}

func (r result) GetError() error {
	return r.err
}

// If GetError not nil. It will not return nil too.
//
// If GetError is nil. It will return nil too.
func (r result) GetErrServiceProvider() ServiceProvider {
	return r.errServiceProvider
}
