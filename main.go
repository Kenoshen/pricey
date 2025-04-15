package pricey

import (
	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
)

type Pricey struct {
	store     Store
	pdfClient *gotenberg.Client
}

func New(store Store, pdfClient *gotenberg.Client) Pricey {
	return Pricey{store: store, pdfClient: pdfClient}
}

func (p *Pricey) Get() {
}
