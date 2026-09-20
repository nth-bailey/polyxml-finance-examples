package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	pacs008 "github.com/nth-bailey/polyxml-finance-examples/generated/go"
)

// PaymentIntentDTO models the incoming FedNow / Stripe instant payment intent JSON
type PaymentIntentDTO struct {
	MessageID        string `json:"message_id"`
	CreatedAt        string `json:"created_at"`
	SettlementMethod string `json:"settlement_method"`
	ClearingNetwork  string `json:"clearing_network"`
	Payment          struct {
		InstructionID  string  `json:"instruction_id"`
		EndToEndID     string  `json:"end_to_end_id"`
		TransactionID  string  `json:"transaction_id"`
		UETR           string  `json:"uetr"`
		Amount         float64 `json:"amount"`
		Currency       string  `json:"currency"`
		SettlementDate string  `json:"settlement_date"`
		ChargeBearer   string  `json:"charge_bearer"`
		Debtor         struct {
			Name    string `json:"name"`
			Address struct {
				StreetName     string `json:"street_name"`
				BuildingNumber string `json:"building_number"`
				PostCode       string `json:"post_code"`
				TownName       string `json:"town_name"`
				Country        string `json:"country"`
			} `json:"address"`
			CountryOfResidence string `json:"country_of_residence"`
		} `json:"debtor"`
		DebtorAccount struct {
			AccountNumber string `json:"account_number"`
			Currency      string `json:"currency"`
			AccountName   string `json:"account_name"`
		} `json:"debtor_account"`
		DebtorAgent struct {
			BICFI            string `json:"bicfi"`
			ClearingSystemID string `json:"clearing_system_id"`
			RoutingNumber    string `json:"routing_number"`
			Name             string `json:"name"`
		} `json:"debtor_agent"`
		CreditorAgent struct {
			BICFI            string `json:"bicfi"`
			ClearingSystemID string `json:"clearing_system_id"`
			RoutingNumber    string `json:"routing_number"`
			Name             string `json:"name"`
		} `json:"creditor_agent"`
		Creditor struct {
			Name    string `json:"name"`
			Address struct {
				StreetName     string `json:"street_name"`
				BuildingNumber string `json:"building_number"`
				PostCode       string `json:"post_code"`
				TownName       string `json:"town_name"`
				Country        string `json:"country"`
			} `json:"address"`
			CountryOfResidence string `json:"country_of_residence"`
		} `json:"creditor"`
		CreditorAccount struct {
			AccountNumber string `json:"account_number"`
			Currency      string `json:"currency"`
			AccountName   string `json:"account_name"`
		} `json:"creditor_account"`
		PurposeCode string `json:"purpose_code"`
		Remittance  string `json:"remittance"`
	} `json:"payment"`
}

func ptr[T any](v T) *T {
	return &v
}

func main() {
	fmt.Println("================================================================================")
	fmt.Println("  PolyXML Go Showcase: FedNow Payment Intent -> ISO 20022 pacs.008.001.10")
	fmt.Println("================================================================================")

	// Locate data directory relative to executable or current working directory
	dataPaths := []string{
		"data/payment_intent_fednow.json",
		"../../data/payment_intent_fednow.json",
	}
	var dataBytes []byte
	var err error
	for _, p := range dataPaths {
		if b, e := os.ReadFile(p); e == nil {
			dataBytes = b
			break
		}
	}
	if len(dataBytes) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Could not locate data/payment_intent_fednow.json\n")
		os.Exit(1)
	}

	// 1. Ingest FinTech Payment Intent JSON
	var intent PaymentIntentDTO
	if err := json.Unmarshal(dataBytes, &intent); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse payment intent JSON: %v\n", err)
		os.Exit(1)
	}

	creDtTm, _ := time.Parse(time.RFC3339, intent.CreatedAt)
	sttlmDt, _ := time.Parse("2006-01-02", intent.Payment.SettlementDate)

	// 2. Adapt into PolyXML-generated ISO 20022 pacs.008 data model
	doc := pacs008.FiToFiCustomerCreditTransfer{
		XMLName: xml.Name{
			Space: "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10",
			Local: "Document",
		},
		GrpHdr: pacs008.GroupHeader{
			MsgID:   intent.MessageID,
			CreDtTm: creDtTm,
			NbOfTxs: 1,
			SttlmInf: pacs008.SettlementInstruction{
				SettlementMethod: pacs008.SettlementMethodCode(intent.SettlementMethod),
				ClearingSystem:   ptr(intent.ClearingNetwork),
			},
		},
		CdtTrfTxInf: []pacs008.CreditTransferTransactionInformation{
			{
				PmtID: pacs008.PaymentIdentification{
					InstructionID: intent.Payment.InstructionID,
					EndToEndID:    intent.Payment.EndToEndID,
					TransactionID: intent.Payment.TransactionID,
					Uetr:          intent.Payment.UETR,
				},
				IntrBkSttlmAmt: pacs008.ActiveOrHistoricCurrencyAndAmount{
					Currency: pacs008.ActiveCurrencyCode(intent.Payment.Currency),
					Value:    intent.Payment.Amount,
				},
				IntrBkSttlmDt: sttlmDt,
				ChargeBearer:  pacs008.ChargeBearerType(intent.Payment.ChargeBearer),
				Dbtr: pacs008.PartyIdentification{
					Name: intent.Payment.Debtor.Name,
					PostalAddress: &pacs008.PostalAddress{
						StreetName:     ptr(intent.Payment.Debtor.Address.StreetName),
						BuildingNumber: ptr(intent.Payment.Debtor.Address.BuildingNumber),
						PostCode:       ptr(intent.Payment.Debtor.Address.PostCode),
						TownName:       intent.Payment.Debtor.Address.TownName,
						Country:        intent.Payment.Debtor.Address.Country,
					},
					CountryOfResidence: ptr(intent.Payment.Debtor.CountryOfResidence),
				},
				DbtrAcct: pacs008.CashAccount{
					ID: pacs008.AccountIdentification{
						ProprietaryAccount: ptr(intent.Payment.DebtorAccount.AccountNumber),
					},
					Currency: ptr(pacs008.ActiveCurrencyCode(intent.Payment.DebtorAccount.Currency)),
					Name:     ptr(intent.Payment.DebtorAccount.AccountName),
				},
				DbtrAgt: pacs008.BranchAndFinancialInstitutionIdentification{
					FinInstnID: pacs008.FinancialInstitutionIdentification{
						Bicfi: ptr(intent.Payment.DebtorAgent.BICFI),
						ClearingSystemMemberID: &pacs008.ClearingSystemMemberIdentification{
							ClearingSystemID: intent.Payment.DebtorAgent.ClearingSystemID,
							MemberID:         intent.Payment.DebtorAgent.RoutingNumber,
						},
						Name: ptr(intent.Payment.DebtorAgent.Name),
					},
				},
				CdtrAgt: pacs008.BranchAndFinancialInstitutionIdentification{
					FinInstnID: pacs008.FinancialInstitutionIdentification{
						Bicfi: ptr(intent.Payment.CreditorAgent.BICFI),
						ClearingSystemMemberID: &pacs008.ClearingSystemMemberIdentification{
							ClearingSystemID: intent.Payment.CreditorAgent.ClearingSystemID,
							MemberID:         intent.Payment.CreditorAgent.RoutingNumber,
						},
						Name: ptr(intent.Payment.CreditorAgent.Name),
					},
				},
				Cdtr: pacs008.PartyIdentification{
					Name: intent.Payment.Creditor.Name,
					PostalAddress: &pacs008.PostalAddress{
						StreetName:     ptr(intent.Payment.Creditor.Address.StreetName),
						BuildingNumber: ptr(intent.Payment.Creditor.Address.BuildingNumber),
						PostCode:       ptr(intent.Payment.Creditor.Address.PostCode),
						TownName:       intent.Payment.Creditor.Address.TownName,
						Country:        intent.Payment.Creditor.Address.Country,
					},
					CountryOfResidence: ptr(intent.Payment.Creditor.CountryOfResidence),
				},
				CdtrAcct: pacs008.CashAccount{
					ID: pacs008.AccountIdentification{
						ProprietaryAccount: ptr(intent.Payment.CreditorAccount.AccountNumber),
					},
					Currency: ptr(pacs008.ActiveCurrencyCode(intent.Payment.CreditorAccount.Currency)),
					Name:     ptr(intent.Payment.CreditorAccount.AccountName),
				},
				Purpose: ptr(intent.Payment.PurposeCode),
				RmtInf: &pacs008.RemittanceInformation{
					Unstructured: ptr(intent.Payment.Remittance),
				},
			},
		},
	}

	// 3. Validation
	if err := doc.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✔ ISO 20022 pacs.008 schema validation passed!")

	// 4. Serialize to XML
	xmlBytes, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "XML marshal error: %v\n", err)
		os.Exit(1)
	}
	xmlOutput := xml.Header + string(xmlBytes)

	// 5. Serialize to JSON using PolyXML struct tags
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
		os.Exit(1)
	}

	// 6. Round-trip Deserialization from JSON
	var roundtripDoc pacs008.FiToFiCustomerCreditTransfer
	if err := json.Unmarshal(jsonBytes, &roundtripDoc); err != nil {
		fmt.Fprintf(os.Stderr, "JSON unmarshal error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGenerated XML Payload (size: %d bytes):\n", len(xmlOutput))
	lines := strings.Split(xmlOutput, "\n")
	previewCount := 25
	if len(lines) < previewCount {
		previewCount = len(lines)
	}
	for i := 0; i < previewCount; i++ {
		fmt.Printf("  %s\n", lines[i])
	}
	if len(lines) > previewCount {
		fmt.Printf("  ... [%d lines truncated]\n", len(lines)-previewCount)
	}

	fmt.Printf("\nGenerated JSON Wire Representation (size: %d bytes):\n", len(jsonBytes))
	jsonLines := strings.Split(string(jsonBytes), "\n")
	previewCountJson := 20
	if len(jsonLines) < previewCountJson {
		previewCountJson = len(jsonLines)
	}
	for i := 0; i < previewCountJson; i++ {
		fmt.Printf("  %s\n", jsonLines[i])
	}
	if len(jsonLines) > previewCountJson {
		fmt.Printf("  ... [%d lines truncated]\n", len(jsonLines)-previewCountJson)
	}

	// 7. Micro-benchmarking
	const iterations = 10000

	startXml := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = xml.Marshal(doc)
	}
	xmlDuration := time.Since(startXml)
	xmlPerOp := float64(xmlDuration.Nanoseconds()) / float64(iterations) / 1000.0

	startJson := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = json.Marshal(doc)
	}
	jsonDuration := time.Since(startJson)
	jsonPerOp := float64(jsonDuration.Nanoseconds()) / float64(iterations) / 1000.0

	startDe := time.Now()
	for i := 0; i < iterations; i++ {
		var target pacs008.FiToFiCustomerCreditTransfer
		_ = json.Unmarshal(jsonBytes, &target)
	}
	deDuration := time.Since(startDe)
	dePerOp := float64(deDuration.Nanoseconds()) / float64(iterations) / 1000.0

	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Println("  PolyXML Go Performance Metrics (10,000 iterations)")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("  XML Serialization:       %8.2f µs/op\n", xmlPerOp)
	fmt.Printf("  JSON Serialization:      %8.2f µs/op\n", jsonPerOp)
	fmt.Printf("  JSON Deserialization:    %8.2f µs/op\n", dePerOp)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("  Debtor:     %s ($%.2f %s)\n", roundtripDoc.CdtTrfTxInf[0].Dbtr.Name,
		roundtripDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Value,
		roundtripDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency)
	fmt.Printf("  Creditor:   %s via %s (Routing: %s)\n",
		roundtripDoc.CdtTrfTxInf[0].Cdtr.Name,
		*roundtripDoc.CdtTrfTxInf[0].CdtrAgt.FinInstnID.Name,
		roundtripDoc.CdtTrfTxInf[0].CdtrAgt.FinInstnID.ClearingSystemMemberID.MemberID)
	fmt.Println("================================================================================")
	_ = filepath.Base
}
