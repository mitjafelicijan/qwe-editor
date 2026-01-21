package main

import "embed"

//go:embed queries/*
var QueriesFS embed.FS

//go:embed content/*
var ContentFS embed.FS
