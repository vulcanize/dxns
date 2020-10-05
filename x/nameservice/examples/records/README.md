# DAG

* CID v1 dag-cbor
* https://github.com/ipfs/go-ipld-cbor

## IPFS

```bash
$ ipfs dag put example1.json
bafyreih7un2ntk235wshncebus5emlozdhdixrrv675my5umb6fgdergae

$ ipfs dag get bafyreih7un2ntk235wshncebus5emlozdhdixrrv675my5umb6fgdergae | jq
{
  "attr1": "value1",
  "attr2": "value2",
  "link1": {
    "/": "QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D"
  },
  "link2": {
    "/": "QmP8jTG1m9GSDJLCbeWhVSVgEzCPPwXRdCRuJtQ5Tz9Kc9"
  }
}
```

```bash
$ ipfs dag put example2.json
bafyreib5bguwuvfctzjqu5ulrwboxhdz2fufsyf3zgtawneyybbr3ny42m

$ ipfs dag get bafyreib5bguwuvfctzjqu5ulrwboxhdz2fufsyf3zgtawneyybbr3ny42m | jq
{
  "attr1": "value1",
  "attr2": "value2",
  "links": [
    {
      "/": "QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D"
    },
    {
      "/": "QmP8jTG1m9GSDJLCbeWhVSVgEzCPPwXRdCRuJtQ5Tz9Kc9"
    }
  ]
}

```
## WNS

```bash
$ wire wns record publish -f example1.yml
bafyreih7un2ntk235wshncebus5emlozdhdixrrv675my5umb6fgdergae

$ wire wns record get --id bafyreih7un2ntk235wshncebus5emlozdhdixrrv675my5umb6fgdergae | jq '.[].attributes'
{
  "attr1": "value1",
  "attr2": "value2",
  "link1": {
    "/": "QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D"
  },
  "link2": {
    "/": "QmP8jTG1m9GSDJLCbeWhVSVgEzCPPwXRdCRuJtQ5Tz9Kc9"
  }
}
```

```bash
$ wire wns record publish -f example2.yml
bafyreib5bguwuvfctzjqu5ulrwboxhdz2fufsyf3zgtawneyybbr3ny42m

$ wire wns record get --id bafyreib5bguwuvfctzjqu5ulrwboxhdz2fufsyf3zgtawneyybbr3ny42m | jq '.[].attributes'
{
  "attr1": "value1",
  "attr2": "value2",
  "links": [
    {
      "/": "QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D"
    },
    {
      "/": "QmP8jTG1m9GSDJLCbeWhVSVgEzCPPwXRdCRuJtQ5Tz9Kc9"
    }
  ]
}
```
