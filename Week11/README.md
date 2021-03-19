学习笔记

// 翻转二叉树
func invertTree(root *TreeNode) *TreeNode {
    // 递归终止条件
    if root == nil {
        return nil
    }

    // 递归过程
    root.Left, root.Right = root.Right, root.Left
    root.Left = invertTree(root.Left)
    root.Right = invertTree(root.Right)

    return root
}

1、个人基础素质（性格、沟通、战斗力、执行力、态度）；
2、个人职业规划和想法；
3、对过去项目的总结和思考（对象存储）
4、操作系统简单原理、网络知识、简单数学知识、Runtime

https://leetcode-cn.com/problems/min-stack/
LRU、LFU、FIFO算法总结
https://blog.csdn.net/u013126379/article/details/52356431

Queue Kafka 消息队列
Cache Redis/Memcached CacheProxy
KV RocksDB、LevelDB、Raft
SQL MySQL 、DBProxy（Vitess）
Config 配置中心
Discovery （Eureka、Nacos、etcd）
服务治理 （可用性、元数据）
调度 Yarn、K8s、Mesos

大数据：存储（HDFS）、计算（Hive、Spark、Presto）、调度：Yarn、OLAP Clickhouse、Kylin、Doris；数仓 HQL、数仓建模