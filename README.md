# jenkins_tui (jui)

A small terminal UI client for interacting with Jenkins. Provides a lightweight TUI to view jobs, builds, and basic Jenkins information from the terminal.

It is not done yet, so do not bother trying to run it.

## TO DO:
- Manage projects that have multiple branches to build
- Send command to start new build
- Auto refresh data where needed

## Features
- Browse Jenkins jobs and builds
- View build status and logs

## Requirements
- Go 1.20+ (or compatible)
- Network access to a Jenkins server

## Project layout
- main.go — application entrypoint
- config/ — configuration files
- view/ — TUI view templates/components
- data/ — runtime data and caches
