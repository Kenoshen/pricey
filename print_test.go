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

func TestQuoteToFullQuote(t *testing.T) {
	q := &FullQuote{
		Id:          64,
		Code:        "INV-001",
		OrderNumber: "O#0012345",
		Logo: &Image{
			Url: "https://raw.githubusercontent.com/sparksuite/simple-html-invoice-template/refs/heads/master/website/images/logo.png",
		},
		IssueDate:      m(time.Now().Add(-3 * 24 * time.Hour)),
		ExpirationDate: m(time.Now().Add(10 * 24 * time.Hour)),
		PaymentTerms:   "pay 1/2 up front, then 1/2 on delivery",
		Notes:          "Thanks so much for your business!\n\nAnd thank you so much for being a valued customer who always pays on time and never requires us to track you down to try and get payment.",
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
			City:   "Mojave Desert Middle of Nowhere",
			State:  "Arizona",
			Zip:    "87654",
		},
		LineItems: []*FullLineItem{{
			Number: "1)",
			Depth:  0,
			Image: &Image{
				Url: "https://cdn11.bigcommerce.com/s-bip927t4m2/images/stencil/1280x1280/products/1667/3417/s-l500__38868.1677889557.png?c=2",
			},
			Description:     "Acme Rocket Patch",
			Quantity:        1050,
			UnitPrice:       3242,
			UnitPriceSuffix: "/unit",
			AmountPrefix:    "$",
			Amount:          20010032,
		}, {
			Number: "2)",
			Depth:  0,
			SubItems: []*FullLineItem{{
				Number:       "2.1)",
				Depth:        1,
				Description:  "Shipping",
				AmountPrefix: "$",
				Amount:       2000,
			}},
			Description:  "Shipping and Handling",
			Quantity:     100,
			AmountPrefix: "$",
			Amount:       3021210,
		}, {
			Number: "3)",
			Depth:  0,
			SubItems: []*FullLineItem{{
				Number:          "3.1)",
				Depth:           1,
				Quantity:        2000,
				QuantitySuffix:  " hours",
				UnitPricePrefix: "$",
				UnitPrice:       4000,
				UnitPriceSuffix: "/hr",
				Description:     "Labor & Disposal",
				AmountPrefix:    "$",
				Amount:          80000,
			}, {
				Number: "3.2)",
				Depth:  1,
				Image: &Image{
					Url: "https://images.thdstatic.com/productImages/4133747e-a5c0-4d5f-8c4e-a33409a0b804/svn/rheem-gas-tank-water-heaters-xg50t06he40u0-64_600.jpg",
				},
				Description:  "Water Heater",
				AmountPrefix: "$",
				Amount:       45000,
			}},
			Description:  "Replace Old Water Heater",
			Quantity:     100,
			AmountPrefix: "$",
			Amount:       332300,
		}},
		SubTotal: 23340,
		Adjustments: []*Adjustment{{
			Description: "Taxes",
			Type:        AdjustmentTypePercent,
			Amount:      70,
		}, {
			Description: "Heavy Equipment Fee",
			Type:        AdjustmentTypeFlat,
			Amount:      3333,
		}},
		Total:        577126482,
		BalanceDue:   1000000,
		BalanceDueOn: m(time.Now().Add(10 * 24 * time.Hour)),
		PayUrl:       "http://google.com",
	}

	// new newPrinter
	// printer.getFullQuote(q, images, contacts...
	// assert

}

func TestFullQuotePDF(t *testing.T) {
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
		Notes:          "Thanks so much for your business!\n\nAnd thank you so much for being a valued customer who always pays on time and never requires us to track you down to try and get payment.",
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
			City:   "Mojave Desert Middle of Nowhere",
			State:  "Arizona",
			Zip:    "87654",
		},
		LineItems: []*FullLineItem{{
			Number: "1)",
			Depth:  0,
			Image: &Image{
				Url: "https://cdn11.bigcommerce.com/s-bip927t4m2/images/stencil/1280x1280/products/1667/3417/s-l500__38868.1677889557.png?c=2",
			},
			Description:     "Acme Rocket Patch",
			Quantity:        1050,
			UnitPrice:       3242,
			UnitPriceSuffix: "/unit",
			AmountPrefix:    "$",
			Amount:          20010032,
		}, {
			Number: "2)",
			Depth:  0,
			SubItems: []*FullLineItem{{
				Number:       "2.1)",
				Depth:        1,
				Description:  "Shipping",
				AmountPrefix: "$",
				Amount:       2000,
			}},
			Description:  "Shipping and Handling",
			Quantity:     100,
			AmountPrefix: "$",
			Amount:       3021210,
		}, {
			Number: "3)",
			Depth:  0,
			SubItems: []*FullLineItem{{
				Number:          "3.1)",
				Depth:           1,
				Quantity:        2000,
				QuantitySuffix:  " hours",
				UnitPricePrefix: "$",
				UnitPrice:       4000,
				UnitPriceSuffix: "/hr",
				Description:     "Labor & Disposal",
				AmountPrefix:    "$",
				Amount:          80000,
			}, {
				Number: "3.2)",
				Depth:  1,
				Image: &Image{
					Url: "https://images.thdstatic.com/productImages/4133747e-a5c0-4d5f-8c4e-a33409a0b804/svn/rheem-gas-tank-water-heaters-xg50t06he40u0-64_600.jpg",
				},
				Description:  "Water Heater",
				AmountPrefix: "$",
				Amount:       45000,
			}},
			Description:  "Replace Old Water Heater",
			Quantity:     100,
			AmountPrefix: "$",
			Amount:       332300,
		}},
		SubTotal: 23340,
		Adjustments: []*Adjustment{{
			Description: "Taxes",
			Type:        AdjustmentTypePercent,
			Amount:      70,
		}, {
			Description: "Heavy Equipment Fee",
			Type:        AdjustmentTypeFlat,
			Amount:      3333,
		}},
		Total:        577126482,
		BalanceDue:   1000000,
		BalanceDueOn: m(time.Now().Add(10 * 24 * time.Hour)),
		PayUrl:       "http://google.com",
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
