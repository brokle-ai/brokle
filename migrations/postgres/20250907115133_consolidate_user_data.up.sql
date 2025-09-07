-- ===================================
-- CONSOLIDATE USER DATA MIGRATION
-- ===================================
-- 
-- This migration consolidates user data by:
-- 1. Moving avatar_url and phone from users to user_profiles
-- 2. Moving all notification preferences from user_preferences to user_profiles
-- 3. Dropping the user_preferences table
-- 4. Eliminating duplicate fields between tables

-- Add new columns to user_profiles table
ALTER TABLE user_profiles ADD COLUMN avatar_url VARCHAR(500);
ALTER TABLE user_profiles ADD COLUMN phone VARCHAR(50);
ALTER TABLE user_profiles ADD COLUMN email_notifications BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN push_notifications BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN marketing_emails BOOLEAN DEFAULT false;
ALTER TABLE user_profiles ADD COLUMN weekly_reports BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN monthly_reports BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN security_alerts BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN billing_alerts BOOLEAN DEFAULT true;
ALTER TABLE user_profiles ADD COLUMN usage_threshold_percent INTEGER DEFAULT 80;

-- Migrate existing data from users table to user_profiles
UPDATE user_profiles SET 
    avatar_url = u.avatar_url,
    phone = u.phone
FROM users u 
WHERE user_profiles.user_id = u.id 
AND (u.avatar_url IS NOT NULL OR u.phone IS NOT NULL);

-- Migrate existing data from user_preferences table to user_profiles
UPDATE user_profiles SET 
    email_notifications = p.email_notifications,
    push_notifications = p.push_notifications,
    marketing_emails = p.marketing_emails,
    weekly_reports = p.weekly_reports,
    monthly_reports = p.monthly_reports,
    security_alerts = p.security_alerts,
    billing_alerts = p.billing_alerts,
    usage_threshold_percent = p.usage_threshold_percent
FROM user_preferences p 
WHERE user_profiles.user_id = p.user_id;

-- Create user_profiles for users who don't have them yet (from registration bug)
INSERT INTO user_profiles (
    user_id, timezone, language, theme, 
    avatar_url, phone,
    email_notifications, push_notifications, marketing_emails,
    weekly_reports, monthly_reports, security_alerts, billing_alerts,
    usage_threshold_percent,
    created_at, updated_at
)
SELECT 
    u.id,
    COALESCE(u.timezone, 'UTC'),
    COALESCE(u.language, 'en'),
    'light',
    u.avatar_url,
    u.phone,
    COALESCE(p.email_notifications, true),
    COALESCE(p.push_notifications, true),
    COALESCE(p.marketing_emails, false),
    COALESCE(p.weekly_reports, true),
    COALESCE(p.monthly_reports, true),
    COALESCE(p.security_alerts, true),
    COALESCE(p.billing_alerts, true),
    COALESCE(p.usage_threshold_percent, 80),
    NOW(),
    NOW()
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id
LEFT JOIN user_preferences p ON u.id = p.user_id
WHERE up.user_id IS NULL;  -- Only for users without profiles

-- Remove redundant columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS phone;

-- Drop user_preferences table completely
DROP TABLE IF EXISTS user_preferences;

-- Create indexes for better performance on new notification columns
CREATE INDEX idx_user_profiles_email_notifications ON user_profiles(email_notifications);
CREATE INDEX idx_user_profiles_security_alerts ON user_profiles(security_alerts);
CREATE INDEX idx_user_profiles_avatar_url ON user_profiles(avatar_url) WHERE avatar_url IS NOT NULL;
CREATE INDEX idx_user_profiles_phone ON user_profiles(phone) WHERE phone IS NOT NULL;

-- Update existing indexes if needed (remove old preference indexes)
DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP INDEX IF EXISTS idx_user_preferences_theme;
DROP INDEX IF EXISTS idx_user_preferences_email_notifications;