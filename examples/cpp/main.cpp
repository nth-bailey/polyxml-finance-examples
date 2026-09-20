#include <iostream>
#include <fstream>
#include <sstream>
#include <chrono>
#include <regex>
#include <filesystem>
#include <iomanip>
#include <vector>
#include <cassert>
#include "pacs_008_core.hpp"

namespace fs = std::filesystem;
using namespace polyxml::generated;

// Statically verify C++20 XmlModel concept
static_assert(XmlModel<FiToFiCustomerCreditTransfer>, "FiToFiCustomerCreditTransfer must satisfy C++20 XmlModel concept");
static_assert(XmlModel<GroupHeader>, "GroupHeader must satisfy C++20 XmlModel concept");
static_assert(XmlModel<CreditTransferTransactionInformation>, "CreditTransferTransactionInformation must satisfy C++20 XmlModel concept");
static_assert(XmlModel<CashAccount>, "CashAccount must satisfy C++20 XmlModel concept");

std::string extract_string(const std::string& json, const std::string& key) {
    std::regex re("\"" + key + "\"\\s*:\\s*\"([^\"]+)\"");
    std::smatch match;
    if (std::regex_search(json, match, re) && match.size() > 1) {
        return match[1].str();
    }
    return "";
}

double extract_double(const std::string& json, const std::string& key) {
    std::regex re("\"" + key + "\"\\s*:\\s*([-+]?[0-9]*\\.?[0-9]+)");
    std::smatch match;
    if (std::regex_search(json, match, re) && match.size() > 1) {
        return std::stod(match[1].str());
    }
    return 0.0;
}

std::string find_data_file() {
    const std::vector<std::string> candidates = {
        "data/payment_intent_fednow.json",
        "../../data/payment_intent_fednow.json",
        "../data/payment_intent_fednow.json"
    };
    for (const auto& p : candidates) {
        if (fs::exists(p)) {
            return fs::absolute(p).string();
        }
    }
    return "";
}

std::string serialize_xml(const FiToFiCustomerCreditTransfer& doc) {
    std::ostringstream oss;
    oss << "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n";
    oss << "<Document xmlns=\"urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10\">\n";
    oss << "  <GrpHdr>\n";
    oss << "    <MsgId>" << doc.grp_hdr.msg_id << "</MsgId>\n";
    oss << "    <CreDtTm>" << doc.grp_hdr.cre_dt_tm << "</CreDtTm>\n";
    oss << "    <NbOfTxs>" << doc.grp_hdr.nb_of_txs << "</NbOfTxs>\n";
    oss << "    <SttlmInf>\n";
    oss << "      <SettlementMethod>" << to_string(doc.grp_hdr.sttlm_inf.settlement_method) << "</SettlementMethod>\n";
    if (doc.grp_hdr.sttlm_inf.clearing_system) {
        oss << "      <ClearingSystem>" << *doc.grp_hdr.sttlm_inf.clearing_system << "</ClearingSystem>\n";
    }
    oss << "    </SttlmInf>\n";
    oss << "  </GrpHdr>\n";

    for (const auto& tx : doc.cdt_trf_tx_inf) {
        oss << "  <CdtTrfTxInf>\n";
        oss << "    <PmtId>\n";
        oss << "      <InstructionId>" << tx.pmt_id.instruction_id << "</InstructionId>\n";
        oss << "      <EndToEndId>" << tx.pmt_id.end_to_end_id << "</EndToEndId>\n";
        oss << "      <TransactionId>" << tx.pmt_id.transaction_id << "</TransactionId>\n";
        oss << "      <UETR>" << tx.pmt_id.uetr << "</UETR>\n";
        oss << "    </PmtId>\n";
        oss << "    <IntrBkSttlmAmt Currency=\"" << tx.intr_bk_sttlm_amt.currency << "\">"
            << std::fixed << std::setprecision(2) << tx.intr_bk_sttlm_amt.value << "</IntrBkSttlmAmt>\n";
        oss << "    <IntrBkSttlmDt>" << tx.intr_bk_sttlm_dt << "</IntrBkSttlmDt>\n";
        oss << "    <ChargeBearer>" << to_string(tx.charge_bearer) << "</ChargeBearer>\n";

        oss << "    <Dbtr>\n";
        oss << "      <Name>" << tx.dbtr.name << "</Name>\n";
        if (tx.dbtr.postal_address) {
            oss << "      <PostalAddress>\n";
            if (tx.dbtr.postal_address->street_name) oss << "        <StreetName>" << *tx.dbtr.postal_address->street_name << "</StreetName>\n";
            if (tx.dbtr.postal_address->building_number) oss << "        <BuildingNumber>" << *tx.dbtr.postal_address->building_number << "</BuildingNumber>\n";
            if (tx.dbtr.postal_address->post_code) oss << "        <PostCode>" << *tx.dbtr.postal_address->post_code << "</PostCode>\n";
            oss << "        <TownName>" << tx.dbtr.postal_address->town_name << "</TownName>\n";
            oss << "        <Country>" << tx.dbtr.postal_address->country << "</Country>\n";
            oss << "      </PostalAddress>\n";
        }
        if (tx.dbtr.country_of_residence) oss << "      <CountryOfResidence>" << *tx.dbtr.country_of_residence << "</CountryOfResidence>\n";
        oss << "    </Dbtr>\n";

        oss << "    <DbtrAcct>\n";
        oss << "      <Id>\n";
        if (tx.dbtr_acct.id.proprietary_account) oss << "        <ProprietaryAccount>" << *tx.dbtr_acct.id.proprietary_account << "</ProprietaryAccount>\n";
        oss << "      </Id>\n";
        if (tx.dbtr_acct.currency) oss << "      <Currency>" << *tx.dbtr_acct.currency << "</Currency>\n";
        if (tx.dbtr_acct.name) oss << "      <Name>" << *tx.dbtr_acct.name << "</Name>\n";
        oss << "    </DbtrAcct>\n";

        oss << "    <DbtrAgt>\n";
        oss << "      <FinInstnId>\n";
        if (tx.dbtr_agt.fin_instn_id.bicfi) oss << "        <BICFI>" << *tx.dbtr_agt.fin_instn_id.bicfi << "</BICFI>\n";
        if (tx.dbtr_agt.fin_instn_id.clearing_system_member_id) {
            oss << "        <ClearingSystemMemberId>\n";
            oss << "          <ClearingSystemId>" << tx.dbtr_agt.fin_instn_id.clearing_system_member_id->clearing_system_id << "</ClearingSystemId>\n";
            oss << "          <MemberId>" << tx.dbtr_agt.fin_instn_id.clearing_system_member_id->member_id << "</MemberId>\n";
            oss << "        </ClearingSystemMemberId>\n";
        }
        if (tx.dbtr_agt.fin_instn_id.name) oss << "        <Name>" << *tx.dbtr_agt.fin_instn_id.name << "</Name>\n";
        oss << "      </FinInstnId>\n";
        oss << "    </DbtrAgt>\n";

        oss << "    <CdtrAgt>\n";
        oss << "      <FinInstnId>\n";
        if (tx.cdtr_agt.fin_instn_id.bicfi) oss << "        <BICFI>" << *tx.cdtr_agt.fin_instn_id.bicfi << "</BICFI>\n";
        if (tx.cdtr_agt.fin_instn_id.clearing_system_member_id) {
            oss << "        <ClearingSystemMemberId>\n";
            oss << "          <ClearingSystemId>" << tx.cdtr_agt.fin_instn_id.clearing_system_member_id->clearing_system_id << "</ClearingSystemId>\n";
            oss << "          <MemberId>" << tx.cdtr_agt.fin_instn_id.clearing_system_member_id->member_id << "</MemberId>\n";
            oss << "        </ClearingSystemMemberId>\n";
        }
        if (tx.cdtr_agt.fin_instn_id.name) oss << "        <Name>" << *tx.cdtr_agt.fin_instn_id.name << "</Name>\n";
        oss << "      </FinInstnId>\n";
        oss << "    </CdtrAgt>\n";

        oss << "    <Cdtr>\n";
        oss << "      <Name>" << tx.cdtr.name << "</Name>\n";
        if (tx.cdtr.postal_address) {
            oss << "      <PostalAddress>\n";
            if (tx.cdtr.postal_address->street_name) oss << "        <StreetName>" << *tx.cdtr.postal_address->street_name << "</StreetName>\n";
            if (tx.cdtr.postal_address->building_number) oss << "        <BuildingNumber>" << *tx.cdtr.postal_address->building_number << "</BuildingNumber>\n";
            if (tx.cdtr.postal_address->post_code) oss << "        <PostCode>" << *tx.cdtr.postal_address->post_code << "</PostCode>\n";
            oss << "        <TownName>" << tx.cdtr.postal_address->town_name << "</TownName>\n";
            oss << "        <Country>" << tx.cdtr.postal_address->country << "</Country>\n";
            oss << "      </PostalAddress>\n";
        }
        if (tx.cdtr.country_of_residence) oss << "      <CountryOfResidence>" << *tx.cdtr.country_of_residence << "</CountryOfResidence>\n";
        oss << "    </Cdtr>\n";

        oss << "    <CdtrAcct>\n";
        oss << "      <Id>\n";
        if (tx.cdtr_acct.id.proprietary_account) oss << "        <ProprietaryAccount>" << *tx.cdtr_acct.id.proprietary_account << "</ProprietaryAccount>\n";
        oss << "      </Id>\n";
        if (tx.cdtr_acct.currency) oss << "      <Currency>" << *tx.cdtr_acct.currency << "</Currency>\n";
        if (tx.cdtr_acct.name) oss << "      <Name>" << *tx.cdtr_acct.name << "</Name>\n";
        oss << "    </CdtrAcct>\n";

        if (tx.purpose) oss << "    <Purpose>" << *tx.purpose << "</Purpose>\n";
        if (tx.rmt_inf && tx.rmt_inf->unstructured) {
            oss << "    <RmtInf>\n";
            oss << "      <Unstructured>" << *tx.rmt_inf->unstructured << "</Unstructured>\n";
            oss << "    </RmtInf>\n";
        }
        oss << "  </CdtTrfTxInf>\n";
    }
    oss << "</Document>";
    return oss.str();
}

std::string serialize_json(const FiToFiCustomerCreditTransfer& doc) {
    std::ostringstream oss;
    oss << "{\n";
    oss << "  \"GrpHdr\": {\n";
    oss << "    \"MsgId\": \"" << doc.grp_hdr.msg_id << "\",\n";
    oss << "    \"CreDtTm\": \"" << doc.grp_hdr.cre_dt_tm << "\",\n";
    oss << "    \"NbOfTxs\": " << doc.grp_hdr.nb_of_txs << ",\n";
    oss << "    \"SttlmInf\": {\n";
    oss << "      \"SettlementMethod\": \"" << to_string(doc.grp_hdr.sttlm_inf.settlement_method) << "\"";
    if (doc.grp_hdr.sttlm_inf.clearing_system) {
        oss << ",\n      \"ClearingSystem\": \"" << *doc.grp_hdr.sttlm_inf.clearing_system << "\"";
    }
    oss << "\n    }\n";
    oss << "  },\n";
    oss << "  \"CdtTrfTxInf\": [\n";
    for (size_t i = 0; i < doc.cdt_trf_tx_inf.size(); ++i) {
        const auto& tx = doc.cdt_trf_tx_inf[i];
        oss << "    {\n";
        oss << "      \"PmtId\": {\n";
        oss << "        \"InstructionId\": \"" << tx.pmt_id.instruction_id << "\",\n";
        oss << "        \"EndToEndId\": \"" << tx.pmt_id.end_to_end_id << "\",\n";
        oss << "        \"TransactionId\": \"" << tx.pmt_id.transaction_id << "\",\n";
        oss << "        \"UETR\": \"" << tx.pmt_id.uetr << "\"\n";
        oss << "      },\n";
        oss << "      \"IntrBkSttlmAmt\": {\n";
        oss << "        \"Currency\": \"" << tx.intr_bk_sttlm_amt.currency << "\",\n";
        oss << "        \"Value\": " << std::fixed << std::setprecision(2) << tx.intr_bk_sttlm_amt.value << "\n";
        oss << "      },\n";
        oss << "      \"IntrBkSttlmDt\": \"" << tx.intr_bk_sttlm_dt << "\",\n";
        oss << "      \"ChargeBearer\": \"" << to_string(tx.charge_bearer) << "\",\n";
        oss << "      \"Dbtr\": { \"Name\": \"" << tx.dbtr.name << "\" },\n";
        oss << "      \"Cdtr\": { \"Name\": \"" << tx.cdtr.name << "\" }\n";
        oss << "    }" << (i + 1 < doc.cdt_trf_tx_inf.size() ? "," : "") << "\n";
    }
    oss << "  ]\n";
    oss << "}";
    return oss.str();
}

int main() {
    std::cout << "================================================================================\n";
    std::cout << "  PolyXML Modern C++20 Showcase: FedNow Payment Intent -> ISO 20022 pacs.008\n";
    std::cout << "================================================================================\n";

    std::string dataPath = find_data_file();
    if (dataPath.empty()) {
        std::cerr << "Error: Could not locate data/payment_intent_fednow.json\n";
        return 1;
    }

    std::ifstream file(dataPath);
    std::stringstream buffer;
    buffer << file.rdbuf();
    std::string jsonStr = buffer.str();

    // 1. Ingest JSON payment intent fields
    FiToFiCustomerCreditTransfer doc;
    doc.grp_hdr.msg_id = extract_string(jsonStr, "message_id");
    doc.grp_hdr.cre_dt_tm = extract_string(jsonStr, "created_at");
    doc.grp_hdr.nb_of_txs = 1;
    doc.grp_hdr.sttlm_inf.settlement_method = settlement_method_code_from_string(extract_string(jsonStr, "settlement_method")).value_or(SettlementMethodCode::Clrg);
    doc.grp_hdr.sttlm_inf.clearing_system = extract_string(jsonStr, "clearing_network");

    CreditTransferTransactionInformation tx;
    tx.pmt_id.instruction_id = extract_string(jsonStr, "instruction_id");
    tx.pmt_id.end_to_end_id = extract_string(jsonStr, "end_to_end_id");
    tx.pmt_id.transaction_id = extract_string(jsonStr, "transaction_id");
    tx.pmt_id.uetr = extract_string(jsonStr, "uetr");

    tx.intr_bk_sttlm_amt.currency = extract_string(jsonStr, "currency");
    tx.intr_bk_sttlm_amt.value = extract_double(jsonStr, "amount");
    tx.intr_bk_sttlm_dt = extract_string(jsonStr, "settlement_date");
    tx.charge_bearer = charge_bearer_type_from_string(extract_string(jsonStr, "charge_bearer")).value_or(ChargeBearerType::Slev);

    tx.dbtr.name = extract_string(jsonStr, "name");
    PostalAddress dbtrAddr;
    dbtrAddr.street_name = extract_string(jsonStr, "street_name");
    dbtrAddr.building_number = extract_string(jsonStr, "building_number");
    dbtrAddr.post_code = extract_string(jsonStr, "post_code");
    dbtrAddr.town_name = extract_string(jsonStr, "town_name");
    dbtrAddr.country = extract_string(jsonStr, "country");
    tx.dbtr.postal_address = dbtrAddr;
    tx.dbtr.country_of_residence = extract_string(jsonStr, "country_of_residence");

    tx.dbtr_acct.id.proprietary_account = extract_string(jsonStr, "account_number");
    tx.dbtr_acct.currency = extract_string(jsonStr, "currency");
    tx.dbtr_acct.name = extract_string(jsonStr, "account_name");

    tx.dbtr_agt.fin_instn_id.bicfi = extract_string(jsonStr, "bicfi");
    ClearingSystemMemberIdentification dbtrMember;
    dbtrMember.clearing_system_id = extract_string(jsonStr, "clearing_system_id");
    dbtrMember.member_id = extract_string(jsonStr, "routing_number");
    tx.dbtr_agt.fin_instn_id.clearing_system_member_id = dbtrMember;

    // Creditor
    tx.cdtr.name = "Horizon Maritime Logistics International";
    PostalAddress cdtrAddr;
    cdtrAddr.street_name = "100 Commercial Wharf";
    cdtrAddr.building_number = "Pier 4";
    cdtrAddr.post_code = "02110";
    cdtrAddr.town_name = "Boston";
    cdtrAddr.country = "US";
    tx.cdtr.postal_address = cdtrAddr;
    tx.cdtr.country_of_residence = "US";

    tx.cdtr_acct.id.proprietary_account = "883719204918";
    tx.cdtr_acct.currency = "USD";
    tx.cdtr_acct.name = "Horizon Fleet Operations Settlement";

    tx.cdtr_agt.fin_instn_id.bicfi = "IRVTUS3NXXX";
    ClearingSystemMemberIdentification cdtrMember;
    cdtrMember.clearing_system_id = "USABA";
    cdtrMember.member_id = "021000018";
    tx.cdtr_agt.fin_instn_id.clearing_system_member_id = cdtrMember;
    tx.cdtr_agt.fin_instn_id.name = "The Bank of New York Mellon";

    tx.purpose = extract_string(jsonStr, "purpose_code");
    RemittanceInformation rmt;
    rmt.unstructured = extract_string(jsonStr, "remittance");
    tx.rmt_inf = rmt;

    doc.cdt_trf_tx_inf.push_back(tx);

    // 2. Validate model using PolyXML methods
    if (!doc.validate()) {
        std::cerr << "Validation failed!\n";
        return 1;
    }
    std::cout << "✔ ISO 20022 pacs.008 C++20 schema validation passed!\n";

    // 1. Inherent XML Serialization
    auto t1 = std::chrono::high_resolution_clock::now();
    std::string xml = serialize_xml(doc);
    auto t2 = std::chrono::high_resolution_clock::now();
    double xml_us = std::chrono::duration<double, std::micro>(t2 - t1).count();

    std::cout << "\n[1] Generated ISO 20022 pacs.008.001.10 XML Message (latency: "
              << std::fixed << std::setprecision(2) << xml_us << "µs):\n";
    std::cout << xml.substr(0, std::min<size_t>(xml.size(), 400)) << "\n...\n" << std::endl;

    // 2. Inherent Native JSON Serialization on Same Model
    auto t3 = std::chrono::high_resolution_clock::now();
    std::string jsonWire = serialize_json(doc);
    auto t4 = std::chrono::high_resolution_clock::now();
    double json_us = std::chrono::duration<double, std::micro>(t4 - t3).count();

    std::cout << "[2] Generated Native JSON on Same Model (latency: "
              << std::fixed << std::setprecision(2) << json_us << "µs):\n";
    std::cout << jsonWire.substr(0, std::min<size_t>(jsonWire.size(), 400)) << "\n...\n" << std::endl;

    // 3. Model Inspection & Concept Checks
    std::cout << "[3] C++20 Value Type Inspection:" << std::endl;
    std::cout << "    MsgId: " << doc.grp_hdr.msg_id << std::endl;
    std::cout << "    UETR:  " << doc.cdt_trf_tx_inf[0].pmt_id.uetr << std::endl;
    std::cout << "    Amount: " << std::fixed << std::setprecision(2) << doc.cdt_trf_tx_inf[0].intr_bk_sttlm_amt.value
              << " " << doc.cdt_trf_tx_inf[0].intr_bk_sttlm_amt.currency << std::endl;
    std::cout << "    Debtor: " << doc.cdt_trf_tx_inf[0].dbtr.name << std::endl;
    std::cout << "    Creditor: " << doc.cdt_trf_tx_inf[0].cdtr.name << " via "
              << *doc.cdt_trf_tx_inf[0].cdtr_agt.fin_instn_id.name << std::endl;
    std::cout << "    C++20 XmlModel concept static_assert check: PASS" << std::endl;

    std::cout << "\n✅ C++20 Modern Payments ↔ ISO 20022 pacs.008 Bridge executed successfully!" << std::endl;
    return 0;
}

