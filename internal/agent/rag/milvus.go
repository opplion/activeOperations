package rag

import (
	"activeOperations/config"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"github.com/google/uuid"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"golang.org/x/sync/errgroup"
	"strings"
)

var (
	MilvusSDK *Retriever
	Client 	  *milvusclient.Client
	onceMilvus   sync.Once
)

type Retriever struct {
	Client      *milvusclient.Client
	ScoreThresh float32
	Embedding   *Embedder
	online_Collection  string
	test_Collection string
	lock       sync.Mutex
}

func MilvusInit() error {
	var err error
	ctx := context.Background()
	onceMilvus.Do(func() {
		err = doMilvusInit(ctx)
	})
	return err
}

func doMilvusInit(ctx context.Context) error {
	cfg := config.GetConfig().Milvus

	cli, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: cfg.Host + ":" + cfg.Port,
	})
	if err != nil {
		return err
	}

	// 尝试描述别名，判断别名是否存在
	res, aliasErr := cli.DescribeAlias(ctx, milvusclient.NewDescribeAliasOption(cfg.CollectionName))
	fmt.Println(aliasErr)
	aliasExists := aliasErr == nil

	var targetCollection string

	if !aliasExists {
		// 别名不存在，需要创建 collection 并绑定别名

		// 这里直接使用 cfg.CollectionName 作为 collection 名（也可以另起名）
		collectionName := NewUniqueCollectionName(cfg.CollectionName)
		err = cli.CreateCollection(
			ctx,
			milvusclient.NewCreateCollectionOption(collectionName, Schema).
			WithIndexOptions(NewIndexOptions(collectionName)...),
		)
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
		}

		err = cli.CreateAlias(
			ctx,
			milvusclient.NewCreateAliasOption(collectionName, cfg.CollectionName),
		)
		if err != nil {
			return fmt.Errorf("failed to create alias %s for collection %s: %w", cfg.CollectionName, collectionName, err)
		}

		targetCollection = collectionName
	} else {
		targetCollection = res.CollectionName
	}

	// 加载 collection
	loadTask, err := cli.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(targetCollection))
	if err != nil {
		return fmt.Errorf("failed to load collection %s: %w", targetCollection, err)
	}
	if err := loadTask.Await(ctx); err != nil {
		return fmt.Errorf("failed to await load of collection %s: %w", targetCollection, err)
	}

	Client = cli
	MilvusSDK = NewMilvusSDK(ctx, cli, 0.5, GetEmbedder(),targetCollection)

	return nil
}

// NewRetriever 创建 Retriever
func NewMilvusSDK(ctx context.Context, cli *milvusclient.Client, scoreThresh float32, emb *Embedder, collectionName string) *Retriever {
	return &Retriever{
		Client:      cli,
		ScoreThresh: scoreThresh,
		Embedding:   emb,
		online_Collection:  collectionName,
		test_Collection: "",
		lock: sync.Mutex{},
	}
}


func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	vectors, err := r.Embedding.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	searchOpt := milvusclient.NewSearchOption(
		r.online_Collection,
		5,
		[]entity.Vector{entity.FloatVector(vectors[0])},
	).WithConsistencyLevel(entity.ClSession).
	WithANNSField("vector").
	WithOutputFields("content")
	result, err := r.Client.Search(ctx, searchOpt)
	if err != nil {
		return nil, fmt.Errorf("milvus search failed: %w", err)
	}
	var results []*schema.Document
	for _, res := range result {
		content:=res.GetColumn("content").FieldData().GetScalars().GetStringData().GetData()
		for i,score := range res.Scores {
			if score < r.ScoreThresh {
				continue
			}
			doc := &schema.Document{
				Content:  content[i],
			}
			results = append(results, doc)
		}
	}
	return results, nil
}
const (
    batchSize    = 10
    maxWorkers   = 2 // 最大并发数，根据 DashScope QPM 调整
)

func (r *Retriever) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
    sem := make(chan struct{}, maxWorkers) // 信号量控制并发
    g, ctx := errgroup.WithContext(ctx)

    for i := 0; i < len(docs); i += batchSize {
        end := i + batchSize
        if end > len(docs) {
            end = len(docs)
        }
        batchDocs := docs[i:end]

        g.Go(func() error {
            sem <- struct{}{}        // 获取令牌
            defer func() { <-sem }() // 释放令牌

            srcText := make([]string, len(batchDocs))
            for j, doc := range batchDocs {
                srcText[j] = doc.Content
            }

            vectors, err := r.Embedding.Embed(ctx, srcText)
            if err != nil {
				lengths := make([]int, len(srcText))
				for i, text := range srcText {
					lengths[i] = len([]rune(text))
				}
				fmt.Printf("Embedding failed. Text lengths (in runes): %v\n", lengths)
                return err
            }

            // 3. 准备字段
            ids := make([]int64, len(batchDocs))
            contents := make([]string, len(batchDocs))
            metadata := make([][]byte, len(batchDocs))
            for j, d := range batchDocs {
                ids[j] = time.Now().UnixNano() + int64(i*1000+j)
                contents[j] = d.Content
                bytes, _ := json.Marshal(d.MetaData)
                metadata[j] = bytes
            }

            // 4. 插入 Milvus
            opt := milvusclient.NewColumnBasedInsertOption(r.test_Collection).
                WithInt64Column("id", ids).
                WithFloatVectorColumn("vector", len(vectors[0]), vectors).
                WithVarcharColumn("content", contents).
                WithColumns(column.NewColumnJSONBytes("metadata", metadata))

            _, err = r.Client.Insert(ctx, opt)
            if err != nil {
                return err
            }
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return nil, nil
}

func NewUniqueCollectionName(prefix string) string {
    uid := uuid.NewString()
	uid = strings.ReplaceAll(uid, "-", "")
    return fmt.Sprintf("%s_%s", prefix, uid)
}

func (r *Retriever) Publish() string {
	r.lock.Lock()
	defer r.lock.Unlock()
	oldCollection := r.online_Collection
	r.online_Collection = r.test_Collection
	return oldCollection
}

func (r *Retriever) Reload() string {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.test_Collection = NewUniqueCollectionName(config.GetConfig().Milvus.CollectionName)
	return r.test_Collection
}