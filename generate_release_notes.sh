#!/bin/bash
git pull --tags
export PREVIOUS_VERSION=$(git tag --sort=-committerdate | head -1)
export CHANGES=$(git log --pretty="- %ai | %an | %s" $PREVIOUS_VERSION..@)
printf "Changes since $PREVIOUS_VERSION\n$CHANGES" 

