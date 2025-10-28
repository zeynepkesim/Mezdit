/**
 * Input validation and sanitization utilities for XSS prevention
 * 
 * Note: Vue.js automatically escapes HTML in templates using {{ }}
 * These utilities provide additional validation and sanitization
 */

/**
 * Sanitize text input by escaping HTML special characters
 * This is a JavaScript-side validation for extra security
 * @param {string} input - The input to sanitize
 * @returns {string} Sanitized input
 */
export function sanitizeInput(input) {
  if (typeof input !== 'string') {
    return '';
  }

  // Create a temporary div element to use browser's built-in HTML escaping
  const div = document.createElement('div');
  div.textContent = input;
  return div.innerHTML;
}

/**
 * Validate text length
 * @param {string} text - Text to validate
 * @param {number} maxLength - Maximum allowed length
 * @returns {boolean} True if valid, false otherwise
 */
export function validateLength(text, maxLength) {
  return text && text.length <= maxLength;
}

/**
 * Validate username format
 * Only allows alphanumeric characters, underscores, and hyphens
 * @param {string} username - Username to validate
 * @returns {boolean} True if valid, false otherwise
 */
export function validateUsername(username) {
  const usernameRegex = /^[a-zA-Z0-9_-]+$/;
  return usernameRegex.test(username);
}

/**
 * Check if input contains potentially dangerous patterns
 * @param {string} input - Input to check
 * @returns {boolean} True if safe, false if dangerous patterns detected
 */
export function isSafeInput(input) {
  if (typeof input !== 'string') {
    return false;
  }

  // Check for common XSS patterns
  const dangerousPatterns = [
    /<script/i,
    /javascript:/i,
    /on\w+\s*=/i, // Event handlers like onclick, onerror
    /<iframe/i,
    /<embed/i,
    /<object/i,
  ];

  return !dangerousPatterns.some(pattern => pattern.test(input));
}

/**
 * Validate and sanitize post content
 * @param {string} content - Post content
 * @param {number} maxLength - Maximum allowed length
 * @returns {object} {valid: boolean, message: string}
 */
export function validatePostContent(content, maxLength = 5000) {
  if (!content || content.trim().length === 0) {
    return { valid: false, message: 'Content cannot be empty' };
  }

  if (content.length > maxLength) {
    return { valid: false, message: `Content is too long (max ${maxLength} characters)` };
  }

  if (!isSafeInput(content)) {
    return { valid: false, message: 'Content contains invalid characters or patterns' };
  }

  return { valid: true, message: '' };
}

/**
 * Validate post header/title
 * @param {string} header - Post header
 * @param {number} maxLength - Maximum allowed length
 * @returns {object} {valid: boolean, message: string}
 */
export function validatePostHeader(header, maxLength = 500) {
  if (!header || header.trim().length === 0) {
    return { valid: false, message: 'Title cannot be empty' };
  }

  if (header.length > maxLength) {
    return { valid: false, message: `Title is too long (max ${maxLength} characters)` };
  }

  if (!isSafeInput(header)) {
    return { valid: false, message: 'Title contains invalid characters or patterns' };
  }

  return { valid: true, message: '' };
}

/**
 * Validate name field
 * @param {string} name - Name to validate
 * @param {number} maxLength - Maximum allowed length
 * @returns {object} {valid: boolean, message: string}
 */
export function validateName(name, maxLength = 255) {
  if (!name || name.trim().length === 0) {
    return { valid: false, message: 'Name cannot be empty' };
  }

  if (name.length > maxLength) {
    return { valid: false, message: `Name is too long (max ${maxLength} characters)` };
  }

  if (!isSafeInput(name)) {
    return { valid: false, message: 'Name contains invalid characters' };
  }

  return { valid: true, message: '' };
}

/**
 * Strip HTML tags from input (used for display sanitization)
 * @param {string} input - Input string
 * @returns {string} String without HTML tags
 */
export function stripHtmlTags(input) {
  if (typeof input !== 'string') {
    return '';
  }

  return input.replace(/<[^>]*>/g, '');
}

// Constants for validation
export const MAX_USERNAME_LENGTH = 100;
export const MAX_NAME_LENGTH = 255;
export const MAX_POST_HEADER_LENGTH = 500;
export const MAX_POST_CONTENT_LENGTH = 5000;
export const MAX_PASSWORD_LENGTH = 128;

