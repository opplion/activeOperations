package rag

import (
	"context"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"activeOperations/config"
	"log"
)

func ReloadRAG() error {
	ctx := context.Background()
	err := Client.DropCollection(ctx, milvusclient.NewDropCollectionOption(config.GetConfig().Milvus.CollectionName))
	if err != nil {
		return err
	}
	err = Client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(config.GetConfig().Milvus.CollectionName, Schema).
		WithIndexOptions(IndexOptions...))
	if err != nil {
		return err
	}

	task, err := Client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(config.GetConfig().Milvus.CollectionName))
	if err != nil {
		return err
	}
	err = task.Await(ctx)
	if err != nil {
		return err
	}
	log.Println("Milvus collection reloaded successfully")
	return nil
}
