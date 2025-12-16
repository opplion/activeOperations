package rag

import (
	"context"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"activeOperations/config"
)

func DefaultForword() error {
	testCollection := MilvusSDK.Reload()
	ctx := context.Background()
	err := Client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(testCollection, Schema).
		WithIndexOptions(NewIndexOptions(testCollection)...))
	if err != nil {
		return err
	}

	task, err := Client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(testCollection))
	if err != nil {
		return err
	}
	err = task.Await(ctx)
	if err != nil {
		return err
	}
	return nil
}

func DefaultBackword() error {
	ctx := context.Background()
	err := Client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(MilvusSDK.test_Collection))
	if err != nil {
		return err
    }
	err = Client.DropCollection(ctx, milvusclient.NewDropCollectionOption(MilvusSDK.test_Collection))
	if err != nil {
		return err
	}
	return nil
}

func DefaultFinal() error {
	oldCollection:=MilvusSDK.Publish()
	ctx := context.Background()
	err := Client.AlterAlias(ctx, milvusclient.NewAlterAliasOption(config.GetConfig().Milvus.CollectionName, MilvusSDK.online_Collection))
	if err != nil {
		return err
	}
	err = Client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(oldCollection))
    if err != nil {
		return err
    }
	err = Client.DropCollection(ctx, milvusclient.NewDropCollectionOption(oldCollection))
	if err != nil {
		return err
	}
	return nil
}