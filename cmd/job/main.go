package main

import (
	"context"
	"fmt"
	"message/cmd/common"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
)

// kafka consumer
func main() {
	closer := common.InitJaeger("job")
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("job-service")
	defer span.Finish()

	// 将span生成context
	rootCtx := opentracing.ContextWithSpan(context.Background(), span)

	cf := common.GetConfig()
	consumer, err := sarama.NewConsumer([]string{cf.Kafka.Addr}, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		span.LogKV("consumer start fail:", err)
		return
	}

	partitionList, err := consumer.Partitions(cf.Kafka.Topic) // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
		return
	}

	// 处理job信息
	for partition := range partitionList { // 遍历所有的分区
		// 针对每个分区创建一个对应的分区消费者
		pc, err := consumer.ConsumePartition(cf.Kafka.Topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
			return
		}
		defer pc.AsyncClose()
		// 异步从每个分区消费信息
		go func(context.Context, sarama.PartitionConsumer) {
			span1, _ := opentracing.StartSpanFromContext(rootCtx, "PartitionConsumer")
			defer span1.Finish()
			span1.LogKV("consumer", "success")

			for msg := range pc.Messages() {
				fmt.Printf(
					"Partition:%d Offset:%d Key:%v Value:%v\n",
					msg.Partition,
					msg.Offset,
					string(msg.Key),
					string(msg.Value),
				)
			}
		}(rootCtx, pc)
	}
}
