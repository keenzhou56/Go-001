// Package rollingnumber 简化实现 Hystrix 的滚动窗口算法，
// 借鉴了 https://github.com/afex/hystrix-go 的实现。
// 没有使用循环队列，但是巧妙地以秒级时间戳为键，在超过10个桶以后删除旧桶，
// hystrix-go 删除旧桶操作在锁里，临界区较大，影响性能。
// 由于计算均值时只用最新的 10 个桶（最近1秒的数据），未及时删除旧桶
// 不会对计算造成错误的影响，所以我将删除操作移出了临界区
package metric

import (
	"sync"
	"time"
)

// RollingNumber 用于追踪计数
// 字典的键是时间戳
type RollingNumber struct {
	Buckets map[int64]*bucket
	Mu      *sync.RWMutex
}

// bucket 为计数桶
type bucket struct {
	Value int64
}

// NewRollingNumber 构造 RollingNumber 实例
func NewRollingNumber() *RollingNumber {
	rn := &RollingNumber{
		Buckets: make(map[int64]*bucket),
		Mu:      &sync.RWMutex{},
	}
	return rn
}

func (rn *RollingNumber) getCurrentBucket() *bucket {
	now := time.Now().Unix()
	var b *bucket
	var ok bool
	if b, ok = rn.Buckets[now]; !ok {
		b = &bucket{}
		rn.Buckets[now] = b
	}
	return b
}

func (rn *RollingNumber) removeOldBuckets() {
	expired := time.Now().Unix() - 10
	for timestamp := range rn.Buckets {
		if timestamp <= expired {
			delete(rn.Buckets, timestamp)
		}
	}
}

// Increment 累加最新桶的计数器
func (rn *RollingNumber) Increment() {
	rn.Mu.Lock()
	b := rn.getCurrentBucket()
	b.Value++
	rn.removeOldBuckets()

	rn.Mu.Unlock()
}

// UpdateMax 将最新桶的计数器置为某个最大值
func (rn *RollingNumber) UpdateMax(n int64) {
	rn.Mu.Lock()
	b := rn.getCurrentBucket()
	if n > b.Value {
		b.Value = n
	}
	rn.Mu.Unlock()
	rn.removeOldBuckets()
}

// Sum 计算最新 10 个桶内计数器的和
func (rn *RollingNumber) Sum(now time.Time) int64 {
	sum := int64(0)

	rn.Mu.RLock()
	defer rn.Mu.RUnlock()

	for timestamp, bucket := range rn.Buckets {
		if timestamp >= now.Unix()-10 {
			sum += bucket.Value
		}
	}
	return sum
}

// Max 获取最新 10 个桶内计数器的最大值
func (rn *RollingNumber) Max(now time.Time) int64 {
	var max int64

	rn.Mu.RLock()
	defer rn.Mu.RUnlock()

	for timestamp, bucket := range rn.Buckets {
		if timestamp >= now.Unix()-10 {
			if bucket.Value > max {
				max = bucket.Value
			}
		}
	}
	return max
}

// Avg 计算最新 10 个桶内计数器的平均值
func (rn *RollingNumber) Avg(now time.Time) int64 {
	return rn.Sum(now) / 10
}