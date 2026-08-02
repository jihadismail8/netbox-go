#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const DOCS_DIR = path.join(ROOT, 'docs');

function markdownFiles(directory) {
  if (!fs.existsSync(directory)) return [];

  return fs
    .readdirSync(directory, { withFileTypes: true })
    .flatMap((entry) => {
      const candidate = path.join(directory, entry.name);
      if (entry.isDirectory()) return markdownFiles(candidate);
      return entry.isFile() && entry.name.endsWith('.md') ? [candidate] : [];
    });
}

function linkTargets(markdown) {
  const targets = [];
  const inline = /!?\[[^\]]*\]\((<[^>]+>|[^)\s]+)(?:\s+["'][^"']*["'])?\)/g;
  const reference = /^\s*\[[^\]]+\]:\s*(<[^>]+>|\S+)/gm;

  for (const expression of [inline, reference]) {
    for (const match of markdown.matchAll(expression)) targets.push(match[1]);
  }
  return targets;
}

function localPath(sourceFile, rawTarget) {
  const target = rawTarget.startsWith('<') && rawTarget.endsWith('>')
    ? rawTarget.slice(1, -1)
    : rawTarget;
  if (
    target === '' ||
    target.startsWith('#') ||
    target.startsWith('//') ||
    /^[a-z][a-z\d+.-]*:/i.test(target)
  ) {
    return null;
  }

  const withoutFragment = target.split('#', 1)[0].split('?', 1)[0];
  if (withoutFragment === '') return null;

  let decoded;
  try {
    decoded = decodeURIComponent(withoutFragment);
  } catch {
    decoded = withoutFragment;
  }

  return path.resolve(path.dirname(sourceFile), decoded);
}

const rootMarkdown = fs
  .readdirSync(ROOT, { withFileTypes: true })
  .filter((entry) => entry.isFile() && entry.name.endsWith('.md'))
  .map((entry) => path.join(ROOT, entry.name));
const files = [...rootMarkdown, ...markdownFiles(DOCS_DIR)].sort();
const failures = [];

for (const file of files) {
  const content = fs.readFileSync(file, 'utf8');
  for (const target of linkTargets(content)) {
    const candidate = localPath(file, target);
    if (candidate !== null && !fs.existsSync(candidate)) {
      failures.push(`${path.relative(ROOT, file)} -> ${target}`);
    }
  }
}

if (failures.length > 0) {
  console.error('Broken local Markdown links:');
  for (const failure of failures) console.error(`  ${failure}`);
  process.exitCode = 1;
} else {
  console.log(`Checked ${files.length} Markdown files: local links are valid`);
}
