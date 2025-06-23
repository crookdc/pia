package squeak

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilder_UnmarshalXML(t *testing.T) {
	data := `
	<?xml version="1.0"?>
	<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
		<soap:Body xmlns:m="http://www.example.org/stock">
		  <m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
			<m:Price>36.5</m:Price>
			<m:Price>66.5</m:Price>
		  </m:GetStockPriceResponse>
		</soap:Body>
	</soap:Envelope>
	`
	builder := Builder{}
	err := xml.Unmarshal([]byte(data), &builder)
	assert.Nil(t, err)
	assert.Equal(t, &ObjectInstance{
		Properties: map[string]Object{
			"_attributes": &ObjectInstance{
				Properties: map[string]Object{
					"soap":          String{"http://www.w3.org/2003/05/soap-envelope"},
					"encodingStyle": String{"http://www.w3.org/2003/05/soap-encoding"},
				},
			},
			"Body": &ObjectInstance{
				Properties: map[string]Object{
					"_attributes": &ObjectInstance{
						Properties: map[string]Object{
							"m": String{"http://www.example.org/stock"},
						},
					},
					"GetStockPriceResponse": &ObjectInstance{
						Properties: map[string]Object{
							"_attributes": &ObjectInstance{
								Properties: make(map[string]Object),
							},
							"Price": &List{
								slice: []Object{
									&ObjectInstance{
										Properties: map[string]Object{
											"_attributes": &ObjectInstance{
												Properties: make(map[string]Object),
											},
											"_inner": String{"34.5"},
										},
									},
									&ObjectInstance{
										Properties: map[string]Object{
											"_attributes": &ObjectInstance{
												Properties: make(map[string]Object),
											},
											"_inner": String{"36.5"},
										},
									},
									&ObjectInstance{
										Properties: map[string]Object{
											"_attributes": &ObjectInstance{
												Properties: make(map[string]Object),
											},
											"_inner": String{"66.5"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, builder.Object())
}
