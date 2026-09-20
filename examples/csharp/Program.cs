using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Xml;
using System.Xml.Serialization;
using Financial.Iso20022.Pacs008;

namespace FedNowPacs008Adapter;

public record PaymentIntentDTO(
    [property: JsonPropertyName("message_id")] string MessageId,
    [property: JsonPropertyName("created_at")] string CreatedAt,
    [property: JsonPropertyName("settlement_method")] string SettlementMethod,
    [property: JsonPropertyName("clearing_network")] string ClearingNetwork,
    [property: JsonPropertyName("payment")] PaymentDetailDTO Payment
);

public record PaymentDetailDTO(
    [property: JsonPropertyName("instruction_id")] string InstructionId,
    [property: JsonPropertyName("end_to_end_id")] string EndToEndId,
    [property: JsonPropertyName("transaction_id")] string TransactionId,
    [property: JsonPropertyName("uetr")] string Uetr,
    [property: JsonPropertyName("amount")] double Amount,
    [property: JsonPropertyName("currency")] string Currency,
    [property: JsonPropertyName("settlement_date")] string SettlementDate,
    [property: JsonPropertyName("charge_bearer")] string ChargeBearer,
    [property: JsonPropertyName("debtor")] PartyDTO Debtor,
    [property: JsonPropertyName("debtor_account")] AccountDTO DebtorAccount,
    [property: JsonPropertyName("debtor_agent")] AgentDTO DebtorAgent,
    [property: JsonPropertyName("creditor_agent")] AgentDTO CreditorAgent,
    [property: JsonPropertyName("creditor")] PartyDTO Creditor,
    [property: JsonPropertyName("creditor_account")] AccountDTO CreditorAccount,
    [property: JsonPropertyName("purpose_code")] string PurposeCode,
    [property: JsonPropertyName("remittance")] string Remittance
);

public record PartyDTO(
    [property: JsonPropertyName("name")] string Name,
    [property: JsonPropertyName("address")] AddressDTO Address,
    [property: JsonPropertyName("country_of_residence")] string CountryOfResidence
);

public record AddressDTO(
    [property: JsonPropertyName("street_name")] string StreetName,
    [property: JsonPropertyName("building_number")] string BuildingNumber,
    [property: JsonPropertyName("post_code")] string PostCode,
    [property: JsonPropertyName("town_name")] string TownName,
    [property: JsonPropertyName("country")] string Country
);

public record AccountDTO(
    [property: JsonPropertyName("account_number")] string AccountNumber,
    [property: JsonPropertyName("currency")] string Currency,
    [property: JsonPropertyName("account_name")] string AccountName
);

public record AgentDTO(
    [property: JsonPropertyName("bicfi")] string Bicfi,
    [property: JsonPropertyName("clearing_system_id")] string ClearingSystemId,
    [property: JsonPropertyName("routing_number")] string RoutingNumber,
    [property: JsonPropertyName("name")] string Name
);

public static class Program
{
    private static string FindDataFile()
    {
        string[] candidates = {
            "data/payment_intent_fednow.json",
            "../../data/payment_intent_fednow.json",
            "../data/payment_intent_fednow.json"
        };
        foreach (var c in candidates)
        {
            if (File.Exists(c))
            {
                return Path.GetFullPath(c);
            }
        }
        throw new FileNotFoundException("Could not locate data/payment_intent_fednow.json");
    }

    public static void Main(string[] args)
    {
        Console.WriteLine("================================================================================");
        Console.WriteLine("  PolyXML C# 12 / .NET 8 Showcase: FedNow Payment Intent -> ISO 20022 pacs.008");
        Console.WriteLine("================================================================================");

        var dataPath = FindDataFile();
        var jsonBytes = File.ReadAllBytes(dataPath);
        var intent = JsonSerializer.Deserialize<PaymentIntentDTO>(jsonBytes)!;

        // 1. Map to PolyXML-generated C# 12 records
        var creDtTm = DateTimeOffset.Parse(intent.CreatedAt);
        var sttlmDt = DateOnly.Parse(intent.Payment.SettlementDate);

        var doc = new Document
        {
            GrpHdr = new GroupHeader(
                intent.MessageId,
                creDtTm,
                1,
                new SettlementInstruction(
                    Enum.Parse<SettlementMethodCode>(intent.SettlementMethod, true),
                    intent.ClearingNetwork
                )
            ),
            CdtTrfTxInf = new List<CreditTransferTransactionInformation>
            {
                new(
                    new PaymentIdentification(
                        intent.Payment.InstructionId,
                        intent.Payment.EndToEndId,
                        intent.Payment.TransactionId,
                        intent.Payment.Uetr
                    ),
                    new ActiveOrHistoricCurrencyAndAmount(
                        new ActiveCurrencyCode(intent.Payment.Currency),
                        intent.Payment.Amount
                    ),
                    sttlmDt,
                    Enum.Parse<ChargeBearerType>(intent.Payment.ChargeBearer, true),
                    new PartyIdentification(
                        intent.Payment.Debtor.Name,
                        new PostalAddress(
                            intent.Payment.Debtor.Address.StreetName,
                            intent.Payment.Debtor.Address.BuildingNumber,
                            intent.Payment.Debtor.Address.PostCode,
                            intent.Payment.Debtor.Address.TownName,
                            intent.Payment.Debtor.Address.Country
                        ),
                        intent.Payment.Debtor.CountryOfResidence
                    ),
                    new CashAccount(
                        new AccountIdentification(null, intent.Payment.DebtorAccount.AccountNumber),
                        new ActiveCurrencyCode(intent.Payment.DebtorAccount.Currency),
                        intent.Payment.DebtorAccount.AccountName
                    ),
                    new BranchAndFinancialInstitutionIdentification(
                        new FinancialInstitutionIdentification(
                            intent.Payment.DebtorAgent.Bicfi,
                            new ClearingSystemMemberIdentification(
                                intent.Payment.DebtorAgent.ClearingSystemId,
                                intent.Payment.DebtorAgent.RoutingNumber
                            ),
                            intent.Payment.DebtorAgent.Name
                        )
                    ),
                    new BranchAndFinancialInstitutionIdentification(
                        new FinancialInstitutionIdentification(
                            intent.Payment.CreditorAgent.Bicfi,
                            new ClearingSystemMemberIdentification(
                                intent.Payment.CreditorAgent.ClearingSystemId,
                                intent.Payment.CreditorAgent.RoutingNumber
                            ),
                            intent.Payment.CreditorAgent.Name
                        )
                    ),
                    new PartyIdentification(
                        intent.Payment.Creditor.Name,
                        new PostalAddress(
                            intent.Payment.Creditor.Address.StreetName,
                            intent.Payment.Creditor.Address.BuildingNumber,
                            intent.Payment.Creditor.Address.PostCode,
                            intent.Payment.Creditor.Address.TownName,
                            intent.Payment.Creditor.Address.Country
                        ),
                        intent.Payment.Creditor.CountryOfResidence
                    ),
                    new CashAccount(
                        new AccountIdentification(null, intent.Payment.CreditorAccount.AccountNumber),
                        new ActiveCurrencyCode(intent.Payment.CreditorAccount.Currency),
                        intent.Payment.CreditorAccount.AccountName
                    ),
                    intent.Payment.PurposeCode,
                    new RemittanceInformation(intent.Payment.Remittance)
                )
            }
        };

        Console.WriteLine("✔ ISO 20022 pacs.008 C# 12 Record model instantiated!");

        // 2. XML Serialization via XmlSerializer
        var xmlSerializer = new XmlSerializer(typeof(Document), "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10");
        var xmlSettings = new XmlWriterSettings
        {
            Indent = true,
            OmitXmlDeclaration = false
        };

        using var stringWriter = new StringWriter();
        using (var xmlWriter = XmlWriter.Create(stringWriter, xmlSettings))
        {
            var namespaces = new XmlSerializerNamespaces();
            namespaces.Add("", "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10");
            xmlSerializer.Serialize(xmlWriter, doc, namespaces);
        }
        var xmlOutput = stringWriter.ToString();

        // 3. JSON Serialization via System.Text.Json
        var jsonOptions = new JsonSerializerOptions
        {
            WriteIndented = true
        };
        var jsonOutput = JsonSerializer.Serialize(doc, jsonOptions);

        // 4. JSON Deserialization roundtrip
        var restoredDoc = JsonSerializer.Deserialize<Document>(jsonOutput, jsonOptions)!;

        Console.WriteLine($"\nGenerated XML Payload (size: {xmlOutput.Length} bytes):");
        var xmlLines = xmlOutput.Split('\n');
        for (int i = 0; i < Math.Min(22, xmlLines.Length); i++)
        {
            Console.WriteLine("  " + xmlLines[i]);
        }
        Console.WriteLine("  ... [truncated]");

        Console.WriteLine($"\nGenerated JSON Wire Representation (size: {jsonOutput.Length} bytes):");
        var jsonLines = jsonOutput.Split('\n');
        for (int i = 0; i < Math.Min(18, jsonLines.Length); i++)
        {
            Console.WriteLine("  " + jsonLines[i]);
        }
        Console.WriteLine("  ... [truncated]");

        // 5. Benchmarking
        const int iterations = 10000;

        var swXml = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            using var sw = new StringWriter();
            using var xw = XmlWriter.Create(sw, xmlSettings);
            var ns = new XmlSerializerNamespaces();
            ns.Add("", "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10");
            xmlSerializer.Serialize(xw, doc, ns);
        }
        swXml.Stop();
        var xmlUs = (double)swXml.Elapsed.TotalMicroseconds / iterations;

        var swJson = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            _ = JsonSerializer.Serialize(doc, jsonOptions);
        }
        swJson.Stop();
        var jsonUs = (double)swJson.Elapsed.TotalMicroseconds / iterations;

        var swDe = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            _ = JsonSerializer.Deserialize<Document>(jsonOutput, jsonOptions);
        }
        swDe.Stop();
        var deUs = (double)swDe.Elapsed.TotalMicroseconds / iterations;

        Console.WriteLine("\n--------------------------------------------------------------------------------");
        Console.WriteLine("  PolyXML C# 12 / .NET 8 Performance Metrics (10,000 iterations)");
        Console.WriteLine("--------------------------------------------------------------------------------");
        Console.WriteLine($"  XML Serialization:       {xmlUs,8:F2} µs/op");
        Console.WriteLine($"  JSON Serialization:      {jsonUs,8:F2} µs/op");
        Console.WriteLine($"  JSON Deserialization:    {deUs,8:F2} µs/op");
        Console.WriteLine("--------------------------------------------------------------------------------");
        Console.WriteLine($"  Debtor:     {restoredDoc.CdtTrfTxInf[0].Dbtr.Name} (${restoredDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Value:F2} {restoredDoc.CdtTrfTxInf[0].IntrBkSttlmAmt.Currency.Value})");
        Console.WriteLine($"  Creditor:   {restoredDoc.CdtTrfTxInf[0].Cdtr.Name} via {restoredDoc.CdtTrfTxInf[0].CdtrAgt.FinInstnId.Name}");
        Console.WriteLine("================================================================================");
    }
}
