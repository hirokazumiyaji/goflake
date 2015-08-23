# goflake
Implement golang of [Twitter's Snowflake](https://github.com/twitter/snowflake)

Usage
---

```
package main

import (
	"fmt"
	"time"

	"github.com/hirokazumiyaji/goflake"
)

func main() {
	startTime := time.Date(2000, 1, 1, 0, 0, 0, time.UTC)

	idWorker, err := goflake.NewIdWorker(1, 1, startTime)
	if err != nil {
		fmt.Println("failed new worker.", err)
		return
	}

	id, err := idWorker.NextId()
	if err != nil {
		fmt.Println("failed generate next id.", err)
	}
}
```

Test
---

```
$ go test
```

Benchmark
---

```
$ go test -bench .
```
