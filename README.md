# redgla
Scalable Ethereum Read Requester

<img width="728" alt="image" src="https://user-images.githubusercontent.com/72970043/220929311-99e7b45b-5ca1-4933-b241-d30bee6ff987.png">

redgla stores multiple Ethereum client endpoints, and if more than the specified number of requests are received, the requests are equally divided and sent, and the results are combined and returned to the client. See the [Scalable](https://github.com/dbadoy/redgla#scalable).
If scalability is not required, it can also be used to make requests to one alive node of the candidate endpoints instead of split-transmitting to multiple nodes. See the [Stable](https://github.com/dbadoy/redgla#stable).



<b>Table of Contents</b>

- [Install](https://github.com/dbadoy/redgla#install)
- [Usage](https://github.com/dbadoy/redgla#usage)
- [Scalable](https://github.com/dbadoy/redgla#scalable)
- [Stable](https://github.com/dbadoy/redgla#stable)
- [Heartbeat](https://github.com/dbadoy/redgla#heartbeat)

## Install
Required: Go (version 1.16 or later)
```
$ go get -u github.com/dbadoy/redgla
```

## Usage
### Configuration
- Config field details (https://github.com/dbadoy/redgla/blob/main/config.go#L28-L46) <br>

ws, wss as endpoints are not allowed yet. Since the close processing for websockets does not work properly, we'll allow it when the implementation is complete.
```go
// https://github.com/dbadoy/redgla/blob/main/config.go#L16-L19
cfg := redgla.DefaultConfig()

cfg.Endpoints = append(cfg.Endpoints, "http://127.0.0.1:8545", "http://mynode.io")

redgla, err := redgla.New(redgla.DefaultHeartbeatFn, cfg)
if err != nil {
  panic(err)
}
```
### HeartbeatFn
[DefaultHeartbeatFn](https://github.com/dbadoy/redgla/blob/main/beater.go#L22) checks if the chain ID is successfully obtained from the client. A method that checks whether a node is operating normally can be declared and injected externally. However, you must set the timeout through the context.(e.g. implement methods such as determining that a node is an 'abnormal node' if the chain ID is not the mainnet chain ID)

```go
func fn(ctx context.Context, endpoint string) error {
  client, err := ethclient.DialContext(ctx, endpoint)
  if err != nil {
    return err
  }

  chainID, err = client.ChainID(ctx)
  if err != nil {
    return err
  }

  if chainID.Uint64() != 1 {
    return errors.New("invalid chain id")
  }

  return nil
}

//
redgla, err := redgla.New(fn, cfg)
```

### Execute
```go
// Even given a normal endpoint, it takes some time to determine
// it. If a request is made immediately after calling Run(),
// ErrNoAliveNode may occur. We could also implement a
// notification via a channel after the first heartbeat is over.
redgla.Run()

redgla.Stop()
```

## Scalable 
What does 'scalability' mean when making read requests to Ethereum clients? A service that makes read requests doesn't consume much computing power. Actual resource consumption comes from Ethereum nodes processing requested blocks, transactions and receipts. Therefore, scale-up or scale-out for the subject requesting reads will not be very effective, and it is necessary to work on the node that will handle the request. However, scaling up of nodes cannot be considered from the requester's point of view, so let's consider a scaling method that divides requests into multiple nodes and then aggregates them.

## Stable
You may not need scalability. However, specifying one endpoint in the service and making a request to it puts service in a single point of failure(SPOF) state. To solve this, it is necessary to register multiple endpoints and send heartbeats periodically to check whether they are normal nodes. This is a native feature of redgla and for this we just need to make a request to one of the endpoints without splitting the request. 
<br>
<b>This is not the main purpose of redgla. I will not prefer adding features just for this(maybe?).</b>



## Heartbeat
Let's consider the process of making requests to multiple nodes. I need to get receipts for 1000 transactions. Therefore, after sending the request to each of the five nodes in groups of 200, we try to receive them and count them. However, what if a specific node's resources are exhausted or a network failure occurs and the request cannot be properly processed? The result of an incomplete (missing) requested value is an error, and we must resend this request. If one of the five nodes continues to fail, even if four return correct results, the requester is not satisfied. To solve this problem, it maintains a list of healthy nodes by periodically sending low-resource requests to nodes. If we send requests to nodes that are considered to be functioning normally, the possibility that a particular node's operation will be in vain is reduced. Of course, there is a possibility that the node will change to an unhealthy state immediately after sending the request assuming it is normal. However, this will eventually be resolved by sending a request to the newly updated normal node list after the next 'heartbeat interval time'. Since speed is matched to the slowest response, it may be more effective to remove nodes that are too slow from the node list. To help manage the list, it would also be nice to provide response times for requests per node.
