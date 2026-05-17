# loby — VHS tape (charmbracelet/vhs)
# Regenerate: vhs demo.tape  -> demo.gif + demo.webm
#
# Three-beat marketing demo: confirm → compose → commit.
#   1. Verify a real address against Lob's USPS-grade data.
#   2. Dry-run the postcard — see exactly what's about to be sent.
#   3. Commit. Get back a real psc_… id, USPS delivery date, and a
#      clickable PDF preview URL.
#
# Beat 3 is a real Lob postcards.create against the live API. The
# --idempotency-key pins the response so re-recording does NOT mail
# duplicates — Lob returns the cached response for 24h.

Output demo.gif
Output demo.webm

Set Shell "zsh"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 50
Set Width 1500
Set Height 850
Set Padding 40
Set Theme "TokyoNight"
Set TypingSpeed 28ms

Env LOBY_KEYRING_PASSWORD ""

# Prep input files invisibly.
Hide
Type "mkdir -p /tmp/loby-demo && cd /tmp/loby-demo"
Enter
Type `printf '{"name":"Alice Park","address_line1":"185 Berry St","address_city":"San Francisco","address_state":"CA","address_zip":"94107"}' > a.json`
Enter
Type `printf '{"name":"loby Demo","address_line1":"185 Berry St","address_city":"San Francisco","address_state":"CA","address_zip":"94107"}' > from.json`
Enter
Type `printf '<html><body><h1 style="text-align:center;margin-top:120px">Hello from loby.</h1></body></html>' > f.html`
Enter
Type `printf '<html><body><p style="text-align:center;margin-top:140px">Sent from a terminal.<br><br>lobycli.com</p></body></html>' > b.html`
Enter
Type "clear"
Enter
Show

# 1. CONFIRM — real Lob deliverability data.
Type "loby verify us '185 Berry St SF CA 94107' \"
Enter
Type "  -j --select deliverability,components.zip_code_plus_4"
Sleep 400ms
Enter
Sleep 3000ms

Hide
Type "clear"
Enter
Show

# 2. COMPOSE — dry-run, no mail sent.
Type "loby postcards create -n -j \"
Enter
Type "  --to @a.json --from @from.json --size 4x6 \"
Enter
Type "  --front @f.html --back @b.html \"
Enter
Type "  | jq '{method,path,to:.body.to.name,size:.body.size}'"
Sleep 400ms
Enter
Sleep 3500ms

Hide
Type "clear"
Enter
Show

# 3. COMMIT — real postcards.create. Real USPS hand-off.
Type "loby postcards create -j \"
Enter
Type "  --to @a.json --from @from.json --size 4x6 \"
Enter
Type "  --front @f.html --back @b.html \"
Enter
Type "  --idempotency-key loby-demo-v1 \"
Enter
Type "  | jq '{id,status,carrier,expected_delivery_date}'"
Sleep 400ms
Enter
Sleep 4500ms
