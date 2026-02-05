"use server";

import { updateProfileInfo, updateProfileEmail, updateProfilePassword, updateProfilePrivacy } from "@/actions/profile/update-profile";
import { revalidatePath } from "next/cache";
import { isValidEmail, isStrongPassword, calculateAge, USERNAME_PATTERN, ALLOWED_FILE_TYPES, MAX_FILE_SIZE } from "@/lib/validation";

// Update Profile Info
export async function updateProfileAction(prevState, formData) {
    const firstName = formData.get("firstName")?.trim();
    const lastName = formData.get("lastName")?.trim();
    const dateOfBirth = formData.get("dateOfBirth");
    const about = formData.get("about")?.trim();
    const username = formData.get("username")?.trim();

    // Basic Validation
    if (!firstName || firstName.length < 2) {
        return { success: false, message: "First name must be at least 2 characters." };
    }
    if (!lastName || lastName.length < 2) {
        return { success: false, message: "Last name must be at least 2 characters." };
    }

    if (dateOfBirth) {
        const age = calculateAge(dateOfBirth);
        if (age < 13 || age > 111) {
            return { success: false, message: "You must be between 13 and 111 years old." };
        }
    }

    if (about && about.length > 400) {
        return { success: false, message: "About me must be at most 400 characters." };
    }

    if (username) {
        if (username.length < 4) {
            return { success: false, message: "Username must be at least 4 characters." };
        }
        if (!USERNAME_PATTERN.test(username)) {
            return { success: false, message: "Invalid username format." };
        }
    }

    // Map to service params
    const payload = {
        first_name: firstName,
        last_name: lastName,
        date_of_birth: dateOfBirth,
        about: about,
        username: username
    };

    try {
        const result = await updateProfileInfo(payload);

        if (!result.success) {
            return { success: false, message: result.error || "Failed to update profile." };
        }

        revalidatePath("/profile/[id]", "layout");
        return { success: true, message: "Profile updated successfully." };
    } catch (error) {
        return { success: false, message: error.message || "An unexpected error occurred." };
    }
}

// Update Email
export async function updateEmailAction(prevState, formData) {
    const email = formData.get("email")?.trim();

    if (!isValidEmail(email)) {
        return { success: false, message: "Please enter a valid email address." };
    }

    try {
        const result = await updateProfileEmail({ email });

        if (!result.success) {
            return { success: false, message: result.error || "Failed to update email." };
        }

        revalidatePath("/profile/[id]", "layout");
        return { success: true, message: "Email updated successfully. Please verify your new email." };
    } catch (error) {
        return { success: false, message: error.message || "An unexpected error occurred." };
    }
}

// Update Password
export async function updatePasswordAction(prevState, formData) {
    const currentPassword = formData.get("currentPassword");
    const newPassword = formData.get("newPassword");
    const confirmPassword = formData.get("confirmPassword");

    if (!currentPassword || !newPassword || !confirmPassword) {
        return { success: false, message: "All password fields are required." };
    }

    if (newPassword !== confirmPassword) {
        return { success: false, message: "New passwords do not match." };
    }

    if (!isStrongPassword(newPassword)) {
        return { success: false, message: "Password needs 1 lowercase, 1 uppercase, 1 number, and 1 symbol." };
    }

    try {
        const result = await updateProfilePassword({
            oldPassword: currentPassword,
            newPassword: newPassword
        });

        if (!result.success) {
            return { success: false, message: result.error || "Failed to update password." };
        }

        return { success: true, message: "Password updated successfully." };
    } catch (error) {
        return { success: false, message: error.message || "An unexpected error occurred." };
    }
}

// Update Privacy
export async function updatePrivacyAction(bool) {
    try {
        const result = await updateProfilePrivacy({ bool });

        if (!result.success) {
            return { success: false, message: result.error || "Failed to update privacy settings." };
        }

        revalidatePath("/profile/[id]", "layout");
        return { success: true, message: "Privacy settings updated." };
    } catch (error) {
        return { success: false, message: "An unexpected error occurred." };
    }
}
