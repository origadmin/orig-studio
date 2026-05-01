/**
 * Slug generation utility for frontend preview.
 *
 * This provides a client-side preview of what the slug will look like.
 * The actual slug is generated on the backend using the definitive algorithm
 * (pure ASCII -> slugify, non-ASCII -> Base58 encode).
 *
 * The frontend only does a simple slugify for preview purposes.
 */

/**
 * Checks if a string contains only ASCII characters.
 */
function isASCII(s: string): boolean {
  for (let i = 0; i < s.length; i++) {
    if (s.charCodeAt(i) > 127) {
      return false;
    }
  }
  return true;
}

/**
 * Generates a URL-friendly slug from a name string.
 * For pure ASCII names: lowercase, replace non-alphanumeric with hyphens.
 * For names with non-ASCII characters: returns a placeholder indicating
 * the slug will be auto-generated on the backend (Base58 encoded).
 *
 * @param name - The tag name to generate a slug from
 * @returns The generated slug string
 */
export function generateSlug(name: string): string {
  if (!name || !name.trim()) {
    return '';
  }

  const trimmed = name.trim();

  if (!isASCII(trimmed)) {
    // Non-ASCII names will be Base58-encoded on the backend.
    // Show a placeholder to indicate auto-generation.
    return '(auto-generated)';
  }

  // Simple slugify for ASCII names
  let slug = trimmed.toLowerCase();
  // Replace non-alphanumeric characters with hyphens
  slug = slug.replace(/[^a-z0-9]+/g, '-');
  // Trim leading/trailing hyphens
  slug = slug.replace(/^-+|-+$/g, '');
  // Collapse consecutive hyphens
  slug = slug.replace(/-{2,}/g, '-');

  return slug;
}
