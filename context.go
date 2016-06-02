package sqlhooks

type Context struct {
	Error error
	Query string
	Args  []interface{}

	values map[string]interface{}
}

func NewContext() *Context {
	return &Context{}
}

func (ctx *Context) Get(key string) interface{} {
	if ctx.values == nil {
		ctx.values = make(map[string]interface{})
	}

	if v, ok := ctx.values[key]; ok {
		return v
	}

	return nil
}

func (ctx *Context) Set(key string, value interface{}) {
	if ctx.values == nil {
		ctx.values = make(map[string]interface{})
	}

	ctx.values[key] = value
}
