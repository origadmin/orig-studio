/**
 * Color Scale Generator for OrigStudio Dynamic Theme System.
 *
 * Generates 11-level color scales (50-950) for 6 color families
 * (primary, brand, success, warning, info, destructive) using
 * the HSL fixed-multiplier algorithm reverse-engineered from index.css.
 *
 * Usage:
 *   bun run scripts/generate-color-scales.ts              # Process all themes
 *   bun run scripts/generate-color-scales.ts --dry-run     # Preview without writing
 *   bun run scripts/generate-color-scales.ts --theme feishu-blue  # Process single theme
 */

import { readFileSync, writeFileSync, readdirSync, existsSync } from 'node:fs';
import { resolve, join } from 'node:path';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const LEVELS = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950] as const;
type Level = (typeof LEVELS)[number];

const SCALE_FAMILIES = ['primary', 'brand', 'success', 'warning', 'info', 'destructive'] as const;
type ScaleFamily = (typeof SCALE_FAMILIES)[number];

/** Light mode scale parameters (from SPEC 3.1.1) */
const LIGHT_SCALE: Record<Exclude<Level, 600>, { lightness: number; saturationRatio: number }> & {
  600: { saturationRatio: number };
} = {
  50:  { lightness: 97, saturationRatio: 0.30 },
  100: { lightness: 94, saturationRatio: 0.40 },
  200: { lightness: 88, saturationRatio: 0.55 },
  300: { lightness: 78, saturationRatio: 0.70 },
  400: { lightness: 65, saturationRatio: 0.85 },
  500: { lightness: 52, saturationRatio: 0.95 },
  600: { saturationRatio: 1.00 },
  700: { lightness: 34, saturationRatio: 0.90 },
  800: { lightness: 26, saturationRatio: 0.80 },
  900: { lightness: 18, saturationRatio: 0.70 },
  950: { lightness: 12, saturationRatio: 0.60 },
};

/** Dark mode scale parameters (from SPEC 3.1.1) */
const DARK_SCALE: Record<Exclude<Level, 600>, { lightness: number; saturationRatio: number }> & {
  600: { saturationRatio: number };
} = {
  50:  { lightness: 12, saturationRatio: 0.40 },
  100: { lightness: 16, saturationRatio: 0.50 },
  200: { lightness: 22, saturationRatio: 0.60 },
  300: { lightness: 30, saturationRatio: 0.70 },
  400: { lightness: 40, saturationRatio: 0.85 },
  500: { lightness: 50, saturationRatio: 0.95 },
  600: { saturationRatio: 1.00 },
  700: { lightness: 60, saturationRatio: 0.95 },
  800: { lightness: 70, saturationRatio: 0.85 },
  900: { lightness: 80, saturationRatio: 0.70 },
  950: { lightness: 88, saturationRatio: 0.55 },
};

/** Minimum saturation floor for low-saturation base colors (SPEC Appendix B.2) */
const MIN_SATURATION_FLOOR = 15;

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface HSL {
  h: number;
  s: number;
  l: number;
}

interface ScaleResult {
  [level: number]: HSL;
}

interface ThemeColors {
  light: Partial<Record<ScaleFamily, HSL>>;
  dark: Partial<Record<ScaleFamily, HSL>>;
}

interface ParsedBlock {
  selector: string;
  startIndex: number;
  endIndex: number;
  content: string;
  variables: Map<string, string>;
}

// ---------------------------------------------------------------------------
// Color Scale Generation Algorithm
// ---------------------------------------------------------------------------

/**
 * Generate an 11-level color scale from a base HSL color.
 *
 * Algorithm: HSL fixed-multiplier method
 * - Hue stays constant across all levels
 * - Saturation = base_S * saturationRatio[level]
 * - Lightness = fixed target value (except level 600 which uses base_L)
 * - For low-saturation base colors (S < 20%), apply saturation floor
 */
function generateScale(baseH: number, baseS: number, baseL: number, mode: 'light' | 'dark'): ScaleResult {
  const scaleConfig = mode === 'light' ? LIGHT_SCALE : DARK_SCALE;
  const result: ScaleResult = {};

  // Apply saturation floor for low-saturation base colors
  const effectiveBaseS = baseS < 20 ? Math.max(baseS, MIN_SATURATION_FLOOR) : baseS;

  for (const level of LEVELS) {
    const config = scaleConfig[level as keyof typeof scaleConfig];
    const h = Math.round(baseH * 10) / 10;
    const s = Math.min(100, Math.max(0, Math.round(effectiveBaseS * config.saturationRatio * 10) / 10));
    const l = level === 600
      ? Math.round(baseL * 10) / 10
      : (config as { lightness: number; saturationRatio: number }).lightness;

    result[level] = { h, s, l };
  }

  return result;
}

// ---------------------------------------------------------------------------
// HSL <-> HEX Conversion
// ---------------------------------------------------------------------------

/**
 * Convert HSL values to HEX string.
 * Used for registry.json preview values.
 */
function hslToHex(h: number, s: number, l: number): string {
  const sNorm = s / 100;
  const lNorm = l / 100;
  const c = (1 - Math.abs(2 * lNorm - 1)) * sNorm;
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
  const m = lNorm - c / 2;

  let r = 0, g = 0, b = 0;
  if (h >= 0 && h < 60)       { r = c; g = x; b = 0; }
  else if (h >= 60 && h < 120)  { r = x; g = c; b = 0; }
  else if (h >= 120 && h < 180) { r = 0; g = c; b = x; }
  else if (h >= 180 && h < 240) { r = 0; g = x; b = c; }
  else if (h >= 240 && h < 300) { r = x; g = 0; b = c; }
  else if (h >= 300 && h < 360) { r = c; g = 0; b = x; }

  const toHex = (v: number) => {
    const hex = Math.round((v + m) * 255).toString(16);
    return hex.length === 1 ? '0' + hex : hex;
  };

  return `#${toHex(r)}${toHex(g)}${toHex(b)}`.toUpperCase();
}

// ---------------------------------------------------------------------------
// CSS Parsing
// ---------------------------------------------------------------------------

/**
 * Parse an HSL CSS variable value string like "215.7 95.3% 42.9%" into HSL object.
 */
function parseHSLValue(value: string): HSL | null {
  const match = value.trim().match(/^([\d.]+)\s+([\d.]+)%\s+([\d.]+)%$/);
  if (!match) return null;
  return {
    h: parseFloat(match[1]),
    s: parseFloat(match[2]),
    l: parseFloat(match[3]),
  };
}

/**
 * Parse a theme CSS file and extract all CSS variable blocks.
 * Returns blocks for [data-theme="xxx"] and [data-theme="xxx"].dark selectors.
 */
function parseThemeCSS(cssText: string): ParsedBlock[] {
  const blocks: ParsedBlock[] = [];
  // Match selectors like [data-theme="xxx"] or [data-theme="xxx"].dark
  const blockRegex = /(\[data-theme="[^"]+"\](?:\.dark)?)\s*\{/g;
  let match: RegExpExecArray | null;

  while ((match = blockRegex.exec(cssText)) !== null) {
    const selector = match[1];
    const startIndex = match.index + match[0].length;
    let depth = 1;
    let endIndex = startIndex;

    while (depth > 0 && endIndex < cssText.length) {
      if (cssText[endIndex] === '{') depth++;
      else if (cssText[endIndex] === '}') depth--;
      if (depth > 0) endIndex++;
    }

    const content = cssText.substring(startIndex, endIndex);
    const variables = new Map<string, string>();

    // Parse CSS variable declarations
    const varRegex = /--([\w-]+)\s*:\s*([^;]+);/g;
    let varMatch: RegExpExecArray | null;
    while ((varMatch = varRegex.exec(content)) !== null) {
      variables.set(varMatch[1], varMatch[2].trim());
    }

    blocks.push({
      selector,
      startIndex: match.index,
      endIndex: endIndex + 1,
      content,
      variables,
    });
  }

  return blocks;
}

/**
 * Extract base color HSL values from a parsed CSS block.
 */
function extractBaseColors(block: ParsedBlock): Partial<Record<ScaleFamily, HSL>> {
  const colors: Partial<Record<ScaleFamily, HSL>> = {};

  for (const family of SCALE_FAMILIES) {
    const value = block.variables.get(family);
    if (value) {
      const hsl = parseHSLValue(value);
      if (hsl) {
        colors[family] = hsl;
      }
    }
  }

  return colors;
}

// ---------------------------------------------------------------------------
// CSS Variable Formatting
// ---------------------------------------------------------------------------

/**
 * Format an HSL object as a CSS variable value string.
 * Example: "215.7 95.3% 42.9%"
 */
function formatHSLValue(hsl: HSL): string {
  const h = Number.isInteger(hsl.h) ? hsl.h.toString() : hsl.h.toString();
  const s = Number.isInteger(hsl.s) ? `${hsl.s}%` : `${hsl.s}%`;
  const l = Number.isInteger(hsl.l) ? `${hsl.l}%` : `${hsl.l}%`;
  return `${h} ${s} ${l}`;
}

/**
 * Generate CSS variable lines for a color family scale.
 * Returns lines like: "  --primary-50: 220 27% 97%;"
 */
function generateScaleCSSLines(family: ScaleFamily, scale: ScaleResult): string[] {
  return LEVELS.map(level => {
    const hsl = scale[level];
    return `  --${family}-${level}: ${formatHSLValue(hsl)};`;
  });
}

// ---------------------------------------------------------------------------
// CSS Injection
// ---------------------------------------------------------------------------

/**
 * Inject color scale variables into a theme CSS block content.
 *
 * Strategy: Insert scale variables after the base variable + foreground variable group.
 * If scale variables already exist, skip (idempotency).
 */
function injectScaleVarsIntoBlock(
  blockContent: string,
  colors: Partial<Record<ScaleFamily, HSL>>,
  mode: 'light' | 'dark',
): string {
  let result = blockContent;

  // Process each family in reverse order to maintain correct string indices
  for (const family of [...SCALE_FAMILIES].reverse()) {
    const baseHSL = colors[family];
    if (!baseHSL) continue;

    // Check if scale variables already exist for this family
    const scaleVarPattern = new RegExp(`--${family}-50\\s*:`);
    if (scaleVarPattern.test(result)) continue;

    const scale = generateScale(baseHSL.h, baseHSL.s, baseHSL.l, mode);
    const scaleLines = generateScaleCSSLines(family, scale);

    // Find the insertion point: after the last variable in the family's group
    // The group is: --{family}, --{family}-foreground, --{family}-muted, etc.
    // We insert after the last variable that starts with --{family}-
    const familyVarRegex = new RegExp(`--${family}(?:-[\\w-]+)?\\s*:[^;]+;`, 'g');
    let lastFamilyVarMatch: RegExpExecArray | null = null;
    let familyMatch: RegExpExecArray | null;

    while ((familyMatch = familyVarRegex.exec(result)) !== null) {
      lastFamilyVarMatch = familyMatch;
    }

    if (lastFamilyVarMatch) {
      const insertIndex = lastFamilyVarMatch.index + lastFamilyVarMatch[0].length;
      const scaleBlock = '\n' + scaleLines.join('\n');
      result = result.substring(0, insertIndex) + scaleBlock + result.substring(insertIndex);
    }
  }

  return result;
}

/**
 * Inject color scale variables into a complete theme CSS file.
 */
function injectScaleVars(cssText: string, themeId: string): string {
  const blocks = parseThemeCSS(cssText);
  if (blocks.length === 0) {
    console.warn(`  Warning: No [data-theme] blocks found in theme "${themeId}"`);
    return cssText;
  }

  // Process blocks in reverse order to maintain string indices
  let result = cssText;
  const reversedBlocks = [...blocks].reverse();

  for (const block of reversedBlocks) {
    const isDark = block.selector.endsWith('.dark');
    const mode = isDark ? 'dark' : 'light';
    const colors = extractBaseColors(block);

    if (Object.keys(colors).length === 0) {
      console.warn(`  Warning: No base colors found in block "${block.selector}"`);
      continue;
    }

    const newContent = injectScaleVarsIntoBlock(block.content, colors, mode);
    result = result.substring(0, block.startIndex + block.selector.length + 2) +
             newContent +
             result.substring(block.endIndex - 1);
  }

  return result;
}

// ---------------------------------------------------------------------------
// Registry Update
// ---------------------------------------------------------------------------

interface RegistryTheme {
  id: string;
  preview: Record<string, string>;
  [key: string]: unknown;
}

interface Registry {
  version: number;
  defaultTheme: string;
  themes: RegistryTheme[];
}

/**
 * Update registry.json with color scale HEX preview values.
 */
function updateRegistry(registry: Registry, themeId: string, lightColors: Partial<Record<ScaleFamily, HSL>>): Registry {
  const theme = registry.themes.find(t => t.id === themeId);
  if (!theme) {
    console.warn(`  Warning: Theme "${themeId}" not found in registry.json`);
    return registry;
  }

  for (const family of SCALE_FAMILIES) {
    const baseHSL = lightColors[family];
    if (!baseHSL) continue;

    theme.preview[family] = hslToHex(baseHSL.h, baseHSL.s, baseHSL.l);

    const scale = generateScale(baseHSL.h, baseHSL.s, baseHSL.l, 'light');
    for (const level of LEVELS) {
      const hsl = scale[level];
      const hex = hslToHex(hsl.h, hsl.s, hsl.l);
      theme.preview[`${family}-${level}`] = hex;
    }
  }

  return registry;
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

/**
 * Validate generated scales against known reference values from index.css.
 * Returns the maximum deviation percentage.
 */
function validateAgainstReference(
  family: ScaleFamily,
  generatedScale: ScaleResult,
  referenceValues: Record<number, HSL>,
): { maxDeviation: number; details: string[] } {
  const details: string[] = [];
  let maxDeviation = 0;

  for (const level of LEVELS) {
    const gen = generatedScale[level];
    const ref = referenceValues[level];
    if (!ref) continue;

    const hDiff = Math.abs(gen.h - ref.h);
    const sDiff = Math.abs(gen.s - ref.s);
    const lDiff = Math.abs(gen.l - ref.l);
    const maxDiff = Math.max(hDiff, sDiff, lDiff);
    maxDeviation = Math.max(maxDeviation, maxDiff);

    if (maxDiff > 0.5) {
      details.push(`  ${family}-${level}: generated HSL(${gen.h}, ${gen.s}%, ${gen.l}%) vs reference HSL(${ref.h}, ${ref.s}%, ${ref.l}%) [diff: ${maxDiff.toFixed(1)}]`);
    }
  }

  return { maxDeviation, details };
}

// ---------------------------------------------------------------------------
// CSS Block Parsing for src/index.css (uses :root/.dark selectors, not [data-theme])
// ---------------------------------------------------------------------------

interface CSSBlock {
  selector: string;
  content: string;
}

/**
 * Parse CSS text into blocks by selector.
 * Handles :root, .dark, [data-theme="xxx"], etc.
 * Works with nested blocks (e.g., :root inside @layer base).
 * Uses character-by-character scanning with brace depth tracking.
 */
function parseCSSBlocks(cssText: string): CSSBlock[] {
  const blocks: CSSBlock[] = [];
  let i = 0;

  while (i < cssText.length) {
    // Skip whitespace and comments
    if (cssText[i] === '/' && cssText[i + 1] === '*') {
      // Skip block comment
      const end = cssText.indexOf('*/', i + 2);
      i = end === -1 ? cssText.length : end + 2;
      continue;
    }
    if (cssText[i] === '/' && cssText[i + 1] === '/') {
      // Skip line comment
      const end = cssText.indexOf('\n', i + 2);
      i = end === -1 ? cssText.length : end + 1;
      continue;
    }

    // Find the next opening brace
    const braceIndex = cssText.indexOf('{', i);
    if (braceIndex === -1) break;

    const selector = cssText.substring(i, braceIndex).trim();

    // Find the matching closing brace
    let depth = 1;
    let contentStart = braceIndex + 1;
    let contentEnd = contentStart;
    while (depth > 0 && contentEnd < cssText.length) {
      if (cssText[contentEnd] === '{') depth++;
      else if (cssText[contentEnd] === '}') depth--;
      if (depth > 0) contentEnd++;
    }

    const content = cssText.substring(contentStart, contentEnd);

    // For @-rules like @layer, recursively parse the inner content
    if (selector.startsWith('@')) {
      const innerBlocks = parseCSSBlocks(content);
      blocks.push(...innerBlocks);
    } else if (selector) {
      blocks.push({ selector, content });
    }

    i = contentEnd + 1;
  }

  return blocks;
}

/**
 * Parse CSS variable declarations from a block content string.
 */
function parseVariables(content: string): Map<string, string> {
  const variables = new Map<string, string>();
  const varRegex = /--([\w-]+)\s*:\s*([^;]+);/g;
  let varMatch: RegExpExecArray | null;
  while ((varMatch = varRegex.exec(content)) !== null) {
    variables.set(varMatch[1], varMatch[2].trim());
  }
  return variables;
}

/**
 * Validate generated scales against reference values from src/index.css.
 */
function validateDefaultTheme(referenceVars: Map<string, string>, mode: 'light' | 'dark'): void {
  let overallMaxDeviation = 0;
  const allDetails: string[] = [];

  for (const family of SCALE_FAMILIES) {
    // Get base color from reference
    const baseValue = referenceVars.get(family);
    if (!baseValue) continue;

    const baseHSL = parseHSLValue(baseValue);
    if (!baseHSL) continue;

    // Extract reference scale values
    const referenceValues: Record<number, HSL> = {};
    for (const level of LEVELS) {
      const varName = `${family}-${level}`;
      const value = referenceVars.get(varName);
      if (value) {
        const hsl = parseHSLValue(value);
        if (hsl) referenceValues[level] = hsl;
      }
    }

    if (Object.keys(referenceValues).length === 0) continue;

    // Generate scale and validate
    const generatedScale = generateScale(baseHSL.h, baseHSL.s, baseHSL.l, mode);
    const { maxDeviation, details } = validateAgainstReference(family, generatedScale, referenceValues);
    overallMaxDeviation = Math.max(overallMaxDeviation, maxDeviation);
    allDetails.push(...details);

    console.log(`    ${family}: max deviation = ${maxDeviation.toFixed(2)}%`);
  }

  if (allDetails.length > 0) {
    console.log(`\n    Deviations > 0.5%:`);
    allDetails.forEach(d => console.log(d));
  }

  if (overallMaxDeviation < 0.5) {
    console.log(`\n    PASS: All deviations < 0.5% (max: ${overallMaxDeviation.toFixed(2)}%)`);
  } else {
    console.log(`\n    FAIL: Max deviation ${overallMaxDeviation.toFixed(2)}% exceeds 0.5% threshold`);
  }
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main() {
  const args = process.argv.slice(2);
  const dryRun = args.includes('--dry-run');
  const themeIndex = args.indexOf('--theme');
  const specificTheme = themeIndex !== -1 && args[themeIndex + 1] ? args[themeIndex + 1] : null;

  const webRoot = resolve(import.meta.dir, '..');
  const themesDir = join(webRoot, 'public', 'themes');
  const registryPath = join(themesDir, 'registry.json');

  console.log('=== Color Scale Generator ===');
  console.log(`Web root: ${webRoot}`);
  console.log(`Themes dir: ${themesDir}`);
  if (dryRun) console.log('Mode: DRY RUN (no files will be written)');
  if (specificTheme) console.log(`Target theme: ${specificTheme}`);
  console.log('');

  // Read registry
  const registryText = readFileSync(registryPath, 'utf-8');
  const registry: Registry = JSON.parse(registryText);

  // Discover theme directories
  const themeDirs = readdirSync(themesDir, { withFileTypes: true })
    .filter(d => d.isDirectory() && existsSync(join(themesDir, d.name, 'index.css')))
    .map(d => d.name)
    .filter(name => !specificTheme || name === specificTheme);

  if (specificTheme && !themeDirs.includes(specificTheme)) {
    console.error(`Error: Theme "${specificTheme}" not found in ${themesDir}`);
    process.exit(1);
  }

  console.log(`Found ${themeDirs.length} theme(s) to process: ${themeDirs.join(', ')}\n`);

  let totalVarsAdded = 0;
  let totalThemesUpdated = 0;

  for (const themeId of themeDirs) {
    console.log(`Processing theme: ${themeId}`);
    const cssPath = join(themesDir, themeId, 'index.css');
    const cssText = readFileSync(cssPath, 'utf-8');

    // Parse and extract base colors
    const blocks = parseThemeCSS(cssText);
    const lightBlock = blocks.find(b => !b.selector.endsWith('.dark'));
    const darkBlock = blocks.find(b => b.selector.endsWith('.dark'));

    const lightColors = lightBlock ? extractBaseColors(lightBlock) : {};
    const darkColors = darkBlock ? extractBaseColors(darkBlock) : {};

    console.log(`  Light mode base colors: ${Object.keys(lightColors).join(', ')}`);
    console.log(`  Dark mode base colors: ${Object.keys(darkColors).join(', ')}`);

    // Check if scale variables already exist
    const hasExistingScales = lightBlock
      ? SCALE_FAMILIES.some(f => lightBlock.variables.has(`${f}-50`))
      : false;

    if (hasExistingScales) {
      console.log(`  Scale variables already exist, will be preserved (idempotent)`);
    }

    // Generate and inject CSS
    const updatedCSS = injectScaleVars(cssText, themeId);

    // Count new variables added
    const updatedBlocks = parseThemeCSS(updatedCSS);
    const updatedLightBlock = updatedBlocks.find(b => !b.selector.endsWith('.dark'));
    const newVarCount = updatedLightBlock
      ? LEVELS.length * Object.keys(lightColors).length * (darkBlock ? 2 : 1)
      : 0;

    console.log(`  Generated ${newVarCount} scale variables`);
    totalVarsAdded += newVarCount;

    if (!dryRun) {
      writeFileSync(cssPath, updatedCSS, 'utf-8');
      console.log(`  Written: ${cssPath}`);
    } else {
      console.log(`  [DRY RUN] Would write: ${cssPath}`);
    }

    // Update registry
    updateRegistry(registry, themeId, lightColors);
    totalThemesUpdated++;

    console.log('');
  }

  // Write updated registry
  if (!dryRun) {
    writeFileSync(registryPath, JSON.stringify(registry, null, 2) + '\n', 'utf-8');
    console.log(`Updated registry: ${registryPath}`);
  } else {
    console.log(`[DRY RUN] Would update registry: ${registryPath}`);
  }

  console.log(`\n=== Summary ===`);
  console.log(`Themes processed: ${totalThemesUpdated}`);
  console.log(`Scale variables added: ${totalVarsAdded}`);
  console.log(`Registry preview entries: ${totalThemesUpdated * SCALE_FAMILIES.length * LEVELS.length}`);

  // Validation: compare default theme generated values with src/index.css reference
  // src/index.css has the known-good color scales in :root and .dark selectors
  if (themeDirs.includes('default') || !specificTheme) {
    console.log(`\n=== Validation: Default Theme vs src/index.css ===`);
    const srcIndexCssPath = join(webRoot, 'src', 'index.css');
    if (!existsSync(srcIndexCssPath)) {
      console.log('  SKIP: src/index.css not found, cannot validate');
    } else {
      const srcCss = readFileSync(srcIndexCssPath, 'utf-8');

      // Parse :root and .dark blocks from src/index.css
      const srcBlocks = parseCSSBlocks(srcCss);
      const rootBlock = srcBlocks.find(b => b.selector === ':root');
      const darkBlockSrc = srcBlocks.find(b => b.selector === '.dark');

      // Validate light mode
      if (rootBlock) {
        const rootVars = parseVariables(rootBlock.content);
        console.log('\n  Light mode validation:');
        validateDefaultTheme(rootVars, 'light');
      }

      // Validate dark mode
      if (darkBlockSrc) {
        const darkVars = parseVariables(darkBlockSrc.content);
        console.log('\n  Dark mode validation:');
        validateDefaultTheme(darkVars, 'dark');
      }
    }
  }
}

main();
