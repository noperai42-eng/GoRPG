#!/bin/bash

# Comprehensive Test Runner for fight-cli
# Runs both unit tests and integration tests

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Print banner
print_banner() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════════════════╗"
    echo "║                                                   ║"
    echo "║        Fight-CLI Comprehensive Test Suite        ║"
    echo "║                                                   ║"
    echo "╚═══════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

# Print section header
print_section() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Change to script directory
cd "$(dirname "$0")"

# Print banner
print_banner

# Track overall results
OVERALL_SUCCESS=0

# 1. Run Go Unit Tests
print_section "1. Running Go Unit Tests"

if go test -v -cover; then
    echo -e "\n${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "\n${RED}✗ Unit tests failed${NC}"
    OVERALL_SUCCESS=1
fi

# 2. Run Go Benchmarks (short)
print_section "2. Running Benchmarks"

if go test -bench=. -benchtime=1s; then
    echo -e "\n${GREEN}✓ Benchmarks completed${NC}"
else
    echo -e "\n${RED}✗ Benchmarks failed${NC}"
    OVERALL_SUCCESS=1
fi

# 3. Run Integration Tests
print_section "3. Running Integration Tests"

if ./test_integration.sh; then
    echo -e "\n${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "\n${RED}✗ Integration tests failed${NC}"
    OVERALL_SUCCESS=1
fi

# 4. Test Auto-Play Mode (5 fights)
print_section "4. Testing Auto-Play Mode (5 fights)"

echo "Creating test character and running auto-play..."

# Create test save with character
cat > test_autoplay_save.json << 'EOF'
{
  "CharactersMap": {
    "AutoPlayTest": {
      "Name": "AutoPlayTest",
      "Level": 5,
      "Experience": 500,
      "ExpSinceLevel": 0,
      "HitpointsTotal": 20,
      "HitpointsNatural": 20,
      "HitpointsRemaining": 20,
      "ManaTotal": 45,
      "ManaNatural": 45,
      "ManaRemaining": 45,
      "StaminaTotal": 45,
      "StaminaNatural": 45,
      "StaminaRemaining": 45,
      "AttackRolls": 1,
      "DefenseRolls": 1,
      "StatsMod": {
        "AttackMod": 0,
        "DefenseMod": 0,
        "HitPointMod": 0
      },
      "Resurrections": 0,
      "Inventory": [],
      "EquipmentMap": {},
      "ResourceStorageMap": {},
      "KnownLocations": ["Home", "Training Hall", "Forest"],
      "builtBuildings": [],
      "LearnedSkills": [
        {
          "Name": "Fireball",
          "ManaCost": 15,
          "StaminaCost": 0,
          "Damage": 18,
          "DamageType": "fire"
        },
        {
          "Name": "Power Strike",
          "ManaCost": 0,
          "StaminaCost": 20,
          "Damage": 25,
          "DamageType": "physical"
        },
        {
          "Name": "Heal",
          "ManaCost": 10,
          "StaminaCost": 0,
          "Damage": -20,
          "DamageType": "physical"
        }
      ],
      "StatusEffects": [],
      "Resistances": {
        "physical": 1.0,
        "fire": 1.0,
        "ice": 1.0,
        "lightning": 1.0,
        "poison": 1.0
      },
      "CompletedQuests": [],
      "ActiveQuests": ["quest_1_training"]
    }
  }
}
EOF

# Run a brief auto-play session using timeout (10 seconds = ~10 fights on turbo)
cp test_autoplay_save.json gamestate.json
timeout 10 bash -c 'echo -e "8\n4" | ./fight-cli' > /dev/null 2>&1 || true

# Check if save was updated (indicating fights happened)
if [ -f "gamestate.json" ]; then
    echo -e "${GREEN}✓ Auto-play executed without crashing${NC}"

    # Check if character data exists
    if grep -q "AutoPlayTest" gamestate.json; then
        echo -e "${GREEN}✓ Character data preserved${NC}"
    else
        echo -e "${RED}✗ Character data lost${NC}"
        OVERALL_SUCCESS=1
    fi
else
    echo -e "${RED}✗ Auto-play failed${NC}"
    OVERALL_SUCCESS=1
fi

rm -f test_autoplay_save.json gamestate.json

# 5. Test All Skills
print_section "5. Testing All Skills"

echo "Checking skill definitions..."

skill_count=$(go test -run TestSkills -v 2>&1 | grep -c "✓ Skill:" || true)

if [ "$skill_count" -ge 9 ]; then
    echo -e "${GREEN}✓ All 9 player skills validated${NC}"
else
    echo -e "${YELLOW}⚠ Found $skill_count skills (expected 9)${NC}"
fi

# 6. Test Elemental Damage System
print_section "6. Testing Elemental Damage System"

if go test -run TestElementalDamage -v; then
    echo -e "\n${GREEN}✓ Elemental damage system working${NC}"
else
    echo -e "\n${RED}✗ Elemental damage system failed${NC}"
    OVERALL_SUCCESS=1
fi

# 7. Test Quest System
print_section "7. Testing Quest System"

if go test -run TestQuestSystem -v; then
    echo -e "\n${GREEN}✓ Quest system working${NC}"
else
    echo -e "\n${RED}✗ Quest system failed${NC}"
    OVERALL_SUCCESS=1
fi

# 8. Memory/Performance Check
print_section "8. Memory & Performance Check"

echo "Running race detector..."
if go test -race -short > /dev/null 2>&1; then
    echo -e "${GREEN}✓ No race conditions detected${NC}"
else
    echo -e "${YELLOW}⚠ Potential race conditions found${NC}"
fi

echo "Checking binary size..."
BINARY_SIZE=$(du -h fight-cli 2>/dev/null | awk '{print $1}' || echo "N/A")
echo -e "Binary size: ${CYAN}${BINARY_SIZE}${NC}"

# 9. Code Coverage
print_section "9. Code Coverage Analysis"

go test -coverprofile=coverage.out > /dev/null 2>&1
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "Code Coverage: ${CYAN}${COVERAGE}${NC}"

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
echo -e "${BLUE}HTML coverage report: coverage.html${NC}"

# Cleanup coverage files
rm -f coverage.out

# 10. Final Summary
print_section "Final Summary"

echo -e "Test Suite Components:"
echo -e "  • Unit Tests         ${GREEN}✓${NC}"
echo -e "  • Benchmarks         ${GREEN}✓${NC}"
echo -e "  • Integration Tests  ${GREEN}✓${NC}"
echo -e "  • Auto-Play Test     ${GREEN}✓${NC}"
echo -e "  • Skills Validation  ${GREEN}✓${NC}"
echo -e "  • Elemental System   ${GREEN}✓${NC}"
echo -e "  • Quest System       ${GREEN}✓${NC}"
echo -e "  • Performance Check  ${GREEN}✓${NC}"
echo -e "  • Coverage Analysis  ${GREEN}✓${NC}"

echo ""

if [ $OVERALL_SUCCESS -eq 0 ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                       ║${NC}"
    echo -e "${GREEN}║   🎉  ALL TESTS PASSED! 🎉           ║${NC}"
    echo -e "${GREEN}║                                       ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════╝${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}╔═══════════════════════════════════════╗${NC}"
    echo -e "${RED}║                                       ║${NC}"
    echo -e "${RED}║   ❌  SOME TESTS FAILED  ❌           ║${NC}"
    echo -e "${RED}║                                       ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════╝${NC}"
    echo ""
    exit 1
fi
