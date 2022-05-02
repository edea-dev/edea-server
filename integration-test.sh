#!/bin/bash

echo "running the test now"
npm install -D @playwright/test
npx playwright test
