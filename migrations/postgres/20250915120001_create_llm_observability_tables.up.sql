-- Create LLM observability tables for comprehensive trace and observation tracking

-- Create llm_traces table for trace management
CREATE TABLE llm_traces (
    id CHAR(26) PRIMARY KEY,
    project_id CHAR(26) NOT NULL,
    session_id CHAR(26) NULL,
    external_trace_id VARCHAR(255) NOT NULL,
    parent_trace_id CHAR(26) REFERENCES llm_traces(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    user_id CHAR(26) NULL,
    tags JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create unique constraint on external_trace_id for SDK integration
CREATE UNIQUE INDEX idx_llm_traces_external_trace_id ON llm_traces(external_trace_id);

-- Create indexes for performance optimization
CREATE INDEX idx_llm_traces_project_id ON llm_traces(project_id);
CREATE INDEX idx_llm_traces_session_id ON llm_traces(session_id);
CREATE INDEX idx_llm_traces_user_id ON llm_traces(user_id);
CREATE INDEX idx_llm_traces_parent_trace_id ON llm_traces(parent_trace_id);
CREATE INDEX idx_llm_traces_created_at ON llm_traces(created_at DESC);
CREATE INDEX idx_llm_traces_updated_at ON llm_traces(updated_at DESC);
CREATE INDEX idx_llm_traces_name ON llm_traces(name);

-- Create GIN index for JSONB columns for efficient tag and metadata queries
CREATE INDEX idx_llm_traces_tags ON llm_traces USING GIN (tags);
CREATE INDEX idx_llm_traces_metadata ON llm_traces USING GIN (metadata);

-- Create llm_observations table for detailed observation tracking
CREATE TABLE llm_observations (
    id CHAR(26) PRIMARY KEY,
    trace_id CHAR(26) NOT NULL REFERENCES llm_traces(id) ON DELETE CASCADE,
    external_observation_id VARCHAR(255) NOT NULL,
    parent_observation_id CHAR(26) REFERENCES llm_observations(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('llm', 'span', 'event', 'generation', 'retrieval', 'embedding', 'agent', 'tool', 'chain')),
    name VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NULL,
    level VARCHAR(20) DEFAULT 'INFO' CHECK (level IN ('DEBUG', 'INFO', 'WARN', 'ERROR', 'DEFAULT')),
    status_message TEXT NULL,
    version VARCHAR(50) NULL,
    model VARCHAR(255) NULL,
    provider VARCHAR(100) NULL,
    input JSONB NULL,
    output JSONB NULL,
    model_parameters JSONB DEFAULT '{}',
    prompt_tokens INTEGER DEFAULT 0 CHECK (prompt_tokens >= 0),
    completion_tokens INTEGER DEFAULT 0 CHECK (completion_tokens >= 0),
    total_tokens INTEGER DEFAULT 0 CHECK (total_tokens >= 0),
    input_cost DECIMAL(12,8) NULL CHECK (input_cost >= 0),
    output_cost DECIMAL(12,8) NULL CHECK (output_cost >= 0),
    total_cost DECIMAL(12,8) NULL CHECK (total_cost >= 0),
    latency_ms INTEGER NULL CHECK (latency_ms >= 0),
    quality_score DECIMAL(3,2) NULL CHECK (quality_score >= 0 AND quality_score <= 1),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create unique constraint on external_observation_id for SDK integration
CREATE UNIQUE INDEX idx_llm_observations_external_obs_id ON llm_observations(external_observation_id);

-- Create indexes for performance optimization
CREATE INDEX idx_llm_observations_trace_id ON llm_observations(trace_id);
CREATE INDEX idx_llm_observations_parent_observation_id ON llm_observations(parent_observation_id);
CREATE INDEX idx_llm_observations_type ON llm_observations(type);
CREATE INDEX idx_llm_observations_provider ON llm_observations(provider);
CREATE INDEX idx_llm_observations_model ON llm_observations(model);
CREATE INDEX idx_llm_observations_provider_model ON llm_observations(provider, model);
CREATE INDEX idx_llm_observations_start_time ON llm_observations(start_time DESC);
CREATE INDEX idx_llm_observations_end_time ON llm_observations(end_time DESC);
CREATE INDEX idx_llm_observations_level ON llm_observations(level);
CREATE INDEX idx_llm_observations_created_at ON llm_observations(created_at DESC);

-- Create composite indexes for common query patterns
CREATE INDEX idx_llm_observations_trace_type ON llm_observations(trace_id, type);
CREATE INDEX idx_llm_observations_trace_start_time ON llm_observations(trace_id, start_time DESC);
CREATE INDEX idx_llm_observations_provider_start_time ON llm_observations(provider, start_time DESC);

-- Create indexes for cost and performance analysis
CREATE INDEX idx_llm_observations_total_cost ON llm_observations(total_cost DESC) WHERE total_cost IS NOT NULL;
CREATE INDEX idx_llm_observations_latency_ms ON llm_observations(latency_ms DESC) WHERE latency_ms IS NOT NULL;
CREATE INDEX idx_llm_observations_quality_score ON llm_observations(quality_score DESC) WHERE quality_score IS NOT NULL;
CREATE INDEX idx_llm_observations_total_tokens ON llm_observations(total_tokens DESC) WHERE total_tokens > 0;

-- Create GIN indexes for JSONB columns
CREATE INDEX idx_llm_observations_input ON llm_observations USING GIN (input);
CREATE INDEX idx_llm_observations_output ON llm_observations USING GIN (output);
CREATE INDEX idx_llm_observations_model_parameters ON llm_observations USING GIN (model_parameters);

-- Create llm_quality_scores table for quality evaluation and scoring
CREATE TABLE llm_quality_scores (
    id CHAR(26) PRIMARY KEY,
    trace_id CHAR(26) NOT NULL REFERENCES llm_traces(id) ON DELETE CASCADE,
    observation_id CHAR(26) NULL REFERENCES llm_observations(id) ON DELETE CASCADE,
    score_name VARCHAR(100) NOT NULL,
    score_value DECIMAL(10,6) NULL,
    string_value TEXT NULL,
    data_type VARCHAR(20) DEFAULT 'NUMERIC' CHECK (data_type IN ('NUMERIC', 'CATEGORICAL', 'BOOLEAN')),
    source VARCHAR(50) DEFAULT 'API' CHECK (source IN ('API', 'AUTO', 'HUMAN', 'EVAL')),
    evaluator_name VARCHAR(100) NULL,
    evaluator_version VARCHAR(50) NULL,
    comment TEXT NULL,
    author_user_id CHAR(26) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Ensure either score_value or string_value is provided based on data_type
    CONSTRAINT chk_quality_score_value CHECK (
        (data_type = 'NUMERIC' AND score_value IS NOT NULL) OR
        (data_type = 'CATEGORICAL' AND string_value IS NOT NULL) OR
        (data_type = 'BOOLEAN' AND score_value IS NOT NULL)
    )
);

-- Create unique constraint to prevent duplicate scores for same trace/observation + score_name
CREATE UNIQUE INDEX idx_llm_quality_scores_unique_trace_score
ON llm_quality_scores(trace_id, score_name)
WHERE observation_id IS NULL;

CREATE UNIQUE INDEX idx_llm_quality_scores_unique_obs_score
ON llm_quality_scores(observation_id, score_name)
WHERE observation_id IS NOT NULL;

-- Create indexes for performance optimization
CREATE INDEX idx_llm_quality_scores_trace_id ON llm_quality_scores(trace_id);
CREATE INDEX idx_llm_quality_scores_observation_id ON llm_quality_scores(observation_id);
CREATE INDEX idx_llm_quality_scores_score_name ON llm_quality_scores(score_name);
CREATE INDEX idx_llm_quality_scores_source ON llm_quality_scores(source);
CREATE INDEX idx_llm_quality_scores_data_type ON llm_quality_scores(data_type);
CREATE INDEX idx_llm_quality_scores_evaluator_name ON llm_quality_scores(evaluator_name);
CREATE INDEX idx_llm_quality_scores_author_user_id ON llm_quality_scores(author_user_id);
CREATE INDEX idx_llm_quality_scores_created_at ON llm_quality_scores(created_at DESC);

-- Create indexes for score analysis
CREATE INDEX idx_llm_quality_scores_score_value ON llm_quality_scores(score_value DESC) WHERE score_value IS NOT NULL;
CREATE INDEX idx_llm_quality_scores_score_name_value ON llm_quality_scores(score_name, score_value DESC) WHERE score_value IS NOT NULL;

-- Create composite indexes for common query patterns
CREATE INDEX idx_llm_quality_scores_trace_name ON llm_quality_scores(trace_id, score_name);
CREATE INDEX idx_llm_quality_scores_obs_name ON llm_quality_scores(observation_id, score_name);

-- Create updated_at trigger function for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic updated_at timestamp updates
CREATE TRIGGER trigger_llm_traces_updated_at
    BEFORE UPDATE ON llm_traces
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_llm_observations_updated_at
    BEFORE UPDATE ON llm_observations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_llm_quality_scores_updated_at
    BEFORE UPDATE ON llm_quality_scores
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to calculate observation latency automatically
CREATE OR REPLACE FUNCTION calculate_observation_latency()
RETURNS TRIGGER AS $$
BEGIN
    -- Only calculate latency if end_time is set and start_time exists
    IF NEW.end_time IS NOT NULL AND NEW.start_time IS NOT NULL THEN
        NEW.latency_ms = EXTRACT(MILLISECONDS FROM (NEW.end_time - NEW.start_time))::INTEGER;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically calculate latency on observation updates
CREATE TRIGGER trigger_llm_observations_calculate_latency
    BEFORE INSERT OR UPDATE ON llm_observations
    FOR EACH ROW
    EXECUTE FUNCTION calculate_observation_latency();

-- Create function to automatically calculate total_tokens if not provided
CREATE OR REPLACE FUNCTION calculate_total_tokens()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate total_tokens if not explicitly provided
    IF NEW.total_tokens = 0 AND (NEW.prompt_tokens > 0 OR NEW.completion_tokens > 0) THEN
        NEW.total_tokens = COALESCE(NEW.prompt_tokens, 0) + COALESCE(NEW.completion_tokens, 0);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically calculate total tokens
CREATE TRIGGER trigger_llm_observations_calculate_total_tokens
    BEFORE INSERT OR UPDATE ON llm_observations
    FOR EACH ROW
    EXECUTE FUNCTION calculate_total_tokens();

-- Create function to automatically calculate total_cost if not provided
CREATE OR REPLACE FUNCTION calculate_total_cost()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate total_cost if not explicitly provided
    IF NEW.total_cost IS NULL AND (NEW.input_cost IS NOT NULL OR NEW.output_cost IS NOT NULL) THEN
        NEW.total_cost = COALESCE(NEW.input_cost, 0) + COALESCE(NEW.output_cost, 0);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically calculate total cost
CREATE TRIGGER trigger_llm_observations_calculate_total_cost
    BEFORE INSERT OR UPDATE ON llm_observations
    FOR EACH ROW
    EXECUTE FUNCTION calculate_total_cost();

-- Add comments for documentation
COMMENT ON TABLE llm_traces IS 'Stores LLM trace information for observability tracking';
COMMENT ON TABLE llm_observations IS 'Stores individual observations within traces (LLM calls, spans, events)';
COMMENT ON TABLE llm_quality_scores IS 'Stores quality evaluation scores for traces and observations';

COMMENT ON COLUMN llm_traces.external_trace_id IS 'External trace ID from SDK for client-server correlation';
COMMENT ON COLUMN llm_observations.external_observation_id IS 'External observation ID from SDK for client-server correlation';
COMMENT ON COLUMN llm_observations.type IS 'Type of observation: llm, span, event, generation, retrieval, embedding, agent, tool, chain';
COMMENT ON COLUMN llm_observations.latency_ms IS 'Automatically calculated latency in milliseconds from start_time to end_time';
COMMENT ON COLUMN llm_observations.total_tokens IS 'Automatically calculated as prompt_tokens + completion_tokens if not provided';
COMMENT ON COLUMN llm_observations.total_cost IS 'Automatically calculated as input_cost + output_cost if not provided';
COMMENT ON COLUMN llm_quality_scores.data_type IS 'Type of score data: NUMERIC, CATEGORICAL, or BOOLEAN';
COMMENT ON COLUMN llm_quality_scores.source IS 'Source of the score: API, AUTO, HUMAN, or EVAL';