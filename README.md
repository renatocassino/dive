# Dive

Dive is a client debugger that use Devel lib to run debugger in terminal.

### But, why you need use Dive?

In Devel, when you start a client you must create the breakpoints before.

Example:

```bash
$ dlv debug main.go
Type 'help' for list of commands.
(dlv) break name0 ./actions/controller.go:43
(dlv) break name0 ./actions/controller.go:49
(dlv) break name0 ./actions/model.go:12
(dlv) continue
```

If you find a bug and must restart the application, you must set everything again (boooring).

## Dive in action

Here, an example using Dive

In your go file, only add the comment with `@brk` in your file:

```go
 func ApiHandler(c echo.Context) error {
        // @brk
        if len(words) == 0 {
                file, err := ioutil.ReadFile("pt-br.txt")
                if err != nil {
.....
```

Now, run:

```bash
$ dive main.go
```

The breakpoint will be setted and the debugger will start :D

Bugs? `<renatocassino@gmail.com>`
