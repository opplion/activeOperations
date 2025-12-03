package rag

import (
	"activeOperations/config"
	"context"
	"sync"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
)


type Embedder struct {
    inner *dashscope.Embedder
}
var (
    embedder *Embedder
    onceEmbed sync.Once
)

func LoadEmbedder() error {
    var initErr error
    onceEmbed.Do(func() {
        emCfg := config.GetConfig().Embedding
        em, err := dashscope.NewEmbedder(context.Background(), &dashscope.EmbeddingConfig{
            APIKey:     emCfg.Apikey,
            Model:      emCfg.Model,
            Dimensions: emCfg.Dimensions,
        })
        if err != nil {
            initErr = err
            return
        }
        embedder = NewFloat32Embedder(em)
        // if embedder == nil {
        //     initErr =  err
        //     return
        // }
    })

    if initErr != nil {
        return initErr
    }
    return nil
}

func NewFloat32Embedder(inner *dashscope.Embedder) *Embedder {
    return &Embedder{inner: inner}
}

func (f *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    v64, err := f.inner.EmbedStrings(ctx,texts)
    if err != nil {
        return nil, err
    }

    v32 := make([][]float32, len(v64))
    for i := range v64 {
        v32[i] = make([]float32, len(v64[i]))
        for j := range v64[i] {
            v32[i][j] = float32(v64[i][j])
        }
    }

    return v32, nil
}

func GetEmbedder() *Embedder {
	return embedder
}