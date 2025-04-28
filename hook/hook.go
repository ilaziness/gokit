package hook

type Closure func()
type Starter func()

type exitType struct {
	hooks []Closure
}

type startType struct {
	hooks []Starter
}

// Exit 应用退出钩子
var Exit *exitType

// Start 应用启动钩子
var Start *startType

func init() {
	Exit = &exitType{}
	Start = &startType{}
}

func (e *exitType) Register(hook Closure) {
	e.hooks = append(e.hooks, hook)
}

func (e *exitType) Trigger() {
	for _, hook := range e.hooks {
		hook()
	}
}

func (s *startType) Register(hook Starter) {
	s.hooks = append(s.hooks, hook)
}

func (s *startType) Trigger() {
	for _, hook := range s.hooks {
		hook()
	}
}
