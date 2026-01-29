-- Revert trigger rename
ALTER TRIGGER evaluators_updated_at ON evaluators RENAME TO evaluation_rules_updated_at;

-- Revert foreign key column rename
ALTER TABLE evaluator_executions RENAME COLUMN evaluator_id TO rule_id;

-- Revert constraints on evaluators table
ALTER TABLE evaluators RENAME CONSTRAINT evaluators_status_check TO evaluation_rules_status_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluators_trigger_type_check TO evaluation_rules_trigger_type_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluators_target_scope_check TO evaluation_rules_target_scope_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluators_scorer_type_check TO evaluation_rules_scorer_type_check;
ALTER TABLE evaluators RENAME CONSTRAINT evaluators_sampling_rate_check TO evaluation_rules_sampling_rate_check;

-- Revert indexes on evaluator_executions table
ALTER INDEX idx_evaluator_executions_evaluator RENAME TO idx_rule_executions_rule;
ALTER INDEX idx_evaluator_executions_project RENAME TO idx_rule_executions_project;
ALTER INDEX idx_evaluator_executions_status RENAME TO idx_rule_executions_status;
ALTER INDEX idx_evaluator_executions_created RENAME TO idx_rule_executions_created;
ALTER INDEX idx_evaluator_executions_evaluator_created RENAME TO idx_rule_executions_rule_created;

-- Revert indexes on evaluators table
ALTER INDEX idx_evaluators_project_id RENAME TO idx_evaluation_rules_project_id;
ALTER INDEX idx_evaluators_status RENAME TO idx_evaluation_rules_status;
ALTER INDEX idx_evaluators_project_status RENAME TO idx_evaluation_rules_project_status;
ALTER INDEX idx_evaluators_project_name RENAME TO idx_evaluation_rules_project_name;

-- Revert table renames
ALTER TABLE evaluator_executions RENAME TO evaluation_rule_executions;
ALTER TABLE evaluators RENAME TO evaluation_rules;
