#!/bin/sh

npm install

if [ "$ENV" = "development" ]; then
    npm run dev
else
    npm run build && npm run start
fi
