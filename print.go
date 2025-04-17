package pricey

import (
	"context"
	"fmt"
	"io"

	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

func print(client *gotenberg.Client, htmlDoc io.Reader) (io.ReadCloser, error) {
	index, err := document.FromReader("some name", htmlDoc)
	if err != nil {
		return nil, err
	}
	fmt.Println("doc:", index.Filename())

	req := gotenberg.NewHTMLRequest(index)

	// req.Margins(gotenberg.NormalMargins)
	// req.Scale(1.0)
	// req.PaperSize(gotenberg.Letter)

	resp, err := client.Send(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to send: %w", err)
	}

	return resp.Body, nil
}
