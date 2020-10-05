# Auction Module

## Params

* Commit length (default 1 day)
* Reveal length (default 1 day)
* Auction fee
  * Commit fee
  * Reveal fee
* Minimum bid amount

## State

`Auction`:

* ID
* Status (COMMIT, REVEAL, FINISHED)
* CreateTime
* CommitsEndTime
* RevealsEndTime
* CommitFee
* RevealFee
* MinimumBid
* WinnerAddress
* WinnerBidAmount

`Bid`:

* AuctionID
* BidderAddress
* Status (COMMITTED, REVEALED, EXPIRED)
* BidAmount
* AuctionFee
* CommitTime
* RevealTime

### Indexes

* Auctions: `0x00 | auctionID -> Auction`
* Bids: `0x01 | auctionID | bidAddress -> Bid`
* AuctionsByBidder: `0x02 | bidderAddress | auctionID -> <empty>`

## Messages

* CreateAuction
* CommitBid (create or update bid)
* RevealBid

## End Block

* PickWinner
