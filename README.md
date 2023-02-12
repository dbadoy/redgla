# redgla
Scalable Ethereum Read Requesters

## Scalable 
What does 'scalability' mean when making read requests to Ethereum clients? A service that makes read requests doesn't consume much computing power. Actual resource consumption comes from Ethereum nodes processing requested blocks, transactions and receipts. Therefore, scale-up or scale-out for the subject requesting reads will not be very effective, and it is necessary to work on the node that will handle the request. However, scaling up of nodes cannot be considered from the requester's point of view, so let's consider a scaling method that divides requests into multiple nodes and then aggregates them.





