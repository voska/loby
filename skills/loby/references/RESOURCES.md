# Lob Resource Glossary

Quick reference for what each Lob resource is and which `loby` command group manages it.

| Resource | ID prefix | `loby` command | Description |
| --- | --- | --- | --- |
| `account` | n/a | `loby account` | Active Lob account profile and balance. |
| `address` | `adr_` | `loby addresses` | Saved recipient address. |
| `bank_account` | `bank_` | `loby bank-accounts` | Bank routing/account used by checks. |
| `billing_group` | `bg_` | `loby billing-groups` | Tag for grouping invoices. |
| `booklet` | `bkl_` | `loby booklets` | Multi-page booklet artwork. |
| `buckslip` | `bck_` | `loby buckslips` | 8.75×3.75 promotional insert. |
| `bulk_us_verification` | n/a | `loby bulk us` | Sync verify ≤100 US addresses. |
| `bulk_intl_verification` | n/a | `loby bulk intl` | Sync verify ≤100 intl addresses. |
| `campaign` | `cmp_` | `loby campaigns` | Direct-mail campaign. |
| `card` | `card_` | `loby cards` | Card stock asset. |
| `check` | `chk_` | `loby checks` | Mailed paper check drawn on a verified bank account. |
| `creative` | `crv_` | `loby creatives` | Artwork attached to a campaign. |
| `event` | `evt_` | `loby events` | Audit-log entry for any resource state change. |
| `identity_validation` | `idv_` | `loby identity` | Recipient identity validation. |
| `informed_delivery_campaign` | `idc_` | `loby informed-delivery` | USPS Informed Delivery campaign. |
| `intl_verification` | `intl_ver_` | `loby verify intl` | International address verification. |
| `letter` | `ltr_` | `loby letters` | Mailed letter. |
| `postcard` | `psc_` | `loby postcards` | Mailed postcard. |
| `qr_code` | `qr_` | `loby qr-codes` | Trackable QR code. |
| `resource_proof` | `proof_` | `loby resource-proofs` | PDF preview of a printed asset. |
| `reverse_geocode_lookup` | n/a | `loby geo reverse` | Lat/lng → ZIP codes. |
| `self_mailer` | `sfm_` | `loby self-mailers` | Bi-folded mailer. |
| `short_url` | `su_` | `loby short-urls` | Trackable short URL. |
| `snap_pack` | `snp_` | `loby snap-packs` | Snap-pack mailer (self-sealing). |
| `template` | `tmpl_` | `loby templates` | Stored HTML with Handlebars variables. |
| `upload` | `upl_` | `loby uploads` | CSV upload for a campaign. |
| `us_autocompletion` | `us_auto_` | `loby addresses autocomplete` | Partial-address suggestions. |
| `us_csv_verification` | `us_csv_` | `loby bulk csv` | Async CSV verification. |
| `us_verification` | `us_ver_` | `loby verify us` | US address verification. |
| `zip_lookup` | n/a | `loby zip` | ZIP code → city/state. |

## Lifecycle notes

- **Mailers** (`postcard`, `letter`, `check`, `self_mailer`, `snap_pack`) can be canceled with `loby <type> cancel <id> --confirm` before they enter production. After that, only `update metadata` is allowed.
- **Campaigns** are editable until `loby campaigns send <id> --confirm`. After send, only metadata can change.
- **Bank accounts** must be verified via two micro-deposits before they can fund checks.
- **CSV uploads** transition `uploaded → verifying → verified → failed`. Only verified uploads can be mailed.
- **Events** are append-only and retained for 90 days.

## Authentication environments

- Keys prefixed `sk_test_…` operate in the sandbox — no postage charged, no actual mail.
- Keys prefixed `sk_live_…` operate in production — real postage, real mail, real money.
- Identify the active environment with `loby auth status --json` (`environment` field).
