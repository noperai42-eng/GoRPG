#!/bin/bash

# Integration Test Suite for fight-cli
# Tests various gameplay scenarios with simulated user input

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Print test header
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Print test result
print_result() {
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS:${NC} $2"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAIL:${NC} $2"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Clean up test files
cleanup() {
    rm -f test_gamestate.json
    rm -f test_save_*.json
    echo -e "${YELLOW}Cleaned up test files${NC}"
}

# Build the game
build_game() {
    print_header "Building Game"

    if go build -o fight-cli fight-cli.go; then
        print_result 0 "Game compiled successfully"
    else
        print_result 1 "Game compilation failed"
        exit 1
    fi
}

# Test 1: New character creation
test_new_character() {
    print_header "Test 1: New Character Creation"

    rm -f test_gamestate.json

    # Create new character and exit
    echo -e "0\nTestHero\nexit" | timeout 5 ./fight-cli > /dev/null 2>&1

    if [ -f "test_gamestate.json" ]; then
        print_result 0 "Character creation saves game state"

        # Check if character exists in save
        if grep -q "TestHero" test_gamestate.json; then
            print_result 0 "Character 'TestHero' found in save file"
        else
            print_result 1 "Character 'TestHero' not found in save file"
        fi

        # Check for quest system
        if grep -q "ActiveQuests" test_gamestate.json; then
            print_result 0 "Quest system initialized"
        else
            print_result 1 "Quest system not initialized"
        fi
    else
        print_result 1 "Game state not saved"
    fi

    rm -f test_gamestate.json
}

# Test 2: Character stats display
test_character_stats() {
    print_header "Test 2: Character Stats Display"

    rm -f test_gamestate.json

    # Create character and view stats
    output=$(echo -e "0\nStatsTest\n5\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "Player Stats"; then
        print_result 0 "Stats display shows header"
    else
        print_result 1 "Stats display missing header"
    fi

    if echo "$output" | grep -q "Level:"; then
        print_result 0 "Stats display shows level"
    else
        print_result 1 "Stats display missing level"
    fi

    if echo "$output" | grep -q "Experience:"; then
        print_result 0 "Stats display shows experience"
    else
        print_result 1 "Stats display missing experience"
    fi

    rm -f test_gamestate.json
}

# Test 3: Resource harvesting
test_resource_harvest() {
    print_header "Test 3: Resource Harvesting"

    rm -f test_gamestate.json

    # Create character and harvest lumber
    output=$(echo -e "0\nHarvester\n1\nLumber\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "Resource Harvested"; then
        print_result 0 "Resource harvesting works"
    else
        print_result 1 "Resource harvesting failed"
    fi

    if echo "$output" | grep -q "Lumber"; then
        print_result 0 "Lumber harvested"
    else
        print_result 1 "Lumber not harvested"
    fi

    rm -f test_gamestate.json
}

# Test 4: Location discovery
test_location_discovery() {
    print_header "Test 4: Location Discovery"

    rm -f test_gamestate.json

    # Create character and search for location
    output=$(echo -e "0\nExplorer\n2\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "TotalWeight"; then
        print_result 0 "Location search executed"
    else
        print_result 1 "Location search failed"
    fi

    rm -f test_gamestate.json
}

# Test 5: Quest log display
test_quest_log() {
    print_header "Test 5: Quest Log"

    rm -f test_gamestate.json

    # Create character and view quest log
    output=$(echo -e "0\nQuestTester\n9\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "QUEST LOG"; then
        print_result 0 "Quest log displays"
    else
        print_result 1 "Quest log not displayed"
    fi

    if echo "$output" | grep -q "ACTIVE QUESTS"; then
        print_result 0 "Active quests section shown"
    else
        print_result 1 "Active quests section missing"
    fi

    if echo "$output" | grep -q "The First Trial"; then
        print_result 0 "Starting quest active"
    else
        print_result 1 "Starting quest not found"
    fi

    rm -f test_gamestate.json
}

# Test 6: Save and load
test_save_load() {
    print_header "Test 6: Save and Load"

    rm -f test_save_*.json

    # Create character and save
    echo -e "0\nSaveTester\n5\nexit" | timeout 5 ./fight-cli > /dev/null 2>&1

    # Check save file exists
    if [ -f "gamestate.json" ]; then
        cp gamestate.json test_save_original.json
        print_result 0 "Game saved successfully"

        # Load the save
        output=$(echo -e "6\nSaveTester\n5\nexit" | timeout 5 ./fight-cli 2>&1)

        if echo "$output" | grep -q "SaveTester"; then
            print_result 0 "Character loaded successfully"
        else
            print_result 1 "Character not loaded"
        fi
    else
        print_result 1 "Game not saved"
    fi

    rm -f test_save_*.json gamestate.json
}

# Test 7: Multiple characters
test_multiple_characters() {
    print_header "Test 7: Multiple Characters"

    rm -f test_gamestate.json

    # Create first character
    echo -e "0\nHero1\n5\nexit" | timeout 5 ./fight-cli > /dev/null 2>&1

    # Create second character
    echo -e "0\nHero2\n5\nexit" | timeout 5 ./fight-cli > /dev/null 2>&1

    # Load first character
    output=$(echo -e "6\nHero1\n5\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "Hero1"; then
        print_result 0 "First character loadable"
    else
        print_result 1 "First character not loadable"
    fi

    # Load second character
    output=$(echo -e "6\nHero2\n5\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "Hero2"; then
        print_result 0 "Second character loadable"
    else
        print_result 1 "Second character not loadable"
    fi

    rm -f test_gamestate.json gamestate.json
}

# Test 8: Backward compatibility (old save format)
test_backward_compatibility() {
    print_header "Test 8: Backward Compatibility"

    # Create an old-style save file (without quest fields)
    cat > gamestate.json << 'EOF'
{
  "CharactersMap": {
    "OldHero": {
      "Name": "OldHero",
      "Level": 10,
      "Experience": 1000,
      "ExpSinceLevel": 0,
      "HitpointsTotal": 50,
      "HitpointsNatural": 50,
      "HitpointsRemaining": 50,
      "ManaTotal": 50,
      "ManaNatural": 50,
      "ManaRemaining": 50,
      "StaminaTotal": 50,
      "StaminaNatural": 50,
      "StaminaRemaining": 50,
      "AttackRolls": 2,
      "DefenseRolls": 2,
      "StatsMod": {
        "AttackMod": 0,
        "DefenseMod": 0,
        "HitPointMod": 0
      },
      "Resurrections": 0,
      "Inventory": [],
      "EquipmentMap": {},
      "ResourceStorageMap": {},
      "KnownLocations": ["Home", "Training Hall"],
      "builtBuildings": [],
      "LearnedSkills": [],
      "StatusEffects": [],
      "Resistances": {}
    }
  },
  "GameLocations": {},
  "AvailableQuests": null
}
EOF

    # Try to load the old save
    output=$(echo -e "6\nOldHero\n5\nexit" | timeout 5 ./fight-cli 2>&1)

    if echo "$output" | grep -q "OldHero"; then
        print_result 0 "Old save file loads without error"
    else
        print_result 1 "Old save file failed to load"
    fi

    if [ -f "gamestate.json" ]; then
        # Check if quest system was added
        if grep -q "ActiveQuests" gamestate.json; then
            print_result 0 "Quest system auto-initialized for old save"
        else
            print_result 1 "Quest system not initialized"
        fi
    fi

    rm -f gamestate.json
}

# Test 9: Hunt combat (single fight)
test_single_combat() {
    print_header "Test 9: Single Combat"

    rm -f test_gamestate.json

    # Create character, then try to hunt (will fail due to interaction, but tests system)
    output=$(echo -e "0\nWarrior\n3\nTraining Hall\n" | timeout 3 ./fight-cli 2>&1 || true)

    if echo "$output" | grep -q "How Many Hunts"; then
        print_result 0 "Combat menu accessible"
    else
        print_result 1 "Combat menu not accessible"
    fi

    rm -f test_gamestate.json
}

# Test 10: Auto-play mode initialization
test_autoplay_init() {
    print_header "Test 10: Auto-Play Mode Initialization"

    rm -f test_gamestate.json

    # Create character and try auto-play (will timeout quickly)
    output=$(echo -e "0\nAutoTester\n8\n" | timeout 2 ./fight-cli 2>&1 || true)

    if echo "$output" | grep -q "AUTO-PLAY MODE"; then
        print_result 0 "Auto-play mode accessible"
    else
        print_result 1 "Auto-play mode not accessible"
    fi

    if echo "$output" | grep -q "speed"; then
        print_result 0 "Speed selection shown"
    else
        print_result 1 "Speed selection not shown"
    fi

    rm -f test_gamestate.json
}

# Print summary
print_summary() {
    echo ""
    print_header "Test Summary"

    echo -e "Total Tests:  ${BLUE}${TESTS_TOTAL}${NC}"
    echo -e "Passed:       ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Failed:       ${RED}${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed!${NC}\n"
        return 0
    else
        echo -e "\n${RED}❌ Some tests failed${NC}\n"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════╗"
    echo "║  Fight-CLI Integration Test Suite   ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"

    # Change to script directory
    cd "$(dirname "$0")"

    # Clean up old test files
    cleanup

    # Build game
    build_game

    # Run tests
    test_new_character
    test_character_stats
    test_resource_harvest
    test_location_discovery
    test_quest_log
    test_save_load
    test_multiple_characters
    test_backward_compatibility
    test_single_combat
    test_autoplay_init

    # Print summary
    print_summary

    # Clean up
    cleanup

    exit $?
}

# Run main
main
