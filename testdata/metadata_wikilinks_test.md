---
publish: true
title: "Metadata Wikilinks Test"
author: "Test Author"
related_notes:
  - "[[Public Test Note]]"
  - "[[Private Test Note]]"
description: "This note tests wikilinks in metadata like [[Public Test Note|this link]]"
tags:
  - "test"
  - "[[Public Test Note]]"
nested_object:
  reference: "[[Public Test Note]]"
  description: "A nested reference"
---

# Metadata Wikilinks Test

This note is used to test wikilink parsing in metadata fields.

The metadata above contains wikilinks in:

- String fields (description)
- List fields (related_notes, tags)
- Nested object fields (nested_object.reference)

These should all be parsed and rendered as clickable links in the metadata section.
