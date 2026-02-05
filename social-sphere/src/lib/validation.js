/**
 * Shared validation utilities for client-side and server-side validation
 */

// Email validation pattern
export const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// Password strength pattern: at least 1 lowercase, 1 uppercase, 1 number, 1 symbol
export const STRONG_PASSWORD_PATTERN = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^\w\s]).+$/;

// username pattern: letters, numbers, dots, underscores, dashes
export const USERNAME_PATTERN = /^[A-Za-z0-9_.-]+$/;

// File validation constants (matching backend FileConstraints)
export const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB
export const MAX_IMAGE_WIDTH = 4096;
export const MAX_IMAGE_HEIGHT = 4096;
export const ALLOWED_FILE_TYPES = ["image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"];
export const ALLOWED_FILE_ACCEPT = "image/jpeg,image/png,image/gif,image/webp";

/**
 * Calculate age from date of birth
 * @param {string} dateOfBirth - Date string in YYYY-MM-DD format
 * @returns {number} Age in years
 */
export function calculateAge(dateOfBirth) {
    const today = new Date();
    const birthDate = new Date(dateOfBirth);
    let age = today.getFullYear() - birthDate.getFullYear();
    const monthDiff = today.getMonth() - birthDate.getMonth();

    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
        age--;
    }

    return age;
}

/**
 * Validate email format
 * @param {string} email - Email to validate
 * @returns {boolean} True if valid
 */
export function isValidEmail(email) {
    return EMAIL_PATTERN.test(email);
}

/**
 * Validate password strength
 * @param {string} password - Password to validate
 * @returns {boolean} True if meets strength requirements
 */
export function isStrongPassword(password) {
    return password.length >= 8 && STRONG_PASSWORD_PATTERN.test(password);
}

/**
 * Validate username format
 * @param {string} username - Username to validate
 * @returns {boolean} True if valid
 */
export function isValidUsername(username) {
    return username.length >= 4 && USERNAME_PATTERN.test(username);
}


/**
 * Validate registration form data (client-side)
 * @param {FormData} formData - Form data to validate
 * @param {File|null} avatarFile - Avatar file object
 * @returns {Promise<{valid: boolean, error: string}>} Validation result
 */
export async function validateRegistrationForm(formData, avatarFile = null) {
    // First name validation
    const firstName = formData.get("first_name")?.trim() || "";
    if (!firstName) {
        return { valid: false, error: "First name is required." };
    }
    if (firstName.length < 2) {
        return { valid: false, error: "First name must be at least 2 characters." };
    }

    // Last name validation
    const lastName = formData.get("last_name")?.trim() || "";
    if (!lastName) {
        return { valid: false, error: "Last name is required." };
    }
    if (lastName.length < 2) {
        return { valid: false, error: "Last name must be at least 2 characters." };
    }

    // Email validation
    const email = formData.get("email")?.trim() || "";
    if (!isValidEmail(email)) {
        return { valid: false, error: "Please enter a valid email address." };
    }

    // Password validation
    const password = formData.get("password");
    const confirmPassword = formData.get("confirmPassword");
    if (!password || !confirmPassword) {
        return { valid: false, error: "Please enter both password and confirm password." };
    }
    if (password.length < 8) {
        return { valid: false, error: "Password must be at least 8 characters." };
    }
    if (!STRONG_PASSWORD_PATTERN.test(password)) {
        return { valid: false, error: "Password needs 1 lowercase, 1 uppercase, 1 number, and 1 symbol." };
    }
    if (password !== confirmPassword) {
        return { valid: false, error: "Passwords do not match" };
    }

    // Date of birth validation
    const dateOfBirth = formData.get("date_of_birth")?.trim() || "";
    if (!dateOfBirth) {
        return { valid: false, error: "Date of birth is required." };
    }
    const age = calculateAge(dateOfBirth);
    if (age < 13 || age > 111) {
        return { valid: false, error: "You must be between 13 and 111 years old." };
    }

    // username validation (optional)
    const username = formData.get("username")?.trim() || "";
    if (username) {
        if (username.length < 4) {
            return { valid: false, error: "Username must be at least 4 characters." };
        }
        if (!USERNAME_PATTERN.test(username)) {
            return { valid: false, error: "Username can only use letters, numbers, dots, underscores, or dashes." };
        }
    }

    // About me validation (optional)
    const aboutMe = formData.get("about")?.trim() || "";
    if (aboutMe && aboutMe.length > 400) {
        return { valid: false, error: "About me must be at most 400 characters." };
    }

    // Avatar validation (optional) - includes dimension check
    if (avatarFile) {
        const avatarValidation = await validateImage(avatarFile);
        if (!avatarValidation.valid) {
            return avatarValidation;
        }
    }

    return { valid: true, error: "" };
}

/**
 * Validate profile update form data (client-side)
 * @param {Object} profileData - Profile data object
 * @param {File|null} avatarFile - Avatar file object
 * @returns {Promise<{valid: boolean, error: string}>} Validation result
 */
export async function validateProfileForm(profileData, avatarFile = null) {
    // First name validation
    const firstName = profileData.first_name?.trim() || "";
    if (!firstName) {
        return { valid: false, error: "First name is required." };
    }
    if (firstName.length < 2) {
        return { valid: false, error: "First name must be at least 2 characters." };
    }

    // Last name validation
    const lastName = profileData.last_name?.trim() || "";
    if (!lastName) {
        return { valid: false, error: "Last name is required." };
    }
    if (lastName.length < 2) {
        return { valid: false, error: "Last name must be at least 2 characters." };
    }

    // Date of birth validation
    const dateOfBirth = profileData.date_of_birth?.trim() || "";
    if (!dateOfBirth) {
        return { valid: false, error: "Date of birth is required." };
    }
    const age = calculateAge(dateOfBirth);
    if (age < 13 || age > 111) {
        return { valid: false, error: "You must be between 13 and 111 years old." };
    }

    // Username validation (optional)
    const username = profileData.username?.trim() || "";
    if (username) {
        if (username.length < 4) {
            return { valid: false, error: "Username must be at least 4 characters." };
        }
        if (!USERNAME_PATTERN.test(username)) {
            return { valid: false, error: "Username can only use letters, numbers, dots, underscores, or dashes." };
        }
    }

    // About me validation (optional)
    const about = profileData.about?.trim() || "";
    if (about && about.length > 400) {
        return { valid: false, error: "About me must be at most 400 characters." };
    }

    // Avatar validation (optional) - includes dimension check
    if (avatarFile) {
        const avatarValidation = await validateImage(avatarFile);
        if (!avatarValidation.valid) {
            return avatarValidation;
        }
    }

    return { valid: true, error: "" };
}

/**
 * Validate login form data (client-side)
 * @param {FormData} formData - Form data to validate
 * @returns {{valid: boolean, error: string}} Validation result
 */
export function validateLoginForm(formData) {
    // Identifier validation
    const identifier = formData.get("identifier")?.trim() || "";
    if (!identifier) {
        return { valid: false, error: "Email or username is required." };
    }

    // Password validation
    const password = formData.get("password");
    if (!password) {
        return { valid: false, error: "Password is required." };
    }

    return { valid: true, error: "" };
}

/**
 * Validate post content
 * @param {string} content - Post content to validate
 * @param {number} minChars - Minimum character count (default: 1)
 * @param {number} maxChars - Maximum character count (default: 5000)
 * @returns {{valid: boolean, error: string}} Validation result
 */
export function validatePostContent(content, minChars = 1, maxChars = 5000) {
    const trimmed = content?.trim() || "";

    if (!trimmed) {
        return { valid: false, error: "Post content is required." };
    }

    if (trimmed.length < minChars) {
        return { valid: false, error: `Post must be at least ${minChars} character${minChars > 1 ? 's' : ''}.` };
    }

    if (trimmed.length > maxChars) {
        return { valid: false, error: `Post must be at most ${maxChars} characters.` };
    }

    return { valid: true, error: "" };
}

/**
 * Validate image file for posts
 * @param {File} file - File object to validate
 * @returns {{valid: boolean, error: string}} Validation result
 */
export function isValidImage(file) {
    // Image is optional - if no file provided, it's valid
    if (!file || file.size === 0) {
        return { valid: true, error: "" };
    }

    if (!ALLOWED_FILE_TYPES.includes(file.type)) {
        return { valid: false, error: "Image must be JPEG, PNG, GIF, or WebP." };
    }

    if (file.size > MAX_FILE_SIZE) {
        return { valid: false, error: "Image must be less than 5MB." };
    }

    return { valid: true, error: "" };
}

/**
 * Validate image file including dimensions (async)
 * @param {File} file - File object to validate
 * @returns {Promise<{valid: boolean, error: string}>} Validation result
 */
export async function validateImage(file) {
    // Run sync validations first
    const syncResult = isValidImage(file);
    if (!syncResult.valid) {
        return syncResult;
    }

    // If no file, skip dimension check
    if (!file || file.size === 0) {
        return { valid: true, error: "" };
    }

    // Check dimensions
    try {
        const dimensions = await getImageDimensions(file);
        if (dimensions.width > MAX_IMAGE_WIDTH || dimensions.height > MAX_IMAGE_HEIGHT) {
            return { valid: false, error: `Image dimensions must be at most ${MAX_IMAGE_WIDTH}x${MAX_IMAGE_HEIGHT} pixels.` };
        }
    } catch {
        return { valid: false, error: "Could not read image dimensions." };
    }

    return { valid: true, error: "" };
}

/**
 * Get image dimensions from a File object
 * @param {File} file - Image file
 * @returns {Promise<{width: number, height: number}>}
 */
function getImageDimensions(file) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => {
            resolve({ width: img.naturalWidth, height: img.naturalHeight });
            URL.revokeObjectURL(img.src);
        };
        img.onerror = () => {
            URL.revokeObjectURL(img.src);
            reject(new Error("Failed to load image"));
        };
        img.src = URL.createObjectURL(file);
    });
}

