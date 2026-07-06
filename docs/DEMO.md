# Demo Recording Guide

This document outlines the recommended flow for recording a short CatScope preview demo.

The demo should use the final logo and real application UI only. Do not include `Logo.candidate.png`, mock screenshots, synthetic crash claims, or external AI API workflows.

## Suggested Flow

1. Start CatScope.
2. Select an Android device.
3. Click **Start Logcat**.
4. Select a package to enable package-aware filtering.
5. Click **Analyze Current Logs**.
6. Copy the generated AI Context.
7. Open an offline log file.
8. Save a session.
9. Open the saved session.

## Recording Notes

- Keep the recording short and focused, ideally 60-90 seconds.
- Use sanitized logs and package names.
- Show that AI Context is generated locally as Markdown.
- Avoid showing tokens, private device identifiers, internal package names, or customer data.
- Export the final demo as `docs/assets/catscope-demo.gif` only after a real recording is available.

No demo GIF is required for the current documentation pass.
