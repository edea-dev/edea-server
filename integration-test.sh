#!/bin/bash

echo "running frontend tests"
npm install -D @playwright/test
npx playwright test
