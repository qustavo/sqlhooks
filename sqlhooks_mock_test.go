package sqlhooks

type HooksMock struct {
	beforeQuery func(c *Context) error
	afterQuery  func(c *Context) error

	beforeExec func(c *Context) error
	afterExec  func(c *Context) error

	beforeBegin func(c *Context) error
	afterBegin  func(c *Context) error

	beforeCommit func(c *Context) error
	afterCommit  func(c *Context) error

	beforeRollback func(c *Context) error
	afterRollback  func(c *Context) error

	beforePrepare func(c *Context) error
	afterPrepare  func(c *Context) error

	beforeStmtQuery func(c *Context) error
	afterStmtQuery  func(c *Context) error

	beforeStmtExec func(c *Context) error
	afterStmtExec  func(*Context) error
}

func (h HooksMock) BeforeQuery(c *Context) error {
	if h.beforeQuery != nil {
		return h.beforeQuery(c)
	}
	return nil
}

func (h HooksMock) AfterQuery(c *Context) error {
	if h.afterQuery != nil {
		return h.afterQuery(c)
	}
	return nil
}

func (h HooksMock) BeforeExec(c *Context) error {
	if h.beforeExec != nil {
		return h.beforeExec(c)
	}
	return nil
}

func (h HooksMock) AfterExec(c *Context) error {
	if h.afterExec != nil {
		return h.afterExec(c)
	}
	return nil
}

func (h HooksMock) BeforeBegin(c *Context) error {
	if h.beforeBegin != nil {
		return h.beforeBegin(c)
	}
	return nil
}

func (h HooksMock) AfterBegin(c *Context) error {
	if h.afterBegin != nil {
		return h.afterBegin(c)
	}
	return nil
}

func (h HooksMock) BeforeCommit(c *Context) error {
	if h.beforeCommit != nil {
		return h.beforeCommit(c)
	}
	return nil
}

func (h HooksMock) AfterCommit(c *Context) error {
	if h.afterCommit != nil {
		return h.afterCommit(c)
	}
	return nil
}

func (h HooksMock) BeforeRollback(c *Context) error {
	if h.beforeRollback != nil {
		return h.beforeRollback(c)
	}
	return nil
}

func (h HooksMock) AfterRollback(c *Context) error {
	if h.afterRollback != nil {
		return h.afterRollback(c)
	}
	return nil
}

func (h HooksMock) BeforePrepare(c *Context) error {
	if h.beforePrepare != nil {
		return h.beforePrepare(c)
	}
	return nil
}

func (h HooksMock) AfterPrepare(c *Context) error {
	if h.afterPrepare != nil {
		return h.afterPrepare(c)
	}
	return nil
}

func (h HooksMock) BeforeStmtQuery(c *Context) error {
	if h.beforeStmtQuery != nil {
		return h.beforeStmtQuery(c)
	}
	return nil
}

func (h HooksMock) AfterStmtQuery(c *Context) error {
	if h.afterStmtQuery != nil {
		return h.afterStmtQuery(c)
	}
	return nil
}

func (h HooksMock) BeforeStmtExec(c *Context) error {
	if h.beforeStmtExec != nil {
		return h.beforeStmtExec(c)
	}
	return nil
}

func (h HooksMock) AfterStmtExec(c *Context) error {
	if h.afterStmtExec != nil {
		return h.afterStmtExec(c)
	}
	return nil
}

func NewHooksMock(before, after func(*Context) error) *HooksMock {
	return &HooksMock{
		beforeQuery:     before,
		beforeExec:      before,
		beforeBegin:     before,
		beforeCommit:    before,
		beforeRollback:  before,
		beforePrepare:   before,
		beforeStmtQuery: before,
		beforeStmtExec:  before,
		afterQuery:      after,
		afterExec:       after,
		afterBegin:      after,
		afterCommit:     after,
		afterRollback:   after,
		afterPrepare:    after,
		afterStmtQuery:  after,
		afterStmtExec:   after,
	}
}
