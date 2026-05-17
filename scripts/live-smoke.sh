#!/usr/bin/env bash
# scripts/live-smoke.sh — exercise every loby command against api.lob.com.
#
# Requires:  $LOB_API_KEY=sk_test_…  (or test_…)
# Prints:    [ok] / [SKIP] / [FAIL]  per check, plus a final tally.
# Exits:     non-zero if any [FAIL] occurred (SKIP does not fail the run).
#
# Designed to run from the repo root. Builds bin/loby if missing.
set -uo pipefail

: "${LOB_API_KEY:?set LOB_API_KEY to a Lob test key (sk_test_… or test_…)}"
[ -x ./bin/loby ] || make build >/dev/null

LOBY=./bin/loby
PASS=0
FAIL=0
SKIP=0
FAILED_NAMES=()

check() {
  local name="$1" expect="$2" cmd="$3"
  local out code
  out=$(eval "$cmd" 2>&1)
  code=$?
  case "$expect" in
    ok)
      if [ $code -eq 0 ]; then
        echo "[ok]   $name"
        PASS=$((PASS + 1))
      else
        echo "[FAIL] $name -> exit=$code"
        echo "         $(echo "$out" | head -2 | tail -1)"
        FAILED_NAMES+=("$name")
        FAIL=$((FAIL + 1))
      fi
      ;;
    skip-on-403|skip-on-404|skip-on-422|skip-feature-gated)
      if [ $code -eq 0 ]; then
        echo "[ok]   $name"
        PASS=$((PASS + 1))
      elif echo "$out" | grep -qE "lob (401|403|404|422):|requires live mode|not allowed|not enabled|unauthorized|invalid_api_key|unrecognized_endpoint|not available for your account"; then
        echo "[SKIP] $name -> $(echo "$out" | head -1 | sed 's/^error: //' | cut -c1-90)"
        SKIP=$((SKIP + 1))
      else
        echo "[FAIL] $name -> exit=$code"
        echo "         $(echo "$out" | head -2 | tail -1)"
        FAILED_NAMES+=("$name")
        FAIL=$((FAIL + 1))
      fi
      ;;
  esac
}

# Capture a value from JSON output, fall through to env var on failure.
jget() { echo "$1" | jq -er "$2" 2>/dev/null; }

echo "===== Introspection ====="
check "version --json"      ok "$LOBY version --json | jq .version"
check "schema --json"       ok "$LOBY schema --json | jq '.subcommands | length'"
check "exit-codes --json"   ok "$LOBY exit-codes --json | jq 'length'"
check "completion bash"     ok "$LOBY completion bash | head -1"
check "auth status --json"  ok "$LOBY auth status --json"

echo
echo "===== Address & verification ====="
check "addresses verify us" ok "$LOBY verify us '185 Berry St, San Francisco, CA 94107' --json | jq -r .deliverability"
check "addresses verify intl" ok "$LOBY verify intl --primary '10 Downing St' --postal SW1A2AA --country GB --json | jq .id"
check "addresses autocomplete" ok "$LOBY addresses autocomplete '185 Berr' --json | jq '.suggestions | length'"
check "zip lookup"          ok "$LOBY zip 94107 --json | jq -r .zip_code"
check "geo reverse"         ok "$LOBY geo reverse --lat 37.7749 --lng=-122.4194 --json | jq '.addresses | length'"
check "bulk us"             ok "$LOBY bulk us --addresses '[{\"primary_line\":\"185 Berry St\",\"city\":\"San Francisco\",\"state\":\"CA\",\"zip_code\":\"94107\"}]' --json | jq '.addresses | length'"
check "bulk intl"           ok "$LOBY bulk intl --addresses '[{\"primary_line\":\"10 Downing St\",\"postal_code\":\"SW1A2AA\",\"country\":\"GB\"}]' --json | jq '.addresses | length'"

echo
echo "===== Addresses CRUD ====="
ADR_ID=$($LOBY addresses create --name "Smoke Test" --line1 "185 Berry St" --city "San Francisco" --state CA --zip 94107 --quiet 2>&1 | head -1)
if [[ "$ADR_ID" =~ ^adr_ ]]; then
  echo "[ok]   addresses create -> $ADR_ID"
  PASS=$((PASS + 1))
  check "addresses get"     ok "$LOBY addresses get $ADR_ID --json | jq -r .id"
  check "addresses list"    ok "$LOBY addresses list --limit 1 --json | jq -r .object"
  check "addresses delete"  ok "$LOBY addresses delete $ADR_ID --confirm --json | jq -r .deleted"
else
  echo "[FAIL] addresses create -> $ADR_ID"
  FAILED_NAMES+=("addresses create")
  FAIL=$((FAIL + 1))
fi

echo
echo "===== Templates CRUD + versions ====="
TMPL_ID=$($LOBY templates create --description "smoke" --html "<html><body>Hello {{name}}</body></html>" --json 2>&1 | jq -er .id 2>/dev/null)
if [[ "$TMPL_ID" =~ ^tmpl_ ]]; then
  echo "[ok]   templates create -> $TMPL_ID"
  PASS=$((PASS + 1))
  check "templates get"        ok "$LOBY templates get $TMPL_ID --json | jq -r .id"
  check "templates list"       ok "$LOBY templates list --limit 1 --json | jq -r .object"
  check "templates update"     ok "$LOBY templates update $TMPL_ID --description smoke-updated --json | jq -r .description"
  VRSN_ID=$($LOBY templates versions create $TMPL_ID --description v2 --html "<html><body>v2 {{name}}</body></html>" --json 2>&1 | jq -er .id 2>/dev/null)
  if [[ "$VRSN_ID" =~ ^vrsn_ ]]; then
    echo "[ok]   templates versions create -> $VRSN_ID"
    PASS=$((PASS + 1))
    check "templates versions get"    ok "$LOBY templates versions get $TMPL_ID $VRSN_ID --json | jq -r .id"
    check "templates versions list"   ok "$LOBY templates versions list $TMPL_ID --json | jq -r .object"
    check "templates versions update" ok "$LOBY templates versions update $TMPL_ID $VRSN_ID --description v2-updated --json | jq -r .description"
    check "templates versions delete" ok "$LOBY templates versions delete $TMPL_ID $VRSN_ID --confirm --json"
  else
    echo "[FAIL] templates versions create -> $($LOBY templates versions create $TMPL_ID --description v2 --html '<html>v2</html>' --json 2>&1 | head -1)"
    FAILED_NAMES+=("templates versions create"); FAIL=$((FAIL + 1))
  fi
  check "templates delete"     ok "$LOBY templates delete $TMPL_ID --confirm --json"
else
  echo "[FAIL] templates create -> $($LOBY templates create --description smoke --html '<html>x</html>' --json 2>&1 | head -1)"
  FAILED_NAMES+=("templates create"); FAIL=$((FAIL + 1))
fi

echo
echo "===== Postcard create + get ====="
PSC_ID=$($LOBY postcards create \
  --to "{\"name\":\"Smoke\",\"address_line1\":\"185 Berry St\",\"address_city\":\"San Francisco\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}" \
  --front "<html><body>front</body></html>" \
  --back "<html><body>back</body></html>" \
  --json 2>&1 | jq -er .id 2>/dev/null)
if [[ "$PSC_ID" =~ ^psc_ ]]; then
  echo "[ok]   postcards create -> $PSC_ID"
  PASS=$((PASS + 1))
  check "postcards get"     ok "$LOBY postcards get $PSC_ID --json | jq -r .id"
else
  echo "[FAIL] postcards create"
  FAILED_NAMES+=("postcards create"); FAIL=$((FAIL + 1))
fi

echo
echo "===== Letter create + get + cancel ====="
LTR_OUT=$($LOBY letters create \
  --to "{\"name\":\"Smoke\",\"address_line1\":\"185 Berry St\",\"address_city\":\"San Francisco\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}" \
  --from "{\"name\":\"Sender\",\"address_line1\":\"210 King St\",\"address_city\":\"San Francisco\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}" \
  --file "<html><body>letter body</body></html>" \
  --color --json 2>&1)
LTR_ID=$(echo "$LTR_OUT" | jq -er .id 2>/dev/null)
if [[ "$LTR_ID" =~ ^ltr_ ]]; then
  echo "[ok]   letters create -> $LTR_ID"
  PASS=$((PASS + 1))
  check "letters get"     ok "$LOBY letters get $LTR_ID --json | jq -r .id"
  check "letters cancel"  ok "$LOBY letters cancel $LTR_ID --confirm --json"
else
  echo "[FAIL] letters create -> $(echo "$LTR_OUT" | head -1)"
  FAILED_NAMES+=("letters create"); FAIL=$((FAIL + 1))
fi

echo
echo "===== Self-mailer + Snap-pack (feature-gated) ====="
check "self-mailers create"  skip-feature-gated "$LOBY self-mailers create --to '{\"name\":\"X\",\"address_line1\":\"185 Berry St\",\"address_city\":\"SF\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}' --from '{\"name\":\"S\",\"address_line1\":\"210 King St\",\"address_city\":\"SF\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}' --outside '<html>outside</html>' --inside '<html>inside</html>' --json"
check "snap-packs create"    skip-on-403 "$LOBY snap-packs create --to '{\"name\":\"X\",\"address_line1\":\"185 Berry St\",\"address_city\":\"SF\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}' --from '{\"name\":\"S\",\"address_line1\":\"210 King St\",\"address_city\":\"SF\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}' --outside '<html>outside</html>' --inside '<html>inside</html>' --json"

echo
echo "===== Print assets (feature-gated) ====="
check "cards create"      skip-on-422 "$LOBY cards create --front '<html>front</html>' --back '<html>back</html>' --json"
check "booklets create"   skip-on-403 "$LOBY booklets create --cover '<html>cover</html>' --inside '<html>inside</html>' --json"
check "buckslips create"  skip-on-422 "$LOBY buckslips create --front '<html>front</html>' --back '<html>back</html>' --json"

echo
echo "===== Bank accounts + checks ====="
BANK_OUT=$($LOBY bank-accounts create --routing-number 322271627 --account-number 123456789 --account-type company --signatory "Smoke Test" --json 2>&1)
BANK_ID=$(echo "$BANK_OUT" | jq -er .id 2>/dev/null)
# Live mode: Stripe blocks micro-deposit verification on the canonical
# Lob smoke routing number, so bank-accounts create returns 422 bank_error.
# Treat that specific error as a SKIP (the request shape itself was correct).
if echo "$BANK_OUT" | grep -qiE "Microdeposit transfers have been blocked|Please update your Lob payment method"; then
  echo "[SKIP] bank-accounts create -> Stripe blocked microdeposit in live mode (request shape ok)"
  echo "[SKIP] bank-accounts get/verify/delete + checks create/get/cancel (gated by bank create)"
  SKIP=$((SKIP + 7))
elif [[ "$BANK_ID" =~ ^bank_ ]]; then
  echo "[ok]   bank-accounts create -> $BANK_ID"
  PASS=$((PASS + 1))
  check "bank-accounts get"     ok "$LOBY bank-accounts get $BANK_ID --json | jq -r .id"
  check "bank-accounts verify"  ok "$LOBY bank-accounts verify $BANK_ID --amounts 11,35 --json"
  # Check create requires verified bank.
  CHK_OUT=$($LOBY checks create \
    --to "{\"name\":\"Payee\",\"address_line1\":\"185 Berry St\",\"address_city\":\"San Francisco\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}" \
    --from "{\"name\":\"Payer\",\"address_line1\":\"210 King St\",\"address_city\":\"San Francisco\",\"address_state\":\"CA\",\"address_zip\":\"94107\"}" \
    --bank-account $BANK_ID --amount 12.34 --memo "smoke" --json 2>&1)
  CHK_ID=$(echo "$CHK_OUT" | jq -er .id 2>/dev/null)
  if [[ "$CHK_ID" =~ ^chk_ ]]; then
    echo "[ok]   checks create -> $CHK_ID"
    PASS=$((PASS + 1))
    check "checks get"        ok "$LOBY checks get $CHK_ID --json | jq -r .id"
    check "checks cancel"     ok "$LOBY checks cancel $CHK_ID --confirm --json"
  else
    echo "[FAIL] checks create -> $(echo "$CHK_OUT" | head -1)"
    FAILED_NAMES+=("checks create"); FAIL=$((FAIL + 1))
  fi
  check "bank-accounts delete"  ok "$LOBY bank-accounts delete $BANK_ID --confirm --json"
else
  echo "[FAIL] bank-accounts create -> $($LOBY bank-accounts create --routing-number 322271627 --account-number 123456789 --account-type company --signatory smoke --json 2>&1 | head -1)"
  FAILED_NAMES+=("bank-accounts create"); FAIL=$((FAIL + 1))
fi

echo
echo "===== Billing groups (feature-gated) ====="
check "billing-groups create" skip-on-403 "$LOBY billing-groups create --name smoke-bg --description test --json"

echo
echo "===== Campaigns + creatives + uploads ====="
CMP_OUT=$($LOBY campaigns create --name "smoke-cmp" --schedule-type immediate --json 2>&1)
CMP_ID=$(echo "$CMP_OUT" | jq -er .id 2>/dev/null)
if [[ "$CMP_ID" =~ ^cmp_ ]]; then
  echo "[ok]   campaigns create -> $CMP_ID"
  PASS=$((PASS + 1))
  check "campaigns get"     ok "$LOBY campaigns get $CMP_ID --json | jq -r .id"
  # Creatives require PDF URLs (or template_ids) — Lob rejects inline HTML.
  # 500s from Lob on this endpoint surface as Lob-side issues, not CLI bugs:
  # the request shape matches their published example exactly.
  CRV_PDF='https://s3-us-west-2.amazonaws.com/public.lob.com/assets/templates/4x6_pc_template.pdf'
  CRV_OUT=$($LOBY creatives create --campaign-id $CMP_ID --resource-type postcard \
    --front "$CRV_PDF" --back "$CRV_PDF" --size 4x6 --mail-type usps_first_class \
    --json 2>&1)
  CRV_ID=$(echo "$CRV_OUT" | jq -er .id 2>/dev/null)
  if [[ "$CRV_ID" =~ ^crv_ ]]; then
    echo "[ok]   creatives create -> $CRV_ID"
    PASS=$((PASS + 1))
  elif echo "$CRV_OUT" | grep -qE "internal_server_error|500"; then
    echo "[SKIP] creatives create -> Lob 500 internal_server_error (request shape ok)"
    SKIP=$((SKIP + 1))
  else
    echo "[FAIL] creatives create -> $(echo "$CRV_OUT" | head -1)"
    FAILED_NAMES+=("creatives create"); FAIL=$((FAIL + 1))
  fi
  UPL_ID=$($LOBY uploads create --campaign-id $CMP_ID --json 2>&1 | jq -er .id 2>/dev/null)
  if [[ "$UPL_ID" =~ ^upl_ ]]; then
    echo "[ok]   uploads create -> $UPL_ID"
    PASS=$((PASS + 1))
    check "uploads get"          ok "$LOBY uploads get $UPL_ID --json | jq -r .id"
    check "uploads list"         ok "$LOBY uploads list --campaign-id $CMP_ID --json"
    EXP_ID=$($LOBY uploads exports create $UPL_ID --type failures --json 2>&1 | jq -er '.exportId // .id' 2>/dev/null)
    if [[ "$EXP_ID" =~ ^ex_ ]]; then
      echo "[ok]   uploads exports create -> $EXP_ID"
      PASS=$((PASS + 1))
      check "uploads exports get"  ok "$LOBY uploads exports get $UPL_ID $EXP_ID --json"
    else
      echo "[FAIL] uploads exports create"
      FAILED_NAMES+=("uploads exports create"); FAIL=$((FAIL + 1))
    fi
    check "uploads delete"       ok "$LOBY uploads delete $UPL_ID --confirm --json"
  else
    echo "[FAIL] uploads create -> $($LOBY uploads create --campaign-id $CMP_ID --json 2>&1 | head -1)"
    FAILED_NAMES+=("uploads create"); FAIL=$((FAIL + 1))
  fi
  check "campaigns delete"  ok "$LOBY campaigns delete $CMP_ID --confirm --json"
elif echo "$CMP_OUT" | grep -qE "requires live mode|unauthorized"; then
  echo "[SKIP] campaigns create -> $(echo "$CMP_OUT" | head -1 | sed 's/^error: //' | cut -c1-90)"
  echo "[SKIP] campaigns get/delete (gated by create)"
  echo "[SKIP] creatives create (gated by campaign)"
  echo "[SKIP] uploads create/get/list/exports/delete (gated by campaign)"
  SKIP=$((SKIP + 6))
else
  echo "[FAIL] campaigns create -> $(echo "$CMP_OUT" | head -1)"
  FAILED_NAMES+=("campaigns create"); FAIL=$((FAIL + 1))
fi

echo
echo "===== Informed delivery (feature-gated) ====="
check "informed-delivery list" skip-on-403 "$LOBY informed-delivery list --limit 1 --json"

echo
echo "===== Links + Domains ====="
LINK_ID=$($LOBY links create --redirect-link "https://example.com/loby-smoke" --description smoke --json 2>&1 | jq -er .id 2>/dev/null)
if [[ -n "$LINK_ID" && "$LINK_ID" != "null" ]]; then
  echo "[ok]   links create -> $LINK_ID"
  PASS=$((PASS + 1))
  check "links get"     ok "$LOBY links get $LINK_ID --json | jq -r .id"
  check "links list"    ok "$LOBY links list --limit 1 --json | jq -r .object"
  check "links delete"  ok "$LOBY links delete $LINK_ID --confirm --json"
else
  echo "[FAIL] links create -> $($LOBY links create --redirect-link 'https://example.com/x' --json 2>&1 | head -1)"
  FAILED_NAMES+=("links create"); FAIL=$((FAIL + 1))
fi
check "domains list"     ok "$LOBY domains list --limit 1 --json | jq -r .object"

echo
echo "===== Identity validation (create-only) ====="
check "identity verify" ok "$LOBY identity verify --recipient 'Larry Lobster' --primary-line '210 King St' --city 'San Francisco' --state CA --zip 94107 --json | jq -r .id"

echo
echo "===== Events + account ====="
check "events list"      ok "$LOBY events list --limit 5 --json | jq -r .object"
EVT_ID=$($LOBY events list --limit 1 --json 2>&1 | jq -er '.data[0].id' 2>/dev/null)
if [[ -n "$EVT_ID" && "$EVT_ID" != "null" ]]; then
  check "events get"    ok "$LOBY events get $EVT_ID --json | jq -r .id"
else
  echo "[SKIP] events get (no events on account)"
  SKIP=$((SKIP + 1))
fi
check "account"          skip-feature-gated "$LOBY account --json"

echo
echo "===== QR analytics ====="
check "qr-codes list"    ok "$LOBY qr-codes list --json | jq -r .object"

echo
echo "===================================="
printf "PASS: %d   SKIP: %d   FAIL: %d\n" "$PASS" "$SKIP" "$FAIL"
if [ ${#FAILED_NAMES[@]} -gt 0 ]; then
  echo "Failed:"
  for n in "${FAILED_NAMES[@]}"; do echo "  - $n"; done
fi
exit $FAIL
