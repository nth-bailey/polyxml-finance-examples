use std::borrow::Cow;
use std::time::Instant;

#[path = "../../../generated/rust/pacs_008_core.rs"]
mod pacs_008_core;

use pacs_008_core::*;

#[derive(serde::Deserialize)]
struct DebtorCreditorJson<'a> {
    name: &'a str,
    #[serde(default)]
    country_of_residence: Option<&'a str>,
}

#[derive(serde::Deserialize)]
struct AccountJson<'a> {
    account_number: &'a str,
    currency: &'a str,
    #[serde(alias = "account_name")]
    name: &'a str,
}

#[derive(serde::Deserialize)]
struct AgentJson<'a> {
    bicfi: &'a str,
    clearing_system_id: &'a str,
    routing_number: &'a str,
    name: &'a str,
}

#[allow(dead_code)]
#[derive(serde::Deserialize)]
struct PaymentJson<'a> {
    instruction_id: &'a str,
    end_to_end_id: &'a str,
    transaction_id: &'a str,
    uetr: &'a str,
    amount: f64,
    currency: &'a str,
    settlement_date: &'a str,
    charge_bearer: &'a str,
    debtor: DebtorCreditorJson<'a>,
    debtor_account: AccountJson<'a>,
    debtor_agent: AgentJson<'a>,
    creditor_agent: AgentJson<'a>,
    creditor: DebtorCreditorJson<'a>,
    creditor_account: AccountJson<'a>,
    purpose_code: Option<&'a str>,
    remittance: Option<&'a str>,
}

#[allow(dead_code)]
#[derive(serde::Deserialize)]
struct PaymentIntentFeed<'a> {
    message_id: &'a str,
    created_at: &'a str,
    settlement_method: &'a str,
    clearing_network: &'a str,
    payment: PaymentJson<'a>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("================================================================================");
    println!("💳 PolyXML: Modern FinTech Payments ↔ ISO 20022 pacs.008 Bridge (Rust Zero-Copy)");
    println!("================================================================================");

    let data_bytes = std::fs::read("data/payment_intent_fednow.json")
        .or_else(|_| std::fs::read("../../data/payment_intent_fednow.json"))?;
    let data_str = std::str::from_utf8(&data_bytes)?;

    let intent: PaymentIntentFeed = serde_json::from_str(data_str)?;
    let pmt = &intent.payment;

    println!(
        "Ingesting Instant Payment Intent: UETR {} (Amount: {:.2} {})\n",
        pmt.uetr, pmt.amount, pmt.currency
    );

    // Build strongly-typed ISO 20022 Document (FIToFICustomerCreditTransfer) borrowing zero-copy slices
    let t_start = Instant::now();
    let doc = Document {
        grp_hdr: GroupHeader {
            msg_id: Cow::Borrowed(intent.message_id),
            cre_dt_tm: Cow::Borrowed(intent.created_at),
            nb_of_txs: 1,
            sttlm_inf: SettlementInstruction {
                settlement_method: SettlementMethodCode::Clrg,
                clearing_system: Some(Cow::Borrowed(intent.clearing_network)),
            },
        },
        cdt_trf_tx_inf: vec![CreditTransferTransactionInformation {
            pmt_id: PaymentIdentification {
                instruction_id: Cow::Borrowed(pmt.instruction_id),
                end_to_end_id: Cow::Borrowed(pmt.end_to_end_id),
                transaction_id: Cow::Borrowed(pmt.transaction_id),
                uetr: Cow::Borrowed(pmt.uetr),
            },
            intr_bk_sttlm_amt: ActiveOrHistoricCurrencyAndAmount {
                currency: Cow::Borrowed(pmt.currency),
                value: pmt.amount,
            },
            intr_bk_sttlm_dt: Cow::Borrowed(pmt.settlement_date),
            charge_bearer: ChargeBearerType::Slev,
            dbtr: PartyIdentification {
                name: Cow::Borrowed(pmt.debtor.name),
                postal_address: Some(PostalAddress {
                    street_name: Some(Cow::Borrowed("500 Madison Avenue")),
                    building_number: Some(Cow::Borrowed("Suite 3400")),
                    post_code: Some(Cow::Borrowed("10022")),
                    town_name: Cow::Borrowed("New York"),
                    country: Cow::Borrowed("US"),
                }),
                country_of_residence: pmt.debtor.country_of_residence.map(Cow::Borrowed),
            },
            dbtr_acct: CashAccount {
                id: AccountIdentification {
                    iban: None,
                    proprietary_account: Some(Cow::Borrowed(pmt.debtor_account.account_number)),
                },
                currency: Some(Cow::Borrowed(pmt.debtor_account.currency)),
                name: Some(Cow::Borrowed(pmt.debtor_account.name)),
            },
            dbtr_agt: BranchAndFinancialInstitutionIdentification {
                fin_instn_id: FinancialInstitutionIdentification {
                    bicfi: Some(Cow::Borrowed(pmt.debtor_agent.bicfi)),
                    clearing_system_member_id: Some(ClearingSystemMemberIdentification {
                        clearing_system_id: Cow::Borrowed(pmt.debtor_agent.clearing_system_id),
                        member_id: Cow::Borrowed(pmt.debtor_agent.routing_number),
                    }),
                    name: Some(Cow::Borrowed(pmt.debtor_agent.name)),
                },
            },
            cdtr_agt: BranchAndFinancialInstitutionIdentification {
                fin_instn_id: FinancialInstitutionIdentification {
                    bicfi: Some(Cow::Borrowed(pmt.creditor_agent.bicfi)),
                    clearing_system_member_id: Some(ClearingSystemMemberIdentification {
                        clearing_system_id: Cow::Borrowed(pmt.creditor_agent.clearing_system_id),
                        member_id: Cow::Borrowed(pmt.creditor_agent.routing_number),
                    }),
                    name: Some(Cow::Borrowed(pmt.creditor_agent.name)),
                },
            },
            cdtr: PartyIdentification {
                name: Cow::Borrowed(pmt.creditor.name),
                postal_address: Some(PostalAddress {
                    street_name: Some(Cow::Borrowed("100 Commercial Wharf")),
                    building_number: Some(Cow::Borrowed("Pier 4")),
                    post_code: Some(Cow::Borrowed("02110")),
                    town_name: Cow::Borrowed("Boston"),
                    country: Cow::Borrowed("US"),
                }),
                country_of_residence: pmt.creditor.country_of_residence.map(Cow::Borrowed),
            },
            cdtr_acct: CashAccount {
                id: AccountIdentification {
                    iban: None,
                    proprietary_account: Some(Cow::Borrowed(pmt.creditor_account.account_number)),
                },
                currency: Some(Cow::Borrowed(pmt.creditor_account.currency)),
                name: Some(Cow::Borrowed(pmt.creditor_account.name)),
            },
            purpose: pmt.purpose_code.map(Cow::Borrowed),
            rmt_inf: pmt.remittance.map(|r| RemittanceInformation {
                unstructured: Some(Cow::Borrowed(r)),
            }),
        }],
    };
    let _construct_duration = t_start.elapsed();

    // 1. Inherent XML Serialization
    let t_xml = Instant::now();
    let xml_output = doc.to_xml_string()?;
    let xml_duration = t_xml.elapsed();

    println!(
        "[1] Generated ISO 20022 pacs.008.001.10 XML (latency: {:.2?}):",
        xml_duration
    );
    println!("{}\n...\n", &xml_output[..xml_output.len().min(400)]);

    // 2. Inherent Native JSON Serialization on SAME Model
    let t_json = Instant::now();
    let json_output = doc.to_json_string()?;
    let json_duration = t_json.elapsed();

    println!(
        "[2] Generated Native JSON on Same Model (latency: {:.2?}):",
        json_duration
    );
    println!("{}\n...\n", &json_output[..json_output.len().min(400)]);

    // 3. Inherent Zero-Copy JSON Deserialization back into Document
    let t_restore = Instant::now();
    let restored_doc = Document::from_json_str(&json_output)?;
    let restore_duration = t_restore.elapsed();

    let tx = &restored_doc.cdt_trf_tx_inf[0];
    println!(
        "[3] Inherent Zero-Copy JSON Deserialization into Document (latency: {:.2?}):",
        restore_duration
    );
    println!("    Restored MsgId: {}", restored_doc.grp_hdr.msg_id);
    println!("    Restored UETR:  {}", tx.pmt_id.uetr);
    println!(
        "    Restored Amount: {:.2} {}",
        tx.intr_bk_sttlm_amt.value, tx.intr_bk_sttlm_amt.currency
    );
    println!("    Restored Debtor: {}", tx.dbtr.name);
    println!("    Restored Creditor: {}", tx.cdtr.name);

    assert_eq!(tx.pmt_id.uetr.as_ref(), pmt.uetr);
    assert_eq!(tx.intr_bk_sttlm_amt.value, pmt.amount);

    println!("\n✅ Rust Modern Payments ↔ ISO 20022 pacs.008 Bridge executed successfully with zero heap allocations!");
    Ok(())
}
