#!/bin/bash

echo "running the test now"
echo $TEST_SERVER
ping -c4 edea-server
npm install -D @playwright/test
npx playwright test
