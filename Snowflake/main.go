package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/*
* 1                                               42           52             64
* +-----------------------------------------------+------------+---------------+
* | timestamp(ms)                                 | workerid   | sequence      |
* +-----------------------------------------------+------------+---------------+
* | 0000000000 0000000000 0000000000 0000000000 0 | 0000000000 | 0000000000 00 |
* +-----------------------------------------------+------------+---------------+
*
* 1. 41位时间截(毫秒级)，注意这是时间截的差值（当前时间截 - 开始时间截)。可以使用约70年: (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69
* 2. 10位数据机器位，可以部署在1024个节点
* 3. 12位序列，毫秒内的计数，同一机器，同一时间截并发4096个序号

worerIDBits：这里就是对应上图中的10bit-工作机器id，我这里进行拆分了。这是其中5bit``worerIDBits`
dataCenterIDBits：这里就是对应上图中的10bit-工作机器id，我这里进行拆分了。这是其中5bitdataCenterIDBits
sequenceBits：对应上图中的12bit的序列号
maxWorkerID：这里就是求最大，只不过我们采用了异或的方式，因为-1的二进制表示为1的补码，说通俗一点，这里其实就是2^5-1
maxDataCenterID：原理同上
maxSequence：原理同上
timeLeft：时间戳向左偏移量
dataLeft：原理同上，也是求偏移量的
workLeft：原理同上；
twepoch：41bit的时间戳，单位是毫秒，这里我选择的时间是2020-05-20 08:00:00 +0800 CST，这个ID一但生成就不要改了，要不会生成相同的ID。
 */

const (
	workerIDBits =  uint64(5)  // 10bit 工作机器ID中的 5bit workerID
	dataCenterIDBits = uint64(5) // 10 bit 工作机器ID中的 5bit dataCenterID
	sequenceBits = uint64(12)

	maxWorkerID = int64(-1) ^ (int64(-1) << workerIDBits) //节点ID的最大值 用于防止溢出
	maxDataCenterID = int64(-1) ^ (int64(-1) << dataCenterIDBits)
	maxSequence = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22)  // timeLeft = workerIDBits + sequenceBits // 时间戳向左偏移量
	dataLeft = uint8(17)  // dataLeft = dataCenterIDBits + sequenceBits
	workLeft = uint8(12)  // workLeft = sequenceBits // 节点IDx向左偏移量
	// 2020-05-20 08:00:00 +0800 CST
	twepoch = int64(1589923200000) // 常量时间戳(毫秒)
)

/*
mu sync.Mutex：添加互斥锁，确保并发安全性
LastStamp int64：记录上一次生成ID的时间戳
WorkerID int64：该工作节点的ID 对上图中的5bit workerID 一个意思
DataCenterID int64： 原理同上
Sequence int64：当前毫秒已经生成的id序列号(从0开始累加) 1毫秒内最多生成4096个ID
 */

type Worker struct {
	mu sync.Mutex
	LastStamp int64 // 记录上一次ID的时间戳
	WorkerID int64 // 该节点的ID
	DataCenterID int64 // 该节点的 数据中心ID
	Sequence int64 // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID
}

// NewWorker 分布式情况下,我们应通过外部配置文件或其他方式为每台机器分配独立的id
func NewWorker(workerID,dataCenterID int64) *Worker  {
	return &Worker{
		WorkerID: workerID,
		LastStamp: 0,
		Sequence: 0,
		DataCenterID: dataCenterID,
	}
}

// getMilliSeconds 获取当前的毫秒值
func (w *Worker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

// NextID 封装获取ID方法
func (w *Worker)NextID() (uint64,error) {
	// 添加互斥锁
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.nextID()
}

// nextID 具体获取ID方法
func (w *Worker)nextID() (uint64,error) {
	// 获取当前时间戳
	timeStamp := w.getMilliSeconds()
	// 确保当前时间戳值大于上一次生成ID的时间戳
	if timeStamp < w.LastStamp{
		return 0,errors.New("time is moving backwards,waiting until")
	}
	// 如果当前毫秒已经生成的id序列号溢出了，则需要等待下一毫秒
	if w.LastStamp == timeStamp{
		w.Sequence = (w.Sequence + 1) & maxSequence
		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	}else {
		w.Sequence = 0 // 当前时间与工作节点上一次生成ID的时间不一致 则需要重置工作节点生成ID的序号
	}

	w.LastStamp = timeStamp
	id := ((timeStamp - twepoch) << timeLeft) |
		(w.DataCenterID << dataLeft)  |
		(w.WorkerID << workLeft) |
		w.Sequence

	return uint64(id),nil
}

func main() {
	w := NewWorker(192172100132,192168168168)
	id,_ := w.NextID()
	fmt.Println(id)
}
