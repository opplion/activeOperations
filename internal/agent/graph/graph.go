package graph

import (
	"activeOperations/internal/agent/rag"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

var (
	globalGraph compose.Runnable[string, *schema.StreamReader[*schema.Message]]
	globalGraphErr error
	onceGraph   sync.Once
)

func GetWorkflow() (compose.Runnable[string, *schema.StreamReader[*schema.Message]], error) {
	onceGraph.Do(func() {
		ctx := context.Background()
		var err error
		wf := compose.NewWorkflow[string, *schema.StreamReader[*schema.Message]]()
		wf.AddLambdaNode("MilvusRetriever", compose.InvokableLambda(MilvusRetrieverHandler)).AddInput(compose.START)
		wf.AddLambdaNode("merge", compose.InvokableLambda(lambdaHandler)).AddInput("MilvusRetriever")
		wf.AddLambdaNode("QwenModel", compose.InvokableLambda(ReactAgentHandler)).AddInput("merge")
		wf.End().AddInput("QwenModel")

		globalGraph,err = wf.Compile(ctx)
		if err!=nil {
			globalGraphErr = fmt.Errorf("failed to compile graph: %v", err)
			return
		}
	})
	return globalGraph,globalGraphErr
}

func ReloadRAG() error {
	ctx:= context.Background()
	chain:= compose.NewChain[document.Source, []string]()
	_ = chain.AppendLoader(rag.Loader)
	_ = chain.AppendDocumentTransformer(rag.Splitter)
	_ = chain.AppendIndexer(rag.MilvusSDK)
	run,err := chain.Compile(ctx)
	if err!=nil {
		return err
	}
	err = rag.DefaultForword()
	if err!=nil {
		return err
	}
	var files []string
	root := "./docs"
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	for i, path := range files {
		_, err := run.Invoke(ctx, document.Source{URI: path})
		fmt.Printf("Processing file %d/%d: %s\n", i+1, len(files), path)
		if err!=nil {
			//回滚
			rag.DefaultBackword()
			return err
		}
	}
	rag.DefaultFinal()
	return nil
}

