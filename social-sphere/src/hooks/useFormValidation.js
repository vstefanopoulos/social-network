import { useState, useCallback } from 'react';

/**
 * Custom hook for form validation
 * @param {Object} initialErrors - Initial error state
 * @returns {Object} Validation hook methods and state
 */
export function useFormValidation(initialErrors = {}) {
    const [errors, setErrors] = useState(initialErrors);

    /**
     * Validate a single field
     * @param {string} name - Field name
     * @param {any} value - Field value
     * @param {Function} validator - Validation function that returns error string or null/undefined
     */
    const validateField = useCallback((name, value, validator) => {
        const error = validator(value);

        setErrors(prev => {
            const newErrors = { ...prev };
            if (error) {
                newErrors[name] = error;
            } else {
                delete newErrors[name];
            }
            return newErrors;
        });

        return !error;
    }, []);

    return {
        errors,
        validateField,
    };
}
