-- Interview Platform Database Schema
-- PostgreSQL Database Initialization Script

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create focus_areas table
CREATE TABLE IF NOT EXISTS focus_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for focus_areas
CREATE INDEX IF NOT EXISTS idx_focus_areas_created_at ON focus_areas(created_at);

-- Create problems table
CREATE TABLE IF NOT EXISTS problems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    focus_area_id UUID NOT NULL REFERENCES focus_areas(id) ON DELETE CASCADE,
    sample_cases JSONB NOT NULL,
    hidden_cases JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for problems (for O(log n) complexity)
CREATE INDEX IF NOT EXISTS idx_problems_created_at ON problems(created_at);

-- Create index on focus_area_id for faster joins
CREATE INDEX IF NOT EXISTS idx_problems_focus_area_id ON problems(focus_area_id);

-- Seed initial focus areas
INSERT INTO focus_areas (name, slug) VALUES
    ('Dynamic Programming', 'dynamic-programming'),
    ('Greedy Algorithms', 'greedy'),
    ('Graph Algorithms', 'graphs'),
    ('Tree Algorithms', 'trees'),
    ('Array and String Manipulation', 'arrays-strings'),
    ('Sorting and Searching', 'sorting-searching'),
    ('Backtracking', 'backtracking'),
    ('Bit Manipulation', 'bit-manipulation'),
    ('Mathematics and Number Theory', 'mathematics'),
    ('Sliding Window', 'sliding-window')
ON CONFLICT (slug) DO NOTHING;

-- Verify setup
SELECT 'Database schema created successfully!' AS message;
SELECT COUNT(*) AS focus_areas_count FROM focus_areas;
