package rag

import (
	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
    "github.com/cloudwego/eino/components/document"
	"context"
    "regexp"
    "strings"
    "io"
	"sync"
)

type cleanParser struct{}

var (
    Loader *file.FileLoader
    once sync.Once
    Splitter document.Transformer
    cleanPaser parser.Parser = &cleanParser{}
    re = regexp.MustCompile(`(?s)<!--.*?-->\r?\n?`)
    reHTML = regexp.MustCompile(`(?s)<script.*?>.*?</script>|<style.*?>.*?</style>|<[^>]+>\r?\n?`)
)

func (dp cleanParser) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	opt := parser.GetCommonOptions(&parser.Options{}, opts...)

	meta := make(map[string]any)
	meta["_source"] = opt.URI

	for k, v := range opt.ExtraMeta {
		meta[k] = v
	}

    cleaned := string(data)
    cleaned = cleanHTMLEntities(cleaned)
    cleaned = re.ReplaceAllString(cleaned, "")
    cleaned = reHTML.ReplaceAllString(cleaned, "")
    
	doc := &schema.Document{
		Content:  cleaned,
		MetaData: meta,
	}

	return []*schema.Document{doc}, nil
}

func LoaderInit() (error) {
    var err error
    once.Do(func() {
        Loader, err = file.NewFileLoader(
            context.Background(),
            &file.FileLoaderConfig{
                UseNameAsID: true,
                Parser:      cleanPaser,
            },
        )
        Splitter, err = recursive.NewSplitter(context.Background(), &recursive.Config{
            ChunkSize:   1000,             // 每段最大长度
            OverlapSize: 0,              // 上下文重叠
            Separators: []string{
                "\n\n",  // 段落分隔
                "\n#",   // Markdown 标题
                "\n- ", "\n* ", // 列表项
                "\n> ",  // 引用块
                "\n```", // 代码块开始/结束
            },
            KeepType: recursive.KeepTypeEnd, // 分隔符保留在片段尾部
        })
    })
    if err!=nil {
        return err
    }
    return nil
}


func cleanHTMLEntities(s string) string {
    entities := map[string]string{
        "&nbsp;": " ",
        "&gt;":   ">",
        "&lt;":   "<",
        "&quot;": `"`,
        "\r\n":  "\n",
    }

    var b strings.Builder
    b.Grow(len(s))

    for i := 0; i < len(s); {
        matched := false
        for k, v := range entities {
            if strings.HasPrefix(s[i:], k) {
                b.WriteString(v)
                i += len(k)
                matched = true
                break
            }
        }
        if !matched {
            b.WriteByte(s[i])
            i++
        }
    }
    return b.String()
}
