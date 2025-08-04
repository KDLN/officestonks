-- Create basic sectors to resolve foreign key constraint issues
-- This fixes Error 1452: Cannot add or update a child row: a foreign key constraint fails

-- First, let's see what the sectors table structure is and insert basic sectors
INSERT INTO sectors (id, name, description) VALUES 
(1, 'Technology', 'Technology and software companies'),
(2, 'Healthcare', 'Healthcare and pharmaceutical companies'),
(3, 'Finance', 'Financial services and banking companies'),
(4, 'Energy', 'Energy and oil companies'),
(5, 'Consumer', 'Consumer goods and retail companies'),
(6, 'Industrial', 'Industrial and manufacturing companies'),
(7, 'Real Estate', 'Real estate and construction companies'),
(8, 'Utilities', 'Utility and infrastructure companies'),
(9, 'Materials', 'Materials and mining companies'),
(10, 'Communications', 'Telecommunications and media companies')
ON DUPLICATE KEY UPDATE 
name = VALUES(name), 
description = VALUES(description);

-- Alternative if description column doesn't exist:
-- INSERT INTO sectors (id, name) VALUES 
-- (1, 'Technology'),
-- (2, 'Healthcare'), 
-- (3, 'Finance'),
-- (4, 'Energy'),
-- (5, 'Consumer'),
-- (6, 'Industrial'),
-- (7, 'Real Estate'),
-- (8, 'Utilities'),
-- (9, 'Materials'),
-- (10, 'Communications')
-- ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Verify the sectors were created
SELECT * FROM sectors ORDER BY id;