# TODO

- run inside the cluster
- add integration test and end to end test (with minikube possibly)
- refactor tests
- allow filtering on spec
- revisit error handling
- accept optional delete propagation and other delete options (possibly as a filed in resource identifier)
- save a list of GVRs in a configmap so i don't have to read the resource identifier on delete
- accept regular expressions or wild cards in resource identifier
- add priority for deletion
- run the delete, flag and list operations in parallel using go routines
- save the spec and metadata of deleted objects in a configmap and have that confimap be flagged for deletion (optional) so it gets cleaned next time
