import fs from "node:fs";
import path from "node:path";
import {
  ChargeBearerType,
  SettlementMethodCode,
  FiToFiCustomerCreditTransferSchema,
  type FiToFiCustomerCreditTransfer,
} from "../../generated/typescript/pacs_008_core.ts";

interface PaymentIntentJSON {
  message_id: string;
  created_at: string;
  settlement_method: string;
  clearing_network: string;
  payment: {
    instruction_id: string;
    end_to_end_id: string;
    transaction_id: string;
    uetr: string;
    amount: number;
    currency: string;
    settlement_date: string;
    charge_bearer: string;
    debtor: {
      name: string;
      address: {
        street_name?: string;
        building_number?: string;
        post_code?: string;
        town_name: string;
        country: string;
      };
      country_of_residence?: string;
    };
    debtor_account: {
      account_number: string;
      currency?: string;
      account_name?: string;
    };
    debtor_agent: {
      bicfi?: string;
      clearing_system_id: string;
      routing_number: string;
      name?: string;
    };
    creditor_agent: {
      bicfi?: string;
      clearing_system_id: string;
      routing_number: string;
      name?: string;
    };
    creditor: {
      name: string;
      address: {
        street_name?: string;
        building_number?: string;
        post_code?: string;
        town_name: string;
        country: string;
      };
      country_of_residence?: string;
    };
    creditor_account: {
      account_number: string;
      currency?: string;
      account_name?: string;
    };
    purpose_code?: string;
    remittance?: string;
  };
}

function findDataFile(): string {
  const candidates = [
    "data/payment_intent_fednow.json",
    "../../data/payment_intent_fednow.json",
    "../data/payment_intent_fednow.json",
  ];
  for (const c of candidates) {
    if (fs.existsSync(c)) {
      return path.resolve(c);
    }
  }
  throw new Error("Could not locate data/payment_intent_fednow.json");
}

function serializeXml(doc: FiToFiCustomerCreditTransfer): string {
  let xml = '<?xml version="1.0" encoding="UTF-8"?>\n';
  xml += '<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10">\n';
  xml += "  <GrpHdr>\n";
  xml += `    <MsgId>${doc.grpHdr.msgId}</MsgId>\n`;
  xml += `    <CreDtTm>${doc.grpHdr.creDtTm}</CreDtTm>\n`;
  xml += `    <NbOfTxs>${doc.grpHdr.nbOfTxs}</NbOfTxs>\n`;
  xml += "    <SttlmInf>\n";
  xml += `      <SettlementMethod>${doc.grpHdr.sttlmInf.settlementMethod}</SettlementMethod>\n`;
  if (doc.grpHdr.sttlmInf.clearingSystem) {
    xml += `      <ClearingSystem>${doc.grpHdr.sttlmInf.clearingSystem}</ClearingSystem>\n`;
  }
  xml += "    </SttlmInf>\n";
  xml += "  </GrpHdr>\n";

  for (const tx of doc.cdtTrfTxInf) {
    xml += "  <CdtTrfTxInf>\n";
    xml += "    <PmtId>\n";
    xml += `      <InstructionId>${tx.pmtId.instructionId}</InstructionId>\n`;
    xml += `      <EndToEndId>${tx.pmtId.endToEndId}</EndToEndId>\n`;
    xml += `      <TransactionId>${tx.pmtId.transactionId}</TransactionId>\n`;
    xml += `      <UETR>${tx.pmtId.uetr}</UETR>\n`;
    xml += "    </PmtId>\n";
    xml += `    <IntrBkSttlmAmt Currency="${tx.intrBkSttlmAmt.currency}">${tx.intrBkSttlmAmt.value.toFixed(2)}</IntrBkSttlmAmt>\n`;
    xml += `    <IntrBkSttlmDt>${tx.intrBkSttlmDt}</IntrBkSttlmDt>\n`;
    xml += `    <ChargeBearer>${tx.chargeBearer}</ChargeBearer>\n`;

    xml += "    <Dbtr>\n";
    xml += `      <Name>${tx.dbtr.name}</Name>\n`;
    if (tx.dbtr.postalAddress) {
      xml += "      <PostalAddress>\n";
      if (tx.dbtr.postalAddress.streetName) xml += `        <StreetName>${tx.dbtr.postalAddress.streetName}</StreetName>\n`;
      if (tx.dbtr.postalAddress.buildingNumber) xml += `        <BuildingNumber>${tx.dbtr.postalAddress.buildingNumber}</BuildingNumber>\n`;
      if (tx.dbtr.postalAddress.postCode) xml += `        <PostCode>${tx.dbtr.postalAddress.postCode}</PostCode>\n`;
      xml += `        <TownName>${tx.dbtr.postalAddress.townName}</TownName>\n`;
      xml += `        <Country>${tx.dbtr.postalAddress.country}</Country>\n`;
      xml += "      </PostalAddress>\n";
    }
    if (tx.dbtr.countryOfResidence) {
      xml += `      <CountryOfResidence>${tx.dbtr.countryOfResidence}</CountryOfResidence>\n`;
    }
    xml += "    </Dbtr>\n";

    xml += "    <DbtrAcct>\n";
    xml += "      <Id>\n";
    if (tx.dbtrAcct.id.proprietaryAccount) {
      xml += `        <ProprietaryAccount>${tx.dbtrAcct.id.proprietaryAccount}</ProprietaryAccount>\n`;
    }
    xml += "      </Id>\n";
    if (tx.dbtrAcct.currency) xml += `      <Currency>${tx.dbtrAcct.currency}</Currency>\n`;
    if (tx.dbtrAcct.name) xml += `      <Name>${tx.dbtrAcct.name}</Name>\n`;
    xml += "    </DbtrAcct>\n";

    xml += "    <DbtrAgt>\n";
    xml += "      <FinInstnId>\n";
    if (tx.dbtrAgt.finInstnId.bicfi) xml += `        <BICFI>${tx.dbtrAgt.finInstnId.bicfi}</BICFI>\n`;
    if (tx.dbtrAgt.finInstnId.clearingSystemMemberId) {
      xml += "        <ClearingSystemMemberId>\n";
      xml += `          <ClearingSystemId>${tx.dbtrAgt.finInstnId.clearingSystemMemberId.clearingSystemId}</ClearingSystemId>\n`;
      xml += `          <MemberId>${tx.dbtrAgt.finInstnId.clearingSystemMemberId.memberId}</MemberId>\n`;
      xml += "        </ClearingSystemMemberId>\n";
    }
    if (tx.dbtrAgt.finInstnId.name) xml += `        <Name>${tx.dbtrAgt.finInstnId.name}</Name>\n`;
    xml += "      </FinInstnId>\n";
    xml += "    </DbtrAgt>\n";

    xml += "    <CdtrAgt>\n";
    xml += "      <FinInstnId>\n";
    if (tx.cdtrAgt.finInstnId.bicfi) xml += `        <BICFI>${tx.cdtrAgt.finInstnId.bicfi}</BICFI>\n`;
    if (tx.cdtrAgt.finInstnId.clearingSystemMemberId) {
      xml += "        <ClearingSystemMemberId>\n";
      xml += `          <ClearingSystemId>${tx.cdtrAgt.finInstnId.clearingSystemMemberId.clearingSystemId}</ClearingSystemId>\n`;
      xml += `          <MemberId>${tx.cdtrAgt.finInstnId.clearingSystemMemberId.memberId}</MemberId>\n`;
      xml += "        </ClearingSystemMemberId>\n";
    }
    if (tx.cdtrAgt.finInstnId.name) xml += `        <Name>${tx.cdtrAgt.finInstnId.name}</Name>\n`;
    xml += "      </FinInstnId>\n";
    xml += "    </CdtrAgt>\n";

    xml += "    <Cdtr>\n";
    xml += `      <Name>${tx.cdtr.name}</Name>\n`;
    if (tx.cdtr.postalAddress) {
      xml += "      <PostalAddress>\n";
      if (tx.cdtr.postalAddress.streetName) xml += `        <StreetName>${tx.cdtr.postalAddress.streetName}</StreetName>\n`;
      if (tx.cdtr.postalAddress.buildingNumber) xml += `        <BuildingNumber>${tx.cdtr.postalAddress.buildingNumber}</BuildingNumber>\n`;
      if (tx.cdtr.postalAddress.postCode) xml += `        <PostCode>${tx.cdtr.postalAddress.postCode}</PostCode>\n`;
      xml += `        <TownName>${tx.cdtr.postalAddress.townName}</TownName>\n`;
      xml += `        <Country>${tx.cdtr.postalAddress.country}</Country>\n`;
      xml += "      </PostalAddress>\n";
    }
    if (tx.cdtr.countryOfResidence) {
      xml += `      <CountryOfResidence>${tx.cdtr.countryOfResidence}</CountryOfResidence>\n`;
    }
    xml += "    </Cdtr>\n";

    xml += "    <CdtrAcct>\n";
    xml += "      <Id>\n";
    if (tx.cdtrAcct.id.proprietaryAccount) {
      xml += `        <ProprietaryAccount>${tx.cdtrAcct.id.proprietaryAccount}</ProprietaryAccount>\n`;
    }
    xml += "      </Id>\n";
    if (tx.cdtrAcct.currency) xml += `      <Currency>${tx.cdtrAcct.currency}</Currency>\n`;
    if (tx.cdtrAcct.name) xml += `      <Name>${tx.cdtrAcct.name}</Name>\n`;
    xml += "    </CdtrAcct>\n";

    if (tx.purpose) xml += `    <Purpose>${tx.purpose}</Purpose>\n`;
    if (tx.rmtInf?.unstructured) {
      xml += "    <RmtInf>\n";
      xml += `      <Unstructured>${tx.rmtInf.unstructured}</Unstructured>\n`;
      xml += "    </RmtInf>\n";
    }
    xml += "  </CdtTrfTxInf>\n";
  }
  xml += "</Document>";
  return xml;
}

function main() {
  console.log("================================================================================");
  console.log("  PolyXML TypeScript 5+ Showcase: FedNow Payment Intent -> ISO 20022 pacs.008");
  console.log("================================================================================");

  const dataPath = findDataFile();
  const raw = fs.readFileSync(dataPath, "utf-8");
  const intent: PaymentIntentJSON = JSON.parse(raw);

  // 1. Map to PolyXML-typed ISO 20022 object
  const candidateDoc: FiToFiCustomerCreditTransfer = {
    grpHdr: {
      msgId: intent.message_id,
      creDtTm: intent.created_at,
      nbOfTxs: 1,
      sttlmInf: {
        settlementMethod: (intent.settlement_method as any) || SettlementMethodCode.Clrg,
        clearingSystem: intent.clearing_network,
      },
    },
    cdtTrfTxInf: [
      {
        pmtId: {
          instructionId: intent.payment.instruction_id,
          endToEndId: intent.payment.end_to_end_id,
          transactionId: intent.payment.transaction_id,
          uetr: intent.payment.uetr,
        },
        intrBkSttlmAmt: {
          currency: intent.payment.currency,
          value: intent.payment.amount,
        },
        intrBkSttlmDt: intent.payment.settlement_date,
        chargeBearer: (intent.payment.charge_bearer as any) || ChargeBearerType.Slev,
        dbtr: {
          name: intent.payment.debtor.name,
          postalAddress: {
            streetName: intent.payment.debtor.address.street_name,
            buildingNumber: intent.payment.debtor.address.building_number,
            postCode: intent.payment.debtor.address.post_code,
            townName: intent.payment.debtor.address.town_name,
            country: intent.payment.debtor.address.country,
          },
          countryOfResidence: intent.payment.debtor.country_of_residence,
        },
        dbtrAcct: {
          id: {
            proprietaryAccount: intent.payment.debtor_account.account_number,
          },
          currency: intent.payment.debtor_account.currency,
          name: intent.payment.debtor_account.account_name,
        },
        dbtrAgt: {
          finInstnId: {
            bicfi: intent.payment.debtor_agent.bicfi,
            clearingSystemMemberId: {
              clearingSystemId: intent.payment.debtor_agent.clearing_system_id,
              memberId: intent.payment.debtor_agent.routing_number,
            },
            name: intent.payment.debtor_agent.name,
          },
        },
        cdtrAgt: {
          finInstnId: {
            bicfi: intent.payment.creditor_agent.bicfi,
            clearingSystemMemberId: {
              clearingSystemId: intent.payment.creditor_agent.clearing_system_id,
              memberId: intent.payment.creditor_agent.routing_number,
            },
            name: intent.payment.creditor_agent.name,
          },
        },
        cdtr: {
          name: intent.payment.creditor.name,
          postalAddress: {
            streetName: intent.payment.creditor.address.street_name,
            buildingNumber: intent.payment.creditor.address.building_number,
            postCode: intent.payment.creditor.address.post_code,
            townName: intent.payment.creditor.address.town_name,
            country: intent.payment.creditor.address.country,
          },
          countryOfResidence: intent.payment.creditor.country_of_residence,
        },
        cdtrAcct: {
          id: {
            proprietaryAccount: intent.payment.creditor_account.account_number,
          },
          currency: intent.payment.creditor_account.currency,
          name: intent.payment.creditor_account.account_name,
        },
        purpose: intent.payment.purpose_code,
        rmtInf: intent.payment.remittance ? { unstructured: intent.payment.remittance } : undefined,
      },
    ],
  };

  // 2. Validate with Zod
  const validatedDoc = FiToFiCustomerCreditTransferSchema.parse(candidateDoc);
  console.log("✔ ISO 20022 pacs.008 Zod runtime validation passed!");

  // 3. Serialize to XML and JSON
  const xmlOutput = serializeXml(validatedDoc);
  const jsonWire = JSON.stringify(validatedDoc, null, 2);

  console.log(`\nGenerated XML Payload (size: ${xmlOutput.length} bytes):`);
  const lines = xmlOutput.split("\n");
  for (let i = 0; i < Math.min(22, lines.length); i++) {
    console.log("  " + lines[i]);
  }
  console.log("  ... [truncated]");

  // 4. Benchmarking
  const iterations = 10000;

  const t1 = performance.now();
  for (let i = 0; i < iterations; i++) {
    serializeXml(validatedDoc);
  }
  const t2 = performance.now();
  const xmlUs = ((t2 - t1) / iterations) * 1000.0;

  const t3 = performance.now();
  for (let i = 0; i < iterations; i++) {
    JSON.stringify(validatedDoc);
  }
  const t4 = performance.now();
  const jsonUs = ((t4 - t3) / iterations) * 1000.0;

  console.log("\n--------------------------------------------------------------------------------");
  console.log("  PolyXML TypeScript Performance Metrics (10,000 iterations)");
  console.log("--------------------------------------------------------------------------------");
  console.log(`  XML Serialization:       ${xmlUs.toFixed(2).padStart(8)} µs/op`);
  console.log(`  JSON Serialization:      ${jsonUs.toFixed(2).padStart(8)} µs/op`);
  console.log("--------------------------------------------------------------------------------");
  console.log(`  Debtor:     ${validatedDoc.cdtTrfTxInf[0].dbtr.name} ($${validatedDoc.cdtTrfTxInf[0].intrBkSttlmAmt.value.toFixed(2)} ${validatedDoc.cdtTrfTxInf[0].intrBkSttlmAmt.currency})`);
  console.log(`  Creditor:   ${validatedDoc.cdtTrfTxInf[0].cdtr.name} via ${validatedDoc.cdtTrfTxInf[0].cdtrAgt.finInstnId.name}`);
  console.log("================================================================================");
}

main();
