# WL golang service lifecycle manager
Manage your services lifecycle in very easy way.

The manager will trigger your services `Begin` func in the sequence you added into it if you call the `Start` func.

And trigger the `End` func in the reversed sequence you added into it if you call the `Rollback` func.

Please follow the guide. Set your service as the `ServiceProvider` interface.

-   [Installation](#installation)
-   [Visualize Concept](#visualize-concept)
-   [Service Provider](#service-provider)
-   [Quick Example](#quick-example)
-   [Real World Example](#real-world-example)

## Installation
```bash
$ go get -u github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr
```

## Visualize Concept
```
+-----------+
|  Manager  |
+-----------+
      |
      v
Add services in order: A → B → C → D
      |
      v
    Start()
      |
      |
      |          +-----------+     +-----------+     +-----------+     +-----------+
      ---------->| Service A | --> | Service B | --> | Service C | --> | Service D | --> no error 
                 +-----------+     +-----------+     +-----------+     +-----------+         |
                       |                 |                 |                 |               |
                       v                 v                 v                 v               |
                    Begin()           Begin()            Begin()           Begin()           |
                                                           |                                 |
                                                           v                                 v
    ------------------ No trigger <-- Rollback <-------- error                            Rollback --> No trigger
    |                                    |                                                   |              |
    |                                    v                                                   v              |
    |                                 Trigger                                             Trigger           |
    |                                    |                                                   |              |
    |                                    v                                                   |              |
    |            +-----------+     +-----------+     +-----------+     +-----------+         |              |
    |     -------| Service A | <-- | Service B | <-- | Service C | <-- | Service D | <--------              |
    |     |      +-----------+     +-----------+     +-----------+     +-----------+                        |
    |     v            ^                 ^                 ^                 ^                              |
    |  no error        |                 |                 |                 |                              |
    |     |          End()             End()             End()             End()                            |
    |     |                              |                                                                  |
    |     |                              v                                                                  |
    |     |  |------------------------ error                                                                |
    v     v  v                                                                                              |
+---------------+                                                                                           |
|  Process End  | <------------------------------------------------------------------------------------------
+---------------+ 

Note: 
    If error be returned in service C Begin():
    - Service D Begin func will not be triggered. 
    - Service C, D End func will not be triggered.
```

## Service Provider
It will provide your service to the lifecycle manager.
```golang
type ServiceProvider interface {
	// Let you know which service. Just a tag for management.
	GetName() string
	// Open Log file, Connect to DB, Start listen socket, Prepare variables(Memory).... Do everything you want
	Begin() error
	// Close Log file, Disconnect to DB, Stop listen socket, Release variables(Memory)..... Do everything you want
	End() error
}
```

You can build your own service provider in your service package.
```golang
package my_service

import (
    "github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
)

type MyCustomServiceProvider struct {
	wlgo_svc_lifecycle_mgr.ServiceProvider
}
func (sp *MyCustomServiceProvider) GetName() string {
	return "my custom name"
}
func (sp *MyCustomServiceProvider) Begin() error {
    // Open Log file, Connect to DB, Start listen socket, Prepare variables(Memory).... Do everything you want
	return nil
}
func (sp *MyCustomServiceProvider) End() error {
    // Close Log file, Disconnect to DB, Stop listen socket, Release variables(Memory)..... Do everything you want
	return nil
}
```

Or you can new a service provider by a conevenient way and use it directly.
```golang
package my_service

import (
    "github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
)

func GetServiceProvider() wlgo_svc_lifecycle_mgr.ServiceProvider {
    return wlgo_svc_lifecycle_mgr.NewServiceProvider(
        "my custom name",
        func() error {
            // Open Log file, Connect to DB, Start listen socket, Prepare variables(Memory).... Do everything you want
            return nil
        },
        func() error {
            // Close Log file, Disconnect to DB, Stop listen socket, Release variables(Memory)..... Do everything you want
            return nil
        },
    )
}
```


## Quick Example
```golang
package main

import (
    "github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
    "<your_service_package_path>/service_package_2"
)

func main() {
    // Prepare manager
    var manager wlgo_svc_lifecycle_mgr.Manager := wlgo_svc_lifecycle_mgr.NewManager()
    // Prepare rollback handler in defer (optional)
    defer func() {
        // Start to run every services End func
        var rollback_result wlgo_svc_lifecycle_mgr.Result := manager.Rollback()
        if rollback_result == nil {
            return
        }
        // You can log the error and check which service return the error. And do everything you want.
        log.Printf(
            "Error happend in process. Msg: %s | Service name: %s",
            r.GetError().Error(),
            r.GetErrServiceProvider().GetName(),
        )
        ...... // Do everything you want
    }()

    // Prepare first service - direct way. For more readable.
    service1 := wlgo_svc_lifecycle_mgr.NewServiceProvider(
        "service1",
        func() error {
            fmt.Println("service1 is started")
        },
        func() error {
            fmt.Println("service1 is ended")
        },
    )
    // Prepare second service - You can put service provider in you package. For more clean code.
    service2 := service_package_2.GetServiceProvider()

    // Add services into manager
    manager.AddService(
        service1,
        service2,
    )
    // Start to run every services Begin func
    var start_result wlgo_svc_lifecycle_mgr.Result := manager.Start()
    // Check result to know has any error be returned
    if start_result.GetError() != nil {
        // Do everything you want......
    }
}
```


## Real World Example

It is from one of my online running project. 

I just made some anonymisation for my packages and add some tutorials.

```golang
package main

import (
    "github.com/weilun-shrimp/wlgo_svc_lifecycle_mgr"
    "internal/my_logger"
    "internal/gin_server"
    "internal/custom_websocket_server"
    "internal/custom_running_service1"
    "internal/custom_running_service2"
    "database/mongodb"
    "database/mysql"
    "database/redis_master"
    "dtabase/redis_readonly"
)

func main() {
    var manager wlgo_svc_lifecycle_mgr.Manager := wlgo_svc_lifecycle_mgr.NewManager()
    defer func() {
        var rollback_result wlgo_svc_lifecycle_mgr.Result := manager.Rollback()
        if rollback_result == nil {
            return
        }
        switch r.GetErrServiceProvider().GetName() {
            case "my_logger":
                fmt.Pringln("logger end fail")
            default:
                my_logger.Write("error", my_logger.Msg{
                    "on": "main/rollback_result",
                    "msg": rollback_result.GetError().Error(),
                    "service_name": rollback_result.GetErrServiceProvider().GetName(),
                })
        }
    }()

    manager.AddService(
        // Direct way to new ServiceProvider - For more readable.
        wlgo_svc_lifecycle_mgr.NewServiceProvider(
            "my_logger",
            func() error {
                err := my_logger.CheckAndOpenLogFile()
                return err
            },
            func() error {
                err := my_logger.CloseLogFile()
                return err
            },
        )
        wlgo_svc_lifecycle_mgr.NewServiceProvider(
            "mongodb",
            func() error {
                err := mongodb.InitConnection()
                return err
            },
            func() error {
                err := mongodb.ReleaseConnection()
                return err
            },
        )
        wlgo_svc_lifecycle_mgr.NewServiceProvider(
            "custom_websocket_server",
            func() error {
                err := custom_websocket_server.StartListen()
                return err
            },
            func() error {
                err := custom_websocket_server.CloseAllConnection()
                return err
            },
        )

        // Put service provider in your package. For more clean code.
        mysql.GetServiceProvider(), 
        redis_master.GetServiceProvider(), 
        redis_readonly.GetServiceProvider(), 
        gin_server.GetServiceProvider(), 
        custom_running_service1.GetServiceProvider(), 
        custom_running_service2.GetServiceProvider(), 
    )

    var start_result wlgo_svc_lifecycle_mgr.Result := manager.Start()
    if start_result.GetError() != nil {
        my_logger.Write("error", my_logger.Msg{
            "on": "main/start_result",
            "msg": start_result.GetError().Error(),
            "service_name": start_result.GetErrServiceProvider().GetName(),
        })
    }
}
```