# redgla
Scalable Ethereum Read Requester

## Scalable 
What does 'scalability' mean when making read requests to Ethereum clients? A service that makes read requests doesn't consume much computing power. Actual resource consumption comes from Ethereum nodes processing requested blocks, transactions and receipts. Therefore, scale-up or scale-out for the subject requesting reads will not be very effective, and it is necessary to work on the node that will handle the request. However, scaling up of nodes cannot be considered from the requester's point of view, so let's consider a scaling method that divides requests into multiple nodes and then aggregates them.

## Heartbeat
Let's consider the process of making requests to multiple nodes. I need to get receipts for 1000 transactions. Therefore, after sending the request to each of the five nodes in groups of 200, we try to receive them and count them. However, what if a specific node's resources are exhausted or a network failure occurs and the request cannot be properly processed? The result of an incomplete (missing) requested value is an error, and we must resend this request. If one of the five nodes continues to fail, even if four return correct results, the requester is not satisfied. To solve this problem, it maintains a list of healthy nodes by periodically sending low-resource requests to nodes. If we send requests to nodes that are considered to be functioning normally, the possibility that a particular node's operation will be in vain is reduced. Of course, there is a possibility that the node will change to an unhealthy state immediately after sending the request assuming it is normal. However, this will eventually be resolved by sending a request to the newly updated normal node list after the next 'heartbeat interval time'. Since speed is matched to the slowest response, it may be more effective to remove nodes that are too slow from the node list. To help manage the list, it would also be nice to provide response times for requests per node.





