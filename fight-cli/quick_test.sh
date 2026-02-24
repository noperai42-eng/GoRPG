#!/bin/bash
# Quick test script - runs a subset of tests quickly

echo "🚀 Running quick tests..."
echo ""

# Run a few key tests
go test -run "TestCharacterCreation|TestSkills|TestQuestSystem|TestSaveLoad" -v

echo ""
echo "✅ Quick tests complete!"
echo ""
echo "For full test suite, run: ./run_all_tests.sh"
