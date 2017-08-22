# Description
This an application inspired by the hyperledger fabric v0.6 marble trading tutorial. This application handles the scenario when 3 people want to trading away their marbles with the fact that their wish can only be satisfy when the 3 people exchange their marbles at the same time in a circular way, ie, the successful exchange of marbles could not happen within one single trade. All trading logic happens in chaincode @ hyperledger fabric v1.0. 
## Future application:
This will greatly increase the liquidity of one asset, without limitaion to only digital currency, provided that there is a mechanism binding that physical asset into its digital representation
# Requirement 
- docker
- docker-compose
- npm 6.x

# Usage
## Start local hyperledger fabric network
```
cd myapp
./startFabric
```
## Start Node server
```
npm install
node app.js
```
# Chaincode interface
### initMarble(stub, args)
create a new marble
### transferMarble(stub, args)
change owner of a specific marble
### transferMarblesBasedOnColor(stub, args)
transfer all marbles of a certain color
### delete(stub, args)
delete a marble
### readMarble(stub, args)
read a marble
### queryMarblesByOwner(stub, args)
find marbles for owner X using rich query
### queryMarbles(stub, args)
find marbles based on an ad hoc rich query
### getHistoryForMarble(stub, args)
get history of values for a marble
### getMarblesByRange(stub, args)
get marbles based on range query
### openTrade(stub, args)
open a new marble trade
### readOpenTrade(stub, args)
read marble trades
### removeOpenTrade(stub, args)
remove marble trade
### swapMarble(stub, args)
swap two marbles between owner depending on input color
### matchTrade(stub, args)
match the open trades


# Limitation
Currently the id of each AnOpenTrade is identified by timestamp generated on nodes. This post the the difficult of consistant timestamp across each node. The current solution is to round off the timestamp to nearest second so as to decrease the posibility of inconsistancy as compare to rounding off to miniSecond.