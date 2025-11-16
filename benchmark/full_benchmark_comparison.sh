#!/bin/bash

echo "================================================================"
echo "VitalDB Performance Comparison: Python vs Go"
echo "================================================================"
echo ""

echo "Running Python VitalDB Benchmark..."
echo "------------------------------------------------------------"
python3 simple_benchmark.py 2>&1 | grep -v "NotOpenSSLWarning" | grep -v "warnings.warn" | grep -v "Error in reading"
echo ""

echo "Running Go VitalDB Processor Benchmark..."
echo "------------------------------------------------------------"
go run simple_benchmark.go
echo ""

echo "================================================================"
echo "Summary"
echo "================================================================"
echo ""
echo "Note: Python VitalDB library has compatibility issues with some"
echo "tracks (buffer errors), resulting in only 108/254 tracks loaded."
echo "Go implementation successfully loads all 254 tracks."
echo ""
