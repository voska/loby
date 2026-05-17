# loby Recipes

Tested end-to-end flows. Every step uses `--json` so an agent can capture IDs from `jq`.

## Recipe 1 — Send a single postcard, the safe way

```bash
# 1. Verify the recipient is deliverable.
deliv=$(loby verify us "185 Berry St, San Francisco, CA 94107" \
  --json --select deliverability,components.zip_code | jq -r .deliverability)
[ "$deliv" = "deliverable" ] || { echo "address not deliverable: $deliv" >&2; exit 1; }

# 2. Save the address so it can be reused.
adr=$(loby addresses create --name "Alice Park" --company "Acme Inc" \
  --line1 "185 Berry St" --city "San Francisco" --state CA --zip 94107 \
  --json | jq -r .id)

# 3. Dry-run the postcard to inspect the body.
loby postcards create --to "$adr" --front @front.html --back @back.html \
  --size 4x6 --dry-run --json

# 4. Send.
psc=$(loby postcards create --to "$adr" --front @front.html --back @back.html \
  --size 4x6 --json | jq -r .id)
echo "mailed: $psc"
```

## Recipe 2 — Bulk verify before mailing

```bash
# Input: addresses.json is an array of {name, address_line1, address_city, address_state, address_zip}.
loby bulk us --addresses @addresses.json --json \
  | jq '[.addresses[] | select(.deliverability == "deliverable")]' \
  > deliverable.json
```

## Recipe 3 — Issue a check from a verified bank account

```bash
# One-time: add and verify the bank account.
bank=$(loby bank-accounts create --routing-number 122100024 --account-number 123456789 \
  --account-type company --signatory "Jane Doe" --description "Operating account" \
  --json | jq -r .id)

# Lob makes two micro-deposits; once they land:
loby bank-accounts verify "$bank" --amounts 11,35 --json

# Send a check.
loby checks create --to "$adr" --bank-account "$bank" --amount 250.00 \
  --memo "Invoice #1234" --json
```

## Recipe 4 — Stream events into a log

```bash
loby events tail --resource-type postcards --interval 10s --json \
  | tee -a events.ndjson
```

> Note: `campaigns`, `creatives`, `uploads`, `informed-delivery`, `account` (credits balance), `billing-groups`, `snap-packs`, `booklets`, `cards`, and `buckslips` are gated by account features or live mode on Lob's side. With a sandbox key they return 401/403/422. Their CLI surface is implemented to spec but unverified end-to-end on test mode.

## Recipe 5 — End-to-end direct mail campaign

```bash
# Templates.
tmpl_front=$(loby templates create --description "Promo front" --html @front.html --engine handlebars --json | jq -r .id)
tmpl_back=$(loby templates create --description "Promo back" --html @back.html --engine handlebars --json | jq -r .id)

# Campaign.
cmp=$(loby campaigns create --name "Q3 Spring Promo" --schedule-type in_future \
  --send-date 2026-07-01 --json | jq -r .id)

# Creative. Lob's /v1/creatives expects PDF URLs or template_ids on
# --front/--back; inline HTML is rejected. Mail type lives in details:{}
# alongside size.
loby creatives create --campaign-id "$cmp" --resource-type postcard \
  --front "$tmpl_front" --back "$tmpl_back" \
  --size 4x6 --mail-type usps_first_class --json

# CSV upload.
upl=$(loby uploads create --campaign-id "$cmp" --json | jq -r .id)
loby uploads file "$upl" ./recipients.csv --json

# Wait for verification. The status is in the upload record itself.
while [ "$(loby uploads get "$upl" --json | jq -r .state)" != "validated" ]; do
  sleep 30
done

# (Optional) inspect row errors before sending. Export create returns the
# job id under .exportId (not .id). `uploads exports` is a live-mode feature.
export_id=$(loby uploads exports create "$upl" --type failures --json | jq -r .exportId)
loby uploads exports get "$upl" "$export_id" --json

# Submit.
loby campaigns send "$cmp" --confirm --json
```

## Recipe 6 — Render a template preview without mailing

```bash
loby templates versions create <tmpl_id> --html @new_body.html --json
# The response includes the preview URL.
```

## Recipe 7 — Inspect a failure

```bash
# Any non-zero exit code is documented in `loby exit-codes --json`.
if ! loby postcards create --to adr_bad --front @x.html --json; then
  code=$?
  echo "exit code $code: $(loby exit-codes --json | jq -r ".[] | select(.code == $code).description")" >&2
fi
```

## Patterns

- **Always run `--dry-run` first** on mutations the user hasn't pre-authorized.
- **Capture IDs from `--json`** rather than scraping human output.
- **Re-run safely.** Every create carries an auto-generated idempotency key; Lob deduplicates for 24h.
- **Tolerate rate limits.** Exit code 7 means retry with backoff (Lob's `Retry-After` is respected automatically).
