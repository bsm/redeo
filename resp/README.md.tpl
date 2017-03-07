# RESP

Low-level primitives for dealing with RESP (REdis Serialization Protocol), client and server-side.

## Server Examples

Reading requests:

```go
package main

import (
  "fmt"
  "strings"

  "github.com/bsm/redeo/resp"
)

func main() {{ "ExampleRequestReader" | code }}
```

Writing responses:

```go
package main

import (
  "bytes"
  "fmt"

  "github.com/bsm/redeo/resp"
)

func main() {{ "ExampleResponseWriter" | code }}
```

## Client Examples

Reading requests:

```go
package main

import (
  "fmt"
  "net"

  "github.com/bsm/redeo/resp"
)

func main() {{ "Example_client" | code }}
```
