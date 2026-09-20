#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

POLYXML_BIN="polyxml"
if [ -f "${ROOT_DIR}/../PolyXML/target/release/polyxml" ] && [ -f "${ROOT_DIR}/../PolyXML/target/debug/polyxml" ]; then
    if [ "${ROOT_DIR}/../PolyXML/target/release/polyxml" -nt "${ROOT_DIR}/../PolyXML/target/debug/polyxml" ]; then
        POLYXML_BIN="${ROOT_DIR}/../PolyXML/target/release/polyxml"
    else
        POLYXML_BIN="${ROOT_DIR}/../PolyXML/target/debug/polyxml"
    fi
elif [ -f "${ROOT_DIR}/../PolyXML/target/release/polyxml" ]; then
    POLYXML_BIN="${ROOT_DIR}/../PolyXML/target/release/polyxml"
elif [ -f "${ROOT_DIR}/../PolyXML/target/debug/polyxml" ]; then
    POLYXML_BIN="${ROOT_DIR}/../PolyXML/target/debug/polyxml"
elif command -v polyxml &>/dev/null; then
    POLYXML_BIN="polyxml"
fi

echo "================================================================================"
echo "💳 PolyXML CLI Streaming & Transcoding Demo (Global Finance)"
echo "   Bridging FedNow Instant Payment Intent ↔ ISO 20022 pacs.008.001.10"
echo "================================================================================"

TMP_DIR="$(mktemp -d /tmp/polyxml-finance-transcode-XXXXXX)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo -e "\n[1] Input: FedNow / Stripe B2B Supplier Payment Intent (JSON)"
head -n 25 data/payment_intent_fednow.json
echo "..."

echo -e "\n[2] Streaming Pipe: ISO 20022 XML -> Canonical JSON via PolyXML CLI:"
cat data/pacs_008_customer_credit_transfer.xml | "${POLYXML_BIN}" transcode --to json --pretty > "${TMP_DIR}/pacs008_pipe.json"
head -n 30 "${TMP_DIR}/pacs008_pipe.json"
echo "..."

echo -e "\n[3] Streaming Pipe: Canonical JSON -> ISO 20022 XML via PolyXML CLI:"
cat "${TMP_DIR}/pacs008_pipe.json" | "${POLYXML_BIN}" transcode --to xml --root Document --pretty > "${TMP_DIR}/pacs008_pipe.xml"
head -n 30 "${TMP_DIR}/pacs008_pipe.xml"
echo "..."

echo -e "\n[4] Schema-Guided Transcoding: ISO 20022 XML -> Strongly-Typed JSON (XSD-Driven):"
"${POLYXML_BIN}" transcode --schema schemas/finance/pacs_008_core.xsd --pretty data/pacs_008_customer_credit_transfer.xml -o "${TMP_DIR}/pacs008_typed.json"
head -n 30 "${TMP_DIR}/pacs008_typed.json"
echo "..."

echo -e "\n[5] Validating ISO 20022 pacs.008 Core XML Schema:"
"${POLYXML_BIN}" validate schemas/finance/pacs_008_core.xsd

echo -e "\n✅ PolyXML CLI Bidirectional Streaming & Schema-Directed Transcoding Complete!"
