#!/usr/bin/env python3
"""
PolyXML Finance Showcase: Modern FinTech Payments ↔ ISO 20022 pacs.008 Bridge (Python)
Translates modern payment intent JSON into ISO 20022 pacs.008.001.10 XML & JSON.
"""

from __future__ import annotations

import json
import sys
import time
from pathlib import Path

# Add generated directory to path
repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "generated" / "python"))

import polyxml
from pacs_008_core import (
    AccountIdentification,
    ActiveOrHistoricCurrencyAndAmount,
    BranchAndFinancialInstitutionIdentification,
    CashAccount,
    ChargeBearerType,
    ClearingSystemMemberIdentification,
    CreditTransferTransactionInformation,
    Document,
    FiToFiCustomerCreditTransfer,
    FinancialInstitutionIdentification,
    GroupHeader,
    PartyIdentification,
    PaymentIdentification,
    PostalAddress,
    RemittanceInformation,
    SettlementInstruction,
    SettlementMethodCode,
)


def main():
    print("=" * 80)
    print("💳 PolyXML: Modern FinTech Payments ↔ ISO 20022 pacs.008 Bridge (Python)")
    print("=" * 80)

    data_file = repo_root / "data" / "payment_intent_fednow.json"
    with open(data_file, "r", encoding="utf-8") as f:
        intent = json.load(f)

    pmt = intent["payment"]
    print(f"Ingesting Instant Payment Intent: UETR {pmt['uetr']} (Amount: {pmt['amount']:.2f} {pmt['currency']})\n")

    # Construct ISO 20022 Document dataclass hierarchy
    t0 = time.perf_counter_ns()
    doc = FiToFiCustomerCreditTransfer(
        grp_hdr=GroupHeader(
            msg_id=intent["message_id"],
            cre_dt_tm=intent["created_at"],
            nb_of_txs=1,
            sttlm_inf=SettlementInstruction(
                settlement_method=SettlementMethodCode.CLRG,
                clearing_system=intent.get("clearing_network", "FEDNOW_INSTANT"),
            ),
        ),
        cdt_trf_tx_inf=[
            CreditTransferTransactionInformation(
                pmt_id=PaymentIdentification(
                    instruction_id=pmt["instruction_id"],
                    end_to_end_id=pmt["end_to_end_id"],
                    transaction_id=pmt["transaction_id"],
                    uetr=pmt["uetr"],
                ),
                intr_bk_sttlm_amt=ActiveOrHistoricCurrencyAndAmount(
                    currency=pmt["currency"],
                    value=float(pmt["amount"]),
                ),
                intr_bk_sttlm_dt=pmt["settlement_date"],
                charge_bearer=ChargeBearerType.SLEV,
                dbtr=PartyIdentification(
                    name=pmt["debtor"]["name"],
                    postal_address=PostalAddress(
                        street_name=pmt["debtor"]["address"]["street_name"],
                        building_number=pmt["debtor"]["address"]["building_number"],
                        post_code=pmt["debtor"]["address"]["post_code"],
                        town_name=pmt["debtor"]["address"]["town_name"],
                        country=pmt["debtor"]["address"]["country"],
                    ),
                    country_of_residence=pmt["debtor"].get("country_of_residence", "US"),
                ),
                dbtr_acct=CashAccount(
                    id=AccountIdentification(
                        proprietary_account=pmt["debtor_account"]["account_number"],
                    ),
                    currency=pmt["debtor_account"].get("currency", "USD"),
                    name=pmt["debtor_account"].get("account_name", "Acme Treasury USD Account"),
                ),
                dbtr_agt=BranchAndFinancialInstitutionIdentification(
                    fin_instn_id=FinancialInstitutionIdentification(
                        bicfi=pmt["debtor_agent"]["bicfi"],
                        clearing_system_member_id=ClearingSystemMemberIdentification(
                            clearing_system_id=pmt["debtor_agent"]["clearing_system_id"],
                            member_id=pmt["debtor_agent"]["routing_number"],
                        ),
                        name=pmt["debtor_agent"]["name"],
                    ),
                ),
                cdtr_agt=BranchAndFinancialInstitutionIdentification(
                    fin_instn_id=FinancialInstitutionIdentification(
                        bicfi=pmt["creditor_agent"]["bicfi"],
                        clearing_system_member_id=ClearingSystemMemberIdentification(
                            clearing_system_id=pmt["creditor_agent"]["clearing_system_id"],
                            member_id=pmt["creditor_agent"]["routing_number"],
                        ),
                        name=pmt["creditor_agent"]["name"],
                    ),
                ),
                cdtr=PartyIdentification(
                    name=pmt["creditor"]["name"],
                    postal_address=PostalAddress(
                        street_name=pmt["creditor"]["address"]["street_name"],
                        building_number=pmt["creditor"]["address"]["building_number"],
                        post_code=pmt["creditor"]["address"]["post_code"],
                        town_name=pmt["creditor"]["address"]["town_name"],
                        country=pmt["creditor"]["address"]["country"],
                    ),
                    country_of_residence=pmt["creditor"].get("country_of_residence", "US"),
                ),
                cdtr_acct=CashAccount(
                    id=AccountIdentification(
                        proprietary_account=pmt["creditor_account"]["account_number"],
                    ),
                    currency=pmt["creditor_account"].get("currency", "USD"),
                    name=pmt["creditor_account"].get("account_name", "Horizon Fleet Operations Settlement"),
                ),
                purpose=pmt.get("purpose_code", "SUPP"),
                rmt_inf=RemittanceInformation(unstructured=pmt.get("remittance")),
            )
        ],
    )
    t_construct = (time.perf_counter_ns() - t0) / 1000.0

    # 1. Inherent XML Serialization via PolyXML native engine
    t1 = time.perf_counter_ns()
    xml_bytes = doc.to_xml(indent=2)
    t_xml = (time.perf_counter_ns() - t1) / 1000.0
    xml_str = xml_bytes.decode("utf-8")

    print(f"[1] Generated ISO 20022 pacs.008.001.10 XML Message (latency: {t_xml:.2f}µs):")
    print(xml_str.strip()[:400] + "\n...\n")

    # 2. Inherent Native JSON Serialization on same model
    t2 = time.perf_counter_ns()
    json_bytes = doc.to_json(indent=2)
    t_json = (time.perf_counter_ns() - t2) / 1000.0
    json_str = json_bytes.decode("utf-8")

    print(f"[2] Generated Native JSON on Same Model (latency: {t_json:.2f}µs):")
    print(json_str.strip()[:400] + "\n...\n")

    # 3. Roundtrip Inherent JSON Deserialization back into Document
    t3 = time.perf_counter_ns()
    restored_doc = FiToFiCustomerCreditTransfer.from_json(json_bytes)
    t_restore = (time.perf_counter_ns() - t3) / 1000.0

    tx = restored_doc.cdt_trf_tx_inf[0]
    print(f"[3] Inherent JSON Deserialization into Document (latency: {t_restore:.2f}µs):")
    print(f"    Restored MsgId: {restored_doc.grp_hdr.msg_id}")
    print(f"    Restored UETR:  {tx.pmt_id.uetr}")
    print(f"    Restored Amount: {tx.intr_bk_sttlm_amt.value:.2f} {tx.intr_bk_sttlm_amt.currency}")
    print(f"    Restored Debtor: {tx.dbtr.name}")
    print(f"    Restored Creditor: {tx.cdtr.name}")

    # 4. PolyXML C-Engine Schemaless XML->JSON Transcoder Demo
    t4 = time.perf_counter_ns()
    transcoded_json = polyxml.xml_to_json(xml_bytes, indent=2)
    t_transcode = (time.perf_counter_ns() - t4) / 1000.0
    print(f"\n[4] PolyXML C-Engine Schemaless XML->JSON Transcoder (latency: {t_transcode:.2f}µs)")
    print(f"    Transcoded payload bytes: {len(transcoded_json)}")

    print("\n✅ Python Modern Payments ↔ ISO 20022 pacs.008 Bridge executed successfully!")


if __name__ == "__main__":
    main()
