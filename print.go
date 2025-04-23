package pricey

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"

	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
	qrsvg "github.com/wamuir/svg-qr-code"
)

func print(client *gotenberg.Client, htmlDoc io.Reader) (io.ReadCloser, error) {
	index, err := document.FromReader("index", htmlDoc)
	if err != nil {
		return nil, err
	}
	footer, err := document.FromString("footer", footerTemplate)
	if err != nil {
		return nil, err
	}
	req := gotenberg.NewHTMLRequest(index)
	req.Footer(footer)

	resp, err := client.Send(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to send: %w", err)
	}
	return resp.Body, nil
}

type priceyPrint struct {
	store            Store
	pdfClient        *gotenberg.Client
	standardTemplate *template.Template
}

//go:embed templates/standard.html
var standardTemplate string

//go:embed templates/footer.html
var footerTemplate string

func newPrinter(store Store, pdfClient *gotenberg.Client) *priceyPrint {
	funcs := template.FuncMap{
		"pennies":          pennies,
		"quantity":         quantity,
		"depthPadding":     depthPadding,
		"adjustmentAmount": adjustmentAmount,
		"qrcode":           qrcode,
	}
	standardTemplate, err := template.New("standard").Funcs(funcs).Parse(standardTemplate)
	if err != nil {
		panic("failed to parse standard template: " + err.Error())
	}

	return &priceyPrint{
		store:            store,
		pdfClient:        pdfClient,
		standardTemplate: standardTemplate,
	}
}

func pennies(v int) string {
	if v == 0 {
		return ""
	}
	d := fmt.Sprintf("%d", v/100)
	ld := len(d)
	sb := strings.Builder{}
	for i := ld - 1; i >= 0; i-- {
		c := d[i]
		if ld-i > 1 && (ld-i)%3 == 1 {
			sb.WriteString(",")
		}
		sb.WriteByte(c)
	}
	d = sb.String()
	ld = len(d)
	sb = strings.Builder{}
	for i := ld - 1; i >= 0; i-- {
		sb.WriteByte(d[i])
	}

	pens := v % 100
	return fmt.Sprintf("%s.%02d", sb.String(), pens)
}

func quantity(v int) string {
	if v == 0 {
		return ""
	}
	whole := v / 100
	frac := v % 100
	if frac == 0 {
		return fmt.Sprintf("%d", whole)
	} else if frac%10 == 0 {
		return fmt.Sprintf("%d.%d", whole, frac/10)
	}
	return fmt.Sprintf("%d.%d", whole, frac)
}

func depthPadding(v int, paddingPixels, base int) int {
	return (v * paddingPixels) + base
}

func adjustmentAmount(a *Adjustment, subTotal int) int {
	if a == nil {
		return 0
	}
	if a.Type == AdjustmentTypeFlat {
		return a.Amount
	}
	if a.Type == AdjustmentTypePercent {
		return subTotal * a.Amount / 100
	}
	return 0
}

func qrcode(url string) template.HTML {
	qr, err := qrsvg.New(url)
	if err != nil {
		fmt.Printf("failed to generate qrcode: %v", err)
		return ""
	}
	if qr == nil {
		fmt.Printf("qrcode was empty for content: %s", url)
		return ""
	}
	qr.Borderwidth = 0
	svgObj := qr.SVG()
	return template.HTML(strings.Replace(qr.SVG().String(), fmt.Sprintf("width=\"%d\" height=\"%d\"", svgObj.Width, svgObj.Height), fmt.Sprintf("viewBox=\"0 0 %d %d\"", svgObj.Width, svgObj.Height), 1))
}

type item struct {
	id    int
	l     *LineItem
	fl    *PrintableLineItem
	found bool
}

func (v *priceyPrint) GetPrintableQuote(ctx context.Context, id int) (*PrintableQuote, error) {
	fullQuote := &PrintableQuote{}
	return fullQuote, v.store.Transaction(func(ctx context.Context) error {
		quote, err := v.store.GetQuote(ctx, id)
		if err != nil {
			return err
		}
		images := map[int]*Image{}
		contacts := map[int]*Contact{}
		lineItems := map[int]*LineItem{}
		adjustments := map[int]*Adjustment{}
		if quote.LogoId >= 0 {
			imageUrl, err := v.store.GetImageUrl(ctx, quote.LogoId)
			if err != nil {
				return err
			}
			if imageUrl != "" {
				images[quote.LogoId] = &Image{
					Id:  quote.LogoId,
					Url: imageUrl,
				}
			}
		}
		if quote.SenderId >= 0 {
			c, err := v.store.GetContact(ctx, quote.SenderId)
			if err != nil {
				return err
			}
			contacts[quote.SenderId] = c
		}
		if quote.BillToId >= 0 {
			c, err := v.store.GetContact(ctx, quote.BillToId)
			if err != nil {
				return err
			}
			contacts[quote.BillToId] = c
		}
		if quote.ShipToId >= 0 {
			c, err := v.store.GetContact(ctx, quote.ShipToId)
			if err != nil {
				return err
			}
			contacts[quote.ShipToId] = c
		}

		for _, lineItemId := range quote.LineItemIds {
			lineItem, err := v.store.GetLineItem(ctx, lineItemId)
			if err != nil {
				return err
			}
			lineItems[lineItemId] = lineItem
			if lineItem.ImageId != nil && images[*lineItem.ImageId] == nil {
				imgUrl, err := v.store.GetImageUrl(ctx, *lineItem.ImageId)
				if err != nil {
					return err
				}
				if imgUrl != "" {
					images[*lineItem.ImageId] = &Image{Id: *lineItem.ImageId, Url: imgUrl}
				}
			}
		}

		for _, adjustmentId := range quote.AdjustmentIds {
			adjustment, err := v.store.GetAdjustment(ctx, adjustmentId)
			if err != nil {
				return err
			}
			adjustments[adjustmentId] = adjustment
		}

		fullQuote = v.getPrintableQuote(quote, images, contacts, lineItems, adjustments)

		return nil
	})
}

func (v *priceyPrint) getPrintableQuote(quote *Quote, images map[int]*Image, contacts map[int]*Contact, lineItems map[int]*LineItem, adjustments map[int]*Adjustment) *PrintableQuote {
	q := &PrintableQuote{Id: quote.Id}
	q.Code = quote.Code
	q.OrderNumber = quote.OrderNumber
	q.PrimaryBackgroundColor = quote.PrimaryBackgroundColor
	q.PrimaryTextColor = quote.PrimaryTextColor
	q.IssueDate = quote.IssueDate
	q.ExpirationDate = quote.ExpirationDate
	q.PaymentTerms = quote.PaymentTerms
	q.Notes = quote.Notes
	q.BalanceDueOn = quote.BalanceDueOn
	q.PayUrl = quote.PayUrl
	q.Sent = quote.Sent
	q.SentOn = quote.SentOn
	q.Sold = quote.Sold
	q.SoldOn = quote.SoldOn
	q.Created = quote.Created
	q.Updated = quote.Updated
	q.Hidden = quote.Hidden
	q.Locked = quote.Locked

	if quote.LogoId >= 0 {
		q.Logo = images[quote.LogoId]
	}
	if quote.SenderId >= 0 {
		q.Sender = contacts[quote.SenderId]
	}
	if quote.BillToId >= 0 {
		q.BillTo = contacts[quote.BillToId]
	}
	if quote.ShipToId >= 0 {
		q.ShipTo = contacts[quote.ShipToId]
	}

	var items []*item
	lookup := map[int]*item{}
	for _, lineItemId := range quote.LineItemIds {
		l, fl := v.getPrintableLineItem(lineItems, images, lineItemId)
		i := &item{l: l, fl: fl, found: false}
		lookup[lineItemId] = i
		if l != nil {
			items = append(items, i)
			if l.ParentId == nil {
				q.LineItems = append(q.LineItems, fl)
				fl.Depth = 0
				fl.Number = strconv.Itoa(len(q.LineItems))
				i.found = true
			}
		}
	}

	for _, item := range items {
		for _, subItemId := range item.l.SubItemIds {
			if subItem, ok := lookup[subItemId]; ok && subItem != nil && subItem.l.ParentId != nil && *subItem.l.ParentId == item.l.Id {
				item.fl.SubItems = append(item.fl.SubItems, subItem.fl)
			}
		}
	}

	visited := map[int]bool{}
	for i, item := range q.LineItems {
		q.SubTotal += v.findAmount(lookup, item, map[int]bool{})
		v.calculateDepthAndNumber("", 0, i, item, visited)
	}

	q.Total = q.SubTotal
	for _, adjustmentId := range quote.AdjustmentIds {
		a := adjustments[adjustmentId]
		if a != nil {
			q.Adjustments = append(q.Adjustments, a)
			q.Total += adjustmentAmount(a, q.SubTotal)
		}
	}

	if quote.BalanceDue != 0 {
		q.BalanceDue = quote.BalanceDue
	} else if quote.BalancePercentDue != 0 {
		q.BalanceDue = q.Total * quote.BalancePercentDue / 100
	}

	return q
}

func (v *priceyPrint) getPrintableLineItem(lineItems map[int]*LineItem, images map[int]*Image, id int) (*LineItem, *PrintableLineItem) {
	l := lineItems[id]
	var fl *PrintableLineItem
	if l != nil {
		var i *Image
		if l.ImageId != nil {
			i = images[*l.ImageId]
		}
		fl = &PrintableLineItem{
			Id:              l.Id,
			Depth:           0,
			Image:           i,
			Description:     l.Description,
			QuantityPrefix:  l.QuantityPrefix,
			Quantity:        l.Quantity,
			QuantitySuffix:  l.QuantitySuffix,
			UnitPricePrefix: l.UnitPricePrefix,
			UnitPrice:       l.UnitPrice,
			UnitPriceSuffix: l.UnitPriceSuffix,
			AmountPrefix:    l.AmountPrefix,
			Amount:          0,
			AmountSuffix:    l.AmountSuffix,
			Created:         l.Created,
			Updated:         l.Updated,
		}
		if l.Amount != nil {
			fl.AmountOverridden = true
			fl.Amount = *l.Amount
		} else if l.Quantity > 0 {
			fl.AmountOverridden = true
			fl.Amount = l.UnitPrice * l.Quantity / 100
		}
	}
	return l, fl
}

func (v *priceyPrint) findAmount(lookup map[int]*item, item *PrintableLineItem, visited map[int]bool) int {
	visited[item.Id] = true
	i := lookup[item.Id]
	if i == nil {
		return 0
	}
	if i.fl.AmountOverridden {
		return item.Amount
	}
	for _, subItem := range item.SubItems {
		if !visited[subItem.Id] {
			item.Amount += v.findAmount(lookup, subItem, visited)
		}
	}
	return item.Amount
}

func (v *priceyPrint) calculateDepthAndNumber(parentNumber string, depth, index int, item *PrintableLineItem, visited map[int]bool) {
	if visited[item.Id] {
		return
	}
	visited[item.Id] = true
	item.Depth = depth
	var sb []string
	if parentNumber != "" {
		sb = append(sb, parentNumber)
	}
	sb = append(sb, strconv.Itoa(index+1))
	item.Number = strings.Join(sb, ".")
	for i, subItem := range item.SubItems {
		v.calculateDepthAndNumber(item.Number, depth+1, i, subItem, visited)
	}
}

func (v *priceyPrint) Standard(ctx context.Context, id int, w io.Writer) error {
	q, err := v.GetPrintableQuote(ctx, id)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = v.standardTemplate.Execute(&buf, q)
	if err != nil {
		return err
	}
	resp, err := print(v.pdfClient, &buf)
	if err != nil {
		return err
	}
	defer resp.Close()
	_, err = io.Copy(w, resp)
	if err != nil {
		return err
	}
	return nil
}

func (v *priceyPrint) StandardHTML(ctx context.Context, id int, w io.Writer) error {
	q, err := v.GetPrintableQuote(ctx, id)
	if err != nil {
		return err
	}
	return v.standardTemplate.Execute(w, q)
}
