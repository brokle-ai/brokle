-- Drop indexes
DROP INDEX IF EXISTS idx_onboarding_questions_step_active;
DROP INDEX IF EXISTS idx_onboarding_questions_display_order;
DROP INDEX IF EXISTS idx_user_onboarding_responses_user_id;
DROP INDEX IF EXISTS idx_user_onboarding_responses_question_id;

-- Drop trigger
DROP TRIGGER IF EXISTS update_onboarding_questions_updated_at ON onboarding_questions;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS user_onboarding_responses;
DROP TABLE IF EXISTS onboarding_questions;