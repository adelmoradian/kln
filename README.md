# kln

[![Go Report Card](https://goreportcard.com/badge/github.com/adelmoradian/kln)](https://goreportcard.com/report/github.com/adelmoradian/kln)

Status: pre-alpha

Kln is utility that helps to keep your cluster clean by finding and deleting
any kubernetes object based a criteria that you provide. The criteria (resource identifier
as I call it) is highly customizable and can include minimum age, metadata, spec
and status.

## Background

Kln is inspired by [kube-janitor](https://github.com/themagicalkarp/kube-janitor)
which is absolutely a great tool! But we needed something that is more flexible
and can handle custom resources (specifically [tekton](https://tekton.dev/)
PipelineRuns and TaskRuns)
