-- Rename evaluation_rules table to evaluators
ALTER TABLE evaluation_rules RENAME TO evaluators;

-- Rename evaluation_rule_executions table to evaluator_executions
ALTER TABLE evaluation_rule_executions RENAME TO evaluator_executions;

-- Rename indexes on evaluators table
ALTER INDEX idx_evaluation_rules_project_id RENAME TO idx_evaluators_project_id;
ALTER INDEX idx_evaluation_rules_status RENAME TO idx_evaluators_status;
ALTER INDEX idx_evaluation_rules_project_status RENAME TO idx_evaluators_project_status;
ALTER INDEX idx_evaluation_rules_project_name RENAME TO idx_evaluators_project_name;

-- Rename indexes on evaluator_executions table
ALTER INDEX idx_rule_executions_rule RENAME TO idx_evaluator_executions_evaluator;
ALTER INDEX idx_rule_executions_project RENAME TO idx_evaluator_executions_project;
ALTER INDEX idx_rule_executions_status RENAME TO idx_evaluator_executions_status;
ALTER INDEX idx_rule_executions_created RENAME TO idx_evaluator_executions_created;
ALTER INDEX idx_rule_executions_rule_created RENAME TO idx_evaluator_executions_evaluator_created;

-- Rename constraints on evaluators table
ALTER TABLE evaluators RENAME CONSTRAINT evaluation_rules_status_check TO evaluators_status_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluation_rules_trigger_type_check TO evaluators_trigger_type_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluation_rules_target_scope_check TO evaluators_target_scope_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluation_rules_scorer_type_check TO evaluators_scorer_type_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluation_rules_sampling_rate_check TO evaluators_sampling_rate_check;

-- Rename foreign key column in evaluator_executions
ALTER TABLE evaluator_executions RENAME COLUMN rule_id TO evaluator_id;

-- Rename trigger on evaluators table
ALTER TRIGGER evaluation_rules_updated_at ON evaluators RENAME TO evaluators_updated_at;
