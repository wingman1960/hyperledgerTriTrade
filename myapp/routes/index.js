var express = require('express');
var router = express.Router();
var query = require(__dirname + '/../query.js')
var invoke = require(__dirname + '/../invoke.js')
var path = require('path');

/* GET home page. */
router.get('/', function(req, res, next) {
  res.render('index', { title: 'Express' });
});

var options = {
    wallet_path: path.join(__dirname, './../creds'),
    user_id: 'PeerAdmin',
    channel_id: 'mychannel',
    chaincode_id: 'marblesTrade7',
    peer_url: 'grpc://localhost:7051',
    event_url: 'grpc://localhost:7053',
    orderer_url: 'grpc://localhost:7050',
    network_url: 'grpc://localhost:7051'
};




invoke = new invoke(options);
query = new query(options);

router.get('/query', function(req, res, next) {
  console.log(req.query.args)
  fcn =  req.query.fcn
  args =  req.query.args
  args = args.split(",")
  query.query(fcn, args).then((resHfc)=> {
    res.send(resHfc)
  });

});



router.post('/invoke', function(req, res, next) {
// var fcn = 'initMarble';
// var args = ["marble5","red","50","tom"];
  var fcn = req.body.fcn;
  var args = req.body.args;
  console.log(req.body)
  invoke.invoke(fcn, args).then((resHfc)=> {
    res.send(resHfc)
  });
});

module.exports = router;

router.post('/delete', function(req, res, next) {
  var fcn = req.body.fcn;
  var args = req.body.args;
  console.log(req.body)
  invoke.invoke(fcn, args)
});

module.exports = router;

// create: initMarble
// transfer: transferMarble
// delete: delete
// queryByRange: getMarblesByRange
// queryMarble: queryMarbles