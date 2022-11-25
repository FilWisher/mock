# Mock

Generate simple-to-use mocks from interface types, following the pattern used
by [WTF Dial](https://github.com/benbjohnson/wtf).

Pass in an interface definition on standard in and it will output the interface
definition plus a simple mock implementation.

```
mock < interface.raw
```

I use this as an editor plugin:

```
" Generate mocks for a highlighted interface
vmap <leader>mm :!mock<CR>
" Select the interface the cursor is within and generate the mocks
nmap <leader>mm va{V:!mock<CR>
```

The mocks it generates look like this:

```
type Example interface {
    Foo(bar string) error
    Bar(context.Context, int)
}

type MockExample struct {
	FooFn func(bar string) error
	BarFn func(a context.Context, b int) 
}

func (o MockExample) Foo(bar string) error {
	return o.FooFn(bar)
}

func (o MockExample) Bar(a context.Context, b int)  {
	return o.BarFn(a, b)
}
```
