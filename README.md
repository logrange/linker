# Linker

[![GoDoc](https://godoc.org/github.com/logrange/linker?status.png)](https://godoc.org/github.com/logrange/linker)

Linker is Dependency Injection and Inversion of Control package. 

Linker's highlights:
 - Components registry
 - Dependency injection
 - Components initialization prioritization
 - Initialization and shutdown control
 
```golang

import (
     "github.com/lograng/linker"
)

type DatabaseAccessService interface {
    RunQuery(query string) DbResult
}
// MySQLAccessService implements DatabaseAccessService
type MySQLAccessService struct {
    Conns int `inject:"mySqlConns, optional:32"`
}
type BigDataService struct {
    dba DatabaseAccessService `inject:"dba"`
}
...
func main() {
    // 1st step is to create the injector
    inj := linker.New()
    // 2nd step is to register components
    inj.Register(
		linker.Component{Name: "dba", Value: &MySQLAccessService{}},
		Component{Name: "", Value: &BigDataService{}},
		Component{Name: "mySqlConns", Value: int(msconns)},
		...
	)
	// 3rd step is to inject dependecies and initialize the registered components
	inj.Init(ctx)
	...
	// 4th de-initialize all compoments properly
	inj.Shutdown()
    )

```
### Annotate fields using fields tags
### Create the injector
### Register components using names or anonymously
### Initialize components
### Shutting down registered components properly
