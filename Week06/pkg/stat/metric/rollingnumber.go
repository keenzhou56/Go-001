// Package rollingnumber 简化实现 Hystrix 的滚动窗口算法，
// 结合了 Hystrix 和 https://github.com/afex/hystrix-go 的实现。
// 使用桶数组，Increment 时删除旧桶
package rollingnumber

import (
	"sync"
	"sync/atomic"
	"time"
)

// RollingNumber 用于追踪计数
type RollingNumber struct {
	buckets [10]*bucket
	lock    *sync.RWMutex
}

// NewRollingNumber 构造 RollingNumber 实例
func NewRollingNumber() *RollingNumber {
	buckets := [10]*bucket{}
	startTime := time.Now().Unix()
	for i := 0; i < 10; i++ {
		buckets[i] = newBucket(startTime + int64(i))
	}
	r := &RollingNumber{
		buckets: buckets,
		lock:    &sync.RWMutex{},
	}
	return r
}

// bucket 为计数桶
type bucket struct {
	windowStart int64
	value       int64
}

func newBucket(startTime int64) *bucket {
	return &bucket{
		windowStart: startTime,
		value:       int64(0),
	}
}

func (r *RollingNumber) getCurrentBucket(time int64) *bucket {
	index := time % 10
	return r.buckets[index]
}

func (r *RollingNumber) resetBuckets(now int64) {
	for i, b := range r.buckets {
		b.windowStart = now + int64(i)
		b.value = int64(0)
	}
}

// Increment 增加一个桶的计数时
// 此时有三种情况，以1号桶为例，当前时间秒数为21，
// 如果桶的 windowStart 为当前时间，将桶的 value 加1。
// 如果桶的 windowStart 为当前分钟的11秒，重置 windowStart 并置 value 为 1。
// 如果桶的 windowStart 为早于当前分钟的11秒，即超过一个计数周期（10秒）未更新
// 过计数器，则先重置计数器
func (r *RollingNumber) Increment(now int64) {
	bucket := r.getCurrentBucket(now)
	if bucket.windowStart == now {
		atomic.AddInt64(&bucket.value, int64(1))
		return
	}
	if bucket.windowStart == now-10 {
		r.lock.Lock()
		bucket.windowStart = now
		bucket.value = 1
		r.lock.Unlock()
		return
	}
	// 如果间隔10秒以上，即前10秒（一个计数周期）之内，没有新增请求，
	// 桶中的数据就没用了，先清空
	r.lock.Lock()
	r.resetBuckets(now)
	bucket.value = 1
	r.lock.Unlock()
}

// Sum 计算最新 10 个桶内计数器的和
func (r *RollingNumber) Sum(now time.Time) int64 {
	sum := int64(0)

	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, b := range r.buckets {
		if b.windowStart >= now.Unix()-10 {
			sum += b.value
		}
	}

	return sum
}

// Avg 计算最新 10 个桶内计数器的平均值
func (r *RollingNumber) Avg(now time.Time) int64 {
	return r.Sum(now) / 10
}