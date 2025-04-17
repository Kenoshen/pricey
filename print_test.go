package pricey

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
)

func TestQuotePDF(t *testing.T) {
	os.MkdirAll("tmp", os.ModePerm)
	pdf, err := os.Create("tmp/test_standard_template.pdf")
	if err != nil {
		t.Error(err)
		return
	}
	defer pdf.Close()
	htmlFile, err := os.Create("tmp/test_standard_template.html")
	if err != nil {
		t.Error(err)
		return
	}
	defer htmlFile.Close()
	client, err := gotenberg.NewClient("http://localhost:3000", http.DefaultClient)
	if err != nil {
		t.Error(err)
		return
	}
	tmp := newPrinter(nil, nil).standardTemplate
	buf := bytes.Buffer{}

	m := func(tm time.Time) *time.Time {
		return &tm
	}
	f := func(fl float64) *float64 {
		return &fl
	}
	err = tmp.Execute(&buf, &FullQuote{
		Id:          64,
		Code:        "INV-001",
		OrderNumber: "O#0012345",
		Logo: &Image{
			Url: "https://raw.githubusercontent.com/sparksuite/simple-html-invoice-template/refs/heads/master/website/images/logo.png",
		},
		IssueDate:      m(time.Now().Add(-3 * 24 * time.Hour)),
		ExpirationDate: m(time.Now().Add(10 * 24 * time.Hour)),
		PaymentTerms:   "pay 1/2 up front, then 1/2 on delivery",
		Notes:          "Thanks so much for your business!",
		Sender: &Contact{
			Name:        "John Doe",
			CompanyName: "Acme Corp",
			Phones:      []string{"123-555-1234", "(555)555-5555"},
			Emails:      []string{"john@acme.org"},
			Websites:    []string{"acme.org/rockets"},
			Street:      "1234 Sunny Rd",
			City:        "Sunnyville",
			State:       "TX",
			Zip:         "12345",
		},
		BillTo: &Contact{
			Name:   "Wile E. Coyote",
			Street: "62nd Flatrock",
			City:   "Mojave Desert",
			State:  "Arizona",
			Zip:    "87654",
		},
		ShipTo: &Contact{
			Name:   "Wile E. Coyote",
			Street: "Left Side of the Road",
			City:   "Mojave Desert",
			State:  "Arizona",
			Zip:    "87654",
		},
		LineItems: []*FullLineItem{{
			Image: &Image{
				Url: "https://cdn11.bigcommerce.com/s-bip927t4m2/images/stencil/1280x1280/products/1667/3417/s-l500__38868.1677889557.png?c=2",
			},
			Description: "Acme Rocket Patch",
			Quantity:    10.5,
			UnitPrice:   32.423,
			Amount:      f(200.32),
			Open:        false,
		}, {
			SubItems: []*FullLineItem{{
				Description: "Shipping",
				Amount:      f(20.0),
			}, {
				Description: "Handling",
				Amount:      f(10.0),
			}},
			Description: "Shipping and Handling",
			Quantity:    1,
			Amount:      f(30),
			Open:        true,
		}},
		SubTotal: 233.4,
		Adjustments: []*Adjustment{{
			Description: "Taxes",
			Type:        AdjustmentTypePercent,
			Amount:      0.7,
		}, {
			Description: "Heavy Equipment Fee",
			Type:        AdjustmentTypeFlat,
			Amount:      33.33,
		}},
		Total:             564.82,
		BalancePercentDue: 0.5,
		BalanceDueOn:      m(time.Now().Add(10 * 24 * time.Hour)),
		PayUrl:            "http://google.com",
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = io.Copy(htmlFile, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
		return
	}
	resp, err := print(client, &buf)
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Close()
	_, err = io.Copy(pdf, resp)
	if err != nil {
		t.Error(err)
		return
	}
}
