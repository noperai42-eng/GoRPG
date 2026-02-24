# Fight-CLI Test Suite

Comprehensive testing framework for the fight-cli RPG game.

## Test Suite Overview

This test suite includes **3 layers of testing**:

1. **Unit Tests** - Test individual functions and components
2. **Integration Tests** - Test full system workflows
3. **Automated Play Tests** - Test actual gameplay scenarios

---

## Quick Start

### Run All Tests
```bash
./run_all_tests.sh
```

This runs the complete test suite including:
- Unit tests with coverage
- Integration tests
- Auto-play validation
- Performance benchmarks
- Quest system tests
- Combat system tests

### Run Unit Tests Only
```bash
go test -v
```

### Run Integration Tests Only
```bash
./test_integration.sh
```

### Run Specific Test
```bash
go test -run TestCharacterCreation -v
go test -run TestCombatSimulation -v
go test -run TestQuestSystem -v
```

### Run Benchmarks
```bash
go test -bench=. -benchtime=5s
```

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # View in browser
```

---

## Test Files

### `fight-cli_test.go`
**Unit tests** for core game functions

**Tests included:**
- `TestCharacterCreation` - Character generation and initialization
- `TestMonsterGeneration` - Monster spawning and attributes
- `TestItemGeneration` - Item creation with rarity tiers
- `TestHealthPotion` - Consumable items
- `TestSkills` - All 9 player skills validation
- `TestElementalDamage` - Resistance/weakness system
- `TestLevelUp` - Experience and leveling mechanics
- `TestQuestSystem` - Quest progression and completion
- `TestAIDecisionMaking` - Auto-play AI logic
- `TestSaveLoad` - Game state persistence
- `TestResourceHarvesting` - Resource gathering
- `TestStatusEffects` - DoT, buffs, debuffs
- `TestCombatSimulation` - Full combat encounter
- `TestAutoPlayMode` - Automated combat (5 fights)
- `TestBackwardCompatibility` - Old save file loading

**Benchmarks:**
- `BenchmarkCharacterGeneration` - Character creation performance
- `BenchmarkMonsterGeneration` - Monster spawning performance
- `BenchmarkDamageCalculation` - Combat calculation speed

### `test_integration.sh`
**Integration tests** for full system workflows

**Tests included:**
1. New character creation
2. Character stats display
3. Resource harvesting
4. Location discovery
5. Quest log display
6. Save and load functionality
7. Multiple characters management
8. Backward compatibility with old saves
9. Single combat initialization
10. Auto-play mode initialization

### `run_all_tests.sh`
**Comprehensive test runner** that executes:
1. Go unit tests with coverage
2. Performance benchmarks
3. Integration tests
4. Auto-play mode test (10 seconds)
5. Skills validation
6. Elemental damage system
7. Quest system
8. Memory/race condition check
9. Code coverage analysis
10. Final summary report

---

## Test Coverage

Current test coverage includes:

### Core Systems
- ✅ Character generation and stats
- ✅ Monster generation with skills
- ✅ Item generation (equipment + consumables)
- ✅ Combat system with skills and status effects
- ✅ Elemental damage and resistances
- ✅ Level-up progression
- ✅ Experience and skill learning

### Gameplay Features
- ✅ Quest system with progression gates
- ✅ Save/load functionality
- ✅ Resource harvesting
- ✅ Location discovery
- ✅ Multiple character support
- ✅ Inventory and equipment management
- ✅ Status effects (poison, burn, stun, regen, buffs)

### Auto-Play
- ✅ AI decision making
- ✅ Auto-combat simulation
- ✅ Quest progress tracking
- ✅ Resource management

### Edge Cases
- ✅ Backward compatibility with old saves
- ✅ Nil map initialization
- ✅ Quest arrays for existing characters
- ✅ Long combat encounters

---

## Example Output

### Successful Test Run
```
╔═══════════════════════════════════════════════════╗
║                                                   ║
║        Fight-CLI Comprehensive Test Suite        ║
║                                                   ║
╚═══════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1. Running Go Unit Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

=== RUN   TestCharacterCreation
    fight-cli_test.go:67: ✓ Character created: TestHero (Level 1, HP: 6, MP: 26, SP: 26)
--- PASS: TestCharacterCreation (0.00s)

=== RUN   TestQuestSystem
    fight-cli_test.go:196: ✓ Quest system: 5 available quests, 1 active
    fight-cli_test.go:205: ✓ Quest progress check complete
--- PASS: TestQuestSystem (0.00s)

✓ Unit tests passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  3. Running Integration Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ PASS: Character creation saves game state
✓ PASS: Character 'TestHero' found in save file
✓ PASS: Quest system initialized

╔═══════════════════════════════════════╗
║                                       ║
║   🎉  ALL TESTS PASSED! 🎉           ║
║                                       ║
╚═══════════════════════════════════════╝
```

---

## Writing New Tests

### Adding a Unit Test

1. Open `fight-cli_test.go`
2. Add a new test function:

```go
func TestMyNewFeature(t *testing.T) {
    rand.Seed(time.Now().UnixNano())

    // Setup
    gameState := createTestGameState()
    char := createTestCharacter("TestChar", 1)

    // Execute
    result := myNewFeature(&char, &gameState)

    // Assert
    if result != expectedValue {
        t.Errorf("Expected %v, got %v", expectedValue, result)
    }

    t.Logf("✓ My feature test passed")
}
```

### Adding an Integration Test

1. Open `test_integration.sh`
2. Add a new test function:

```bash
test_my_feature() {
    print_header "Test N: My Feature"

    rm -f test_gamestate.json

    # Run game with test input
    output=$(echo -e "commands\nhere\nexit" | timeout 5 ./fight-cli 2>&1)

    # Check output
    if echo "$output" | grep -q "Expected Output"; then
        print_result 0 "Feature works correctly"
    else
        print_result 1 "Feature failed"
    fi

    rm -f test_gamestate.json
}
```

3. Call it from `main()`:

```bash
test_my_feature
```

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd fight-cli
          ./run_all_tests.sh
```

---

## Test Data

### Helper Functions

**`createTestCharacter(name, level)`**
- Creates a fully initialized character for testing
- Includes starting items and locations

**`createTestGameState()`**
- Creates a complete game state
- Initializes locations, monsters, and quests

### Test Save Files

Tests automatically create and clean up temporary save files:
- `test_gamestate.json` - General testing
- `test_save_*.json` - Save/load testing
- `test_autoplay_save.json` - Auto-play testing

---

## Performance Testing

### Benchmarks

Run with increasing time for more accurate results:

```bash
# Quick benchmark (1 second)
go test -bench=. -benchtime=1s

# Detailed benchmark (10 seconds)
go test -bench=. -benchtime=10s

# With memory allocation stats
go test -bench=. -benchmem
```

### Example Output
```
BenchmarkCharacterGeneration-8      50000    28456 ns/op
BenchmarkMonsterGeneration-8        30000    45123 ns/op
BenchmarkDamageCalculation-8     1000000     1234 ns/op
```

---

## Debugging Tests

### Run Test with Verbose Output
```bash
go test -v -run TestCombatSimulation
```

### Run Test Until Failure
```bash
go test -count=100 -run TestCombatSimulation
```

### Debug Specific Issue
```bash
# Print only failed tests
go test -v | grep -E '(FAIL|PASS:)'

# Run with race detector
go test -race -v

# Run with CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

---

## Test Maintenance

### Cleaning Test Files
```bash
# Remove all test artifacts
rm -f test_*.json gamestate.json coverage.* *.prof
```

### Updating Tests After Code Changes

When you add new features:
1. Add unit tests to `fight-cli_test.go`
2. Add integration tests to `test_integration.sh`
3. Update this README
4. Run full test suite: `./run_all_tests.sh`

---

## Test Philosophy

Our testing approach:

1. **Fast Unit Tests** - Test individual functions in isolation
2. **Comprehensive Integration** - Test full workflows
3. **Realistic Scenarios** - Test actual gameplay
4. **Backward Compatibility** - Test old save files
5. **Performance Monitoring** - Track speed over time

---

## Known Limitations

- Integration tests use timeouts for interactive portions
- Auto-play tests run for limited duration
- Some tests are non-deterministic due to RNG
- Combat simulation may timeout on very unbalanced fights

---

## Contributing

When submitting changes:
1. Add tests for new features
2. Ensure all existing tests pass
3. Maintain >70% code coverage
4. Update test documentation

---

## Support

For issues with tests:
- Check that `fight-cli` compiles: `go build fight-cli.go`
- Verify test files are executable: `chmod +x *.sh`
- Clean test files: `rm -f test_*.json gamestate.json`
- Run tests individually to isolate issues

---

**Happy Testing! 🎮✨**
