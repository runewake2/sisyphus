# Unreleased: graph --full places nodes in linear time

- Date: 2026-10-20
- Bookmark: samw/ai/graph-linear-time
- Components: sisyphus
- Changelog: `sisyphus graph --full` places its nodes in linear time: the edges are bucketed by their source and the ready issues are a heap, instead of a scan of every edge for every node. At 20,000 issues the command drops from 1.4 s to 0.3 s, the same cost as a non-full graph.

## Summary

The performance engineer measured the tracker at 1k, 5k, and 20k issues: every command is linear
except `graph --full`, whose node placement scanned every drawn edge for each placed node and
re-sorted the ready list every iteration (70% of the CPU samples). The placement order is unchanged:
the heap pops the smallest ready name, as the sorted list did.

## Changes

- `buildGraph` buckets `graph.Edges` by their source once, and keeps the ready issues in a min-heap
  (`nameHeap`), so each edge is touched once.
- `BenchmarkGraphFull` covers the path.

## Issues

- Closed [[draw-graph-nodes-in-linear-time]], filed by the performance-engineer role with the
  profile and a validated prototype (1,375 ms to 294 ms at 20k issues, 4.7x).
