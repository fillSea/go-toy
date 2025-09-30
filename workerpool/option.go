package workerpool

// Option 用于配置workerpool的选项
type Option func(*Pool)

// WithBlock 配置workerpool是否阻塞新的Schedule调用
func WithBlock(block bool) Option {
	return func(p *Pool) {
		p.block = block
	}
}

// WithPreAllocWorkers 配置workerpool是否在创建pool的时候就预创建workers
func WithPreAllocWorkers(preAlloc bool) Option {
	return func(p *Pool) {
		p.preAlloc = preAlloc
	}
}