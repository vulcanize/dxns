# Multisig

First, read https://hub.cosmos.network/master/resources/gaiacli.html#multisig-transactions for required background on how multisigs work in cosmos-sdk.

## Example - Send Coins

Create the p1 key:

```bash
$ wnscli keys add p1-full
Enter a passphrase to encrypt your key to disk:
Repeat the passphrase:
{
  "name": "p1-full",
  "type": "local",
  "address": "cosmos1dxdcnzqrqtpfapq9ackmg4smee9npcmycardk6",
  "pubkey": "cosmospub1addwnpepqdxa4x6lmynugg2f8nzlkccjsnsgq3klg39plmgfemuhgmxucfk723efuyy",
  "mnemonic": "harbor remove two orphan tail good social egg ring fence elder lady wait couch cry speed unfold party cluster remember roast undo flavor culture"
}
```

Create another key with just p1's public key:

```bash
$ wnscli keys add p1 --pubkey $(wnscli keys show --pubkey p1-full)
```

Create the p2 key:

```bash
$ wnscli keys add p2-full
Enter a passphrase to encrypt your key to disk:
Repeat the passphrase:
{
  "name": "p2-full",
  "type": "local",
  "address": "cosmos1x587alzhuupvxrz9zfn7lg57w7q07y4upead6u",
  "pubkey": "cosmospub1addwnpepqg4e87juau57fkdp2xxe2w8rjhryjhzh66zn528rdn783uk5a9v9vk0rh8q",
  "mnemonic": "impact soccer portion enlist hope strategy holiday girl column obtain aunt useless lyrics soccer draw mercy stable coffee normal drastic claim chaos theme upset"
}
```

Create another key with just p2's public key:

```bash
$ wnscli keys add p2 --pubkey $(wnscli keys show --pubkey p2-full)
```

Create the p3 key:

```bash
$ wnscli keys add p3-full
Enter a passphrase to encrypt your key to disk:
Repeat the passphrase:
{
  "name": "p3-full",
  "type": "local",
  "address": "cosmos1nnajer7gg5stfxfsa3sm3pv9qznwxk8xhmu808",
  "pubkey": "cosmospub1addwnpepqdgz7kp02mefcr53h3cgt9rp8vgfs6at0x0eepdga42s8ce96d55wj0kk2s",
  "mnemonic": "rebuild predict tennis exotic squeeze proof hill theme rigid twist cannon river high index ocean fury want diagram breeze expand forget tongue kiwi valid"
}
```

Create another key with just p3's public key:

```bash
$ wnscli keys add p3 --pubkey $(wnscli keys show --pubkey p3-full)
```

Create multisig public key:

```bash
$ wnscli keys add p1p2p3 --multisig-threshold=2 --multisig=p1,p2,p3
Key "p1p2p3" saved to disk.
```

```bash
$ wnscli keys list
[
  {
    "name": "p1-full",
    "type": "local",
    "address": "cosmos1dxdcnzqrqtpfapq9ackmg4smee9npcmycardk6",
    "pubkey": "cosmospub1addwnpepqdxa4x6lmynugg2f8nzlkccjsnsgq3klg39plmgfemuhgmxucfk723efuyy"
  },
  {
    "name": "p1",
    "type": "offline",
    "address": "cosmos1dxdcnzqrqtpfapq9ackmg4smee9npcmycardk6",
    "pubkey": "cosmospub1addwnpepqdxa4x6lmynugg2f8nzlkccjsnsgq3klg39plmgfemuhgmxucfk723efuyy"
  },
  {
    "name": "p1p2p3",
    "type": "multi",
    "address": "cosmos1kgqrtk622v3ce08qaffnkgk4gewqkzexr98lcw",
    "pubkey": "cosmospub1ytql0csgqgfzd666axrjzq3tj0a9emefunv6z5vdj5uw89wxf9w904598g5wxm8u0redf62c2cfzd666axrjzq6dm2d4lkf8css5j0x9ld339p8qsprd73z2rlksnnhew3kdesndu5fzd666axrjzq6s9avz74hjns8fr0rssk2xzwcsnp46k7vlnjz63m24q03jt5mfguj5pg2j"
  },
  {
    "name": "p2-full",
    "type": "local",
    "address": "cosmos1x587alzhuupvxrz9zfn7lg57w7q07y4upead6u",
    "pubkey": "cosmospub1addwnpepqg4e87juau57fkdp2xxe2w8rjhryjhzh66zn528rdn783uk5a9v9vk0rh8q"
  },
  {
    "name": "p2",
    "type": "offline",
    "address": "cosmos1x587alzhuupvxrz9zfn7lg57w7q07y4upead6u",
    "pubkey": "cosmospub1addwnpepqg4e87juau57fkdp2xxe2w8rjhryjhzh66zn528rdn783uk5a9v9vk0rh8q"
  },
  {
    "name": "p3-full",
    "type": "local",
    "address": "cosmos1nnajer7gg5stfxfsa3sm3pv9qznwxk8xhmu808",
    "pubkey": "cosmospub1addwnpepqdgz7kp02mefcr53h3cgt9rp8vgfs6at0x0eepdga42s8ce96d55wj0kk2s"
  },
  {
    "name": "p3",
    "type": "offline",
    "address": "cosmos1nnajer7gg5stfxfsa3sm3pv9qznwxk8xhmu808",
    "pubkey": "cosmospub1addwnpepqdgz7kp02mefcr53h3cgt9rp8vgfs6at0x0eepdga42s8ce96d55wj0kk2s"
  },
  {
    "name": "root",
    "type": "local",
    "address": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
    "pubkey": "cosmospub1addwnpepqt4dz2kjn3fjxe9j776k2kpynxzqln695503v4prs5rjjc05ma3dsljnlpn"
  }
]
```

Move funds into the multisig account:

```bash
$ wnscli tx send $(wnscli keys show -a root) $(wnscli keys show -a p1p2p3) 100000000uwire --from root
```

Check funds in multisig account:

```bash
$ wnscli query account $(wnscli keys show -a p1p2p3)
{
  "type": "cosmos-sdk/Account",
  "value": {
    "address": "cosmos1kgqrtk622v3ce08qaffnkgk4gewqkzexr98lcw",
    "coins": [
      {
        "denom": "uwire",
        "amount": "100000000"
      }
    ],
    "public_key": null,
    "account_number": "8",
    "sequence": "0"
  }
}
```

Initiate tx from multisig address:

```bash
$ wnscli tx send $(wnscli keys show -a p1p2p3) cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned 1000000uwire \
  --generate-only > unsignedTx.json

$ cat unsignedTx.json | jq
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [
      {
        "type": "cosmos-sdk/MsgSend",
        "value": {
          "from_address": "cosmos1kgqrtk622v3ce08qaffnkgk4gewqkzexr98lcw",
          "to_address": "cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned",
          "amount": [
            {
              "denom": "uwire",
              "amount": "1000000"
            }
          ]
        }
      }
    ],
    "fee": {
      "amount": [],
      "gas": "200000"
    },
    "signatures": null,
    "memo": ""
  }
}
```

```bash
$ wnscli tx sign \
  unsignedTx.json \
  --multisig=$(wnscli keys show -a p1p2p3) \
  --from=p1-full \
  --output-document=p1signature.json

$ cat p1signature.json| jq                                                                                                                                                               3s
{
  "pub_key": {
    "type": "tendermint/PubKeySecp256k1",
    "value": "A03am1/ZJ8QhSTzF+2MShOCARt9ESh/tCc75dGzcwm3l"
  },
  "signature": "q7oyj2Pjwz5vnZtlRruKTdZXMixKhJ9OE35eDyXH7AJTRUzDw6KZ/0vnPEyMxGS6i6bD+Q30snhv1UCrSf+rJQ=="
}
```

P2 signs:

```bash
$ wnscli tx sign \
  unsignedTx.json \
  --multisig=$(wnscli keys show -a p1p2p3) \
  --from=p2-full \
  --output-document=p2signature.json

$ cat p2signature.json| jq
{
  "pub_key": {
    "type": "tendermint/PubKeySecp256k1",
    "value": "AiuT+lzvKeTZoVGNlTjjlcZJXFfWhToo42z8ePLU6VhW"
  },
  "signature": "p0Lk4HGZ0uee4FeOOJ5M2ciU7lOQAyHrRARia5iYgnw1zZMQxkH+qZWnUiUdpeb18f2Y4IJsbcdM6dRFC7EGBQ=="
}
```

Generate multisig tx (2/3 met):

```bash
$ wnscli tx multisign \
  unsignedTx.json \
  p1p2p3 \
  p1signature.json p2signature.json > signedTx.json

$ cat signedTx.json
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [
      {
        "type": "cosmos-sdk/MsgSend",
        "value": {
          "from_address": "cosmos1kgqrtk622v3ce08qaffnkgk4gewqkzexr98lcw",
          "to_address": "cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned",
          "amount": [
            {
              "denom": "uwire",
              "amount": "1000000"
            }
          ]
        }
      }
    ],
    "fee": {
      "amount": [],
      "gas": "200000"
    },
    "signatures": [
      {
        "pub_key": {
          "type": "tendermint/PubKeyMultisigThreshold",
          "value": {
            "threshold": "2",
            "pubkeys": [
              {
                "type": "tendermint/PubKeySecp256k1",
                "value": "AiuT+lzvKeTZoVGNlTjjlcZJXFfWhToo42z8ePLU6VhW"
              },
              {
                "type": "tendermint/PubKeySecp256k1",
                "value": "A03am1/ZJ8QhSTzF+2MShOCARt9ESh/tCc75dGzcwm3l"
              },
              {
                "type": "tendermint/PubKeySecp256k1",
                "value": "A1AvWC9W8pwOkbxwhZRhOxCYa6t5n5yFqO1VA+Ml02lH"
              }
            ]
          }
        },
        "signature": "CgUIAxIBwBJAp0Lk4HGZ0uee4FeOOJ5M2ciU7lOQAyHrRARia5iYgnw1zZMQxkH+qZWnUiUdpeb18f2Y4IJsbcdM6dRFC7EGBRJAq7oyj2Pjwz5vnZtlRruKTdZXMixKhJ9OE35eDyXH7AJTRUzDw6KZ/0vnPEyMxGS6i6bD+Q30snhv1UCrSf+rJQ=="
      }
    ],
    "memo": ""
  }
}
```

Broadcast tx:

```bash
$ wnscli tx broadcast signedTx.json
```

Check target account balance:

```bash
$ wnscli query account cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned
```

## Example - Reserve Name

```bash
$ wnscli tx nameservice reserve-name dxos \
  --from $(wnscli keys show -a p1p2p3) \
  --generate-only > unsignedTx.json

$ wnscli tx sign \
  unsignedTx.json \
  --multisig=$(wnscli keys show -a p1p2p3) \
  --from=p1-full \
  --output-document=p1signature.json

$ wnscli tx sign \
  unsignedTx.json \
  --multisig=$(wnscli keys show -a p1p2p3) \
  --from=p2-full \
  --output-document=p2signature.json

$ wnscli tx multisign \
  unsignedTx.json \
  p1p2p3 \
  p1signature.json p2signature.json > signedTx.json

$ wnscli tx broadcast signedTx.json

$ wnscli query nameservice whois dxos
{
  "ownerPublicKey": "IsH34ggCEibrWumHIQIrk/pc7ynk2aFRjZU445XGSVxX1oU6KONs/Hjy1OlYVhIm61rphyEDTdqbX9knxCFJPMX7YxKE4IBG30RKH+0Jzvl0bNzCbeUSJuta6YchA1AvWC9W8pwOkbxwhZRhOxCYa6t5n5yFqO1VA+Ml02lH",
  "ownerAddress": "cosmos1kgqrtk622v3ce08qaffnkgk4gewqkzexr98lcw",
  "height": 875
}
```
