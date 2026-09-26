package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	pacs008 "github.com/polyxml/polyxml-finance-examples/generated/go"
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

	// 4. Inherent XML Serialization
	startXml := time.Now()
	xmlBytes, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "XML marshal error: %v\n", err)
		os.Exit(1)
	}
	xmlOutput := xml.Header + string(xmlBytes)
	xmlDuration := time.Since(startXml)

	fmt.Printf("\n[1] Generated ISO 20022 pacs.008.001.10 XML Message (latency: %v):\n", xmlDuration)
	previewLen := 400
	if len(xmlOutput) < previewLen {
		previewLen = len(xmlOutput)
	}
	fmt.Printf("%s\n...\n\n", xmlOutput[:previewLen])

	// 5. Inherent Native JSON Serialization on Same Model
	startJson := time.Now()
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
		os.Exit(1)
	}
	jsonDuration := time.Since(startJson)

	fmt.Printf("[2] Generated Native JSON on Same Model (latency: %v):\n", jsonDuration)
	jsonPreviewLen := 400
	if len(jsonBytes) < jsonPreviewLen {
		jsonPreviewLen = len(jsonBytes)
	}
	fmt.Printf("%s\n...\n\n", string(jsonBytes[:jsonPreviewLen]))

	// 6. Roundtrip Inherent JSON Deserialization back into Document
	startDe := time.Now()
	var roundtripDoc pacs008.FiToFiCustomerCreditTransfer
	if err := json.Unmarshal(jsonBytes, &roundtripDoc); err != nil {
		fmt.Fprintf(os.Stderr, "JSON unmarshal error: %v\n", err)
		os.Exit(1)
	}
	deDuration := time.Since(startDe)

	fmt.Printf("[3] Inherent JSON Deserialization into Document (latency: %v):\n", deDuration)
	fmt.Printf("    Restored MsgId: %s\n", roundtripDoc.GrpHdr.MsgID)
	fmt.Printf("    Restored UETR:  %s\n", roundtripDoc.CdtTrfTxInf[0].PmtID.Uetr)
	fmt.Printf("    Restored Amount: %.2f %s\n", roundtripDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Value, roundtripDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency)
	fmt.Printf("    Restored Debtor: %s\n", roundtripDoc.CdtTrfTxInf[0].Dbtr.Name)
	fmt.Printf("    Restored Creditor: %s\n", roundtripDoc.CdtTrfTxInf[0].Cdtr.Name)

	if roundtripDoc.CdtTrfTxInf[0].PmtID.Uetr != intent.Payment.UETR {
		panic("UETR mismatch in Go JSON roundtrip")
	}

	fmt.Println("\n✅ Go Modern Payments ↔ ISO 20022 pacs.008 Bridge executed successfully!")
}

