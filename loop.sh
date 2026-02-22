#!/bin/bash
# loop.sh - Factory Droid execution engine
# Usage:
#   ./loop.sh           # Build mode, unlimited
#   ./loop.sh plan      # Plan mode (create IMPLEMENTATION_PLAN.md)
#   ./loop.sh 20        # Build mode, max 20 iterations
#   ./loop.sh plan 5    # Plan mode, max 5 iterations

set -e

MODE="${1:-build}"
MAX_ITERATIONS="${2:-999999}"

# Handle ./loop.sh 20 (number as first arg = build mode with limit)
if [[ "$MODE" =~ ^[0-9]+$ ]]; then
  MAX_ITERATIONS="$MODE"
  MODE="build"
fi

# Select prompt file and completion signal based on mode
if [ "$MODE" = "plan" ]; then
  PROMPT_FILE="PROMPT_plan.md"
  COMPLETE_SIGNAL="<promise>PLAN_COMPLETE</promise>"
else
  PROMPT_FILE="PROMPT_build.md"
  COMPLETE_SIGNAL="<promise>BUILD_COMPLETE</promise>"

  # Build mode requires an existing implementation plan
  if [ ! -f "IMPLEMENTATION_PLAN.md" ]; then
    echo "Error: IMPLEMENTATION_PLAN.md not found"
    echo "Run './loop.sh plan' first to create the implementation plan"
    exit 1
  fi
fi

# Verify prompt file exists
if [ ! -f "$PROMPT_FILE" ]; then
  echo "Error: $PROMPT_FILE not found"
  exit 1
fi

echo "Factory Droid Starting"
echo "   Mode: $MODE"
echo "   Max iterations: $MAX_ITERATIONS"
echo "   Prompt: $PROMPT_FILE"
echo "================================================"

for ((i=1; i<=MAX_ITERATIONS; i++)); do
  echo ""
  echo "==============================================="
  echo "  Iteration $i of $MAX_ITERATIONS"
  echo "==============================================="

  # Run Factory Droid in headless exec mode
  result=$(cat "$PROMPT_FILE" | droid exec --auto high)

  echo "$result"

  # Push after each iteration
  git push 2>/dev/null || true

  # Check for completion signal
  if [[ "$result" == *"$COMPLETE_SIGNAL"* ]]; then
    echo ""
    echo "All $MODE tasks complete at iteration $i"
    exit 0
  fi

  echo "Iteration $i complete. Continuing..."
done

echo ""
echo "Max iterations ($MAX_ITERATIONS) reached"
exit 1
