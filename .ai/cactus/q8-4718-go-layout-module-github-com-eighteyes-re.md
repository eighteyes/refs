# q8 — Go layout: module github.com/eighteyes/refs, package main at repo root

status: answered
act: steer
kind: choice
thread: go-port
agent: 72f7858d-f45b-4670-a659-3b5441573403
cwd: .
asked at: 2026-11-07T05:56:20.787282+00:00

## Context

**Proceeding with root-main. Retap to redirect, or edit the module path if eighteyes is wrong.**

- hooks.json calls refs hook on PATH; if refs is missing it prints an install hint instead of erroring
- bin/ and the python file are removed
- local install: go install .

— slugs —
touched: CLAUDE.md

## Files

- hooks/hooks.json

## Options

- root-main — main.go + split files at root, go install github.com/eighteyes/refs@latest
+ shortest install line; plugin files sit beside it  (doing)
- cmd-refs — cmd/refs/main.go, logic in internal/
- more dirs for a ~400-line tool

## Answer

root-main
answered at: 2026-11-07T06:16:18.183147+00:00
