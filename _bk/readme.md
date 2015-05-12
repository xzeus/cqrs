
************************************************************
*
*	CQRS	-	Command Query Responsiblity Segregation
*	ES		-	Event Sourcing
*
************************************************************


TODO:
------------------------------------------------------------

- Describe and then handle idempotent vs non-idempotent actions
	- At most once, At least once
	- Act then persist, Persist then act, Idempotent
		https://www.parleys.com/play/545b4d2ae4b08fbc478d8bf9/chapter34/about
	- How do we handle resume on failure?

- Enable persistence scheme per domain (micro-service)
	- Multi-persist, Multi-mem, Delayed-ship, Local-persist, Local-mem

------------------------------------------------------------


Aggregate Load Optimization?

Read snapshot state and concurrently try to process it as is and with the additional immediately consistent lookup?





Command
- A data structure that implies an imperitive call to action
- Encapsulates all information needed to execute the requested action
- Can be denied, a request not a demand

Command Handler
- Validates a command against it's aggregate state
- Creates a resulting event, either success or error
- Can produce only one event


An event handler must represent all needed state in it's aggregate


All storage problems must respect key space distribution


Events must be addressable by [ Domain Id | Aggregate Id | Version ] within Tenancy boundaries forming a 128 bit UUID






Introduction / Networking

Understand your role and what kinds of things you're interested in and the motivations for those things (money, info, passion, et)

Explain what we are doing





***************************

CQRS

Separation of Read and Write operations

Can optimize each side for it's sweet spot

Partitioning
- Read/Write Operations

Domain Driven Design
- Application
- Domain
- Aggregate



***************************

Event Sourcing

Immutable events

Rich audit log

Challenges:

Requires more storage (unless you cound log management)

Cannot query data store beyond aggregate



***************************

Distributed Key Value Store

Key space partitioning

Dynamo, Tablets

Hash Distribution - Linear probing, etc




***************************

Message Queues

Latency vs Reliablity




***************************

Microservices

Boundary around a highly cohesive state machine

Domain Driven Design - Bounded Context

Shared Nothing Stateless Architecture

Value of Conventions / common design and libraries

Very fast provisioning required

Different set of management tools
- Are management and monitoring tools also microservices?





***************************


apiserver
=========

var inc, outc chan[]byte
...
for data := range inc {
ndata := modify(data)
if ok := outc <-ndata; !ok {
close(inc, cerror(outc))
break
}
}
close(outc, cerror(inc))



Region - Roughly data center boundary

32 bit

Cluster - Set of connected physical hosts

32 bit

Host - Individual server hosting processes (MAC?)

64 bit

Service - Isolated process running on host (domain)

64 bit

Instance - For K/V store

64 bit

Type

32 bit

1111 0000 0000 0000 MASK = Subsystem Scope
[ Region | Cluster | Host | Service ]
- Each layer up should be much lower volume but higher priority
0000 1000 0000 0000 MASK = Command/Event


Version

32 bit



Can network partitions be reconsiled by splitting the domain along region and/or cluster lines and utilizing a merge domain to handle merging disparate region domains?


For non-idempotent side-effect operations can we specify a risk threshold on T(ime) and/or P(robability of failure) that must be exceeded before retry is attempted in partial failure cases?


Propagate confidence as a spectrum instead of a bool

For failure consider the average time between recent heartbeats to accomodate changing network conditions


***************************



***************************