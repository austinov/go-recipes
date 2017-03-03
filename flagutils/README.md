# flagutils

flagutils helps to distinguish if a flag was passed or not.

Usage:
```
    import (
      util "flagutils"
      "flag"
      "fmt"
    )
    var debug util.BoolFlag
    flag.Var(&debug, "debug", "debug mode")
    flag.Parse()

    if debug.Exist {
      fmt.Printf("debug mode not set")
    } else {
      fmt.Printf("debug mode set to %v", debug.Value)
    }
```