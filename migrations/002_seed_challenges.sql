-- Seed data for DevArena challenges
-- Run this script after the initial migration

-- First, insert tags (created_at uses DEFAULT NOW())
INSERT INTO tags (id, name, slug, category, color) VALUES
    ('tag-css', 'CSS', 'css', 'frontend', '#264de4'),
    ('tag-layout', 'Layout', 'layout', 'frontend', '#38bdf8'),
    ('tag-react', 'React', 'react', 'frontend', '#61dafb'),
    ('tag-hooks', 'Hooks', 'hooks', 'frontend', '#00d8ff'),
    ('tag-dom', 'DOM', 'dom', 'frontend', '#f0db4f'),
    ('tag-html', 'HTML', 'html', 'frontend', '#e34c26'),
    ('tag-accessibility', 'Accessibility', 'accessibility', 'frontend', '#4CAF50'),
    ('tag-nodejs', 'Node.js', 'nodejs', 'backend', '#339933'),
    ('tag-express', 'Express', 'express', 'backend', '#000000'),
    ('tag-system-design', 'System Design', 'system-design', 'backend', '#9c27b0'),
    ('tag-algorithms', 'Algorithms', 'algorithms', 'fundamentals', '#ff5722'),
    ('tag-sql', 'SQL', 'sql', 'database', '#4479a1'),
    ('tag-databases', 'Databases', 'databases', 'database', '#336791'),
    ('tag-nlp', 'NLP', 'nlp', 'ai', '#ff6f00'),
    ('tag-python', 'Python', 'python', 'backend', '#3776ab'),
    ('tag-tensorflow', 'TensorFlow', 'tensorflow', 'ai', '#ff6f00'),
    ('tag-cv', 'CV', 'cv', 'ai', '#ff9800'),
    ('tag-debug', 'Debug', 'debug', 'fundamentals', '#f44336'),
    ('tag-auth', 'Auth', 'auth', 'backend', '#673ab7'),
    ('tag-golang', 'Golang', 'golang', 'backend', '#00add8'),
    ('tag-goroutines', 'Goroutines', 'goroutines', 'backend', '#00add8'),
    ('tag-redux', 'Redux', 'redux', 'frontend', '#764abc'),
    ('tag-state', 'State', 'state', 'frontend', '#764abc'),
    ('tag-3d', '3D', '3d', 'frontend', '#000000'),
    ('tag-canvas', 'Canvas', 'canvas', 'frontend', '#e535ab'),
    ('tag-ssg', 'SSG', 'ssg', 'frontend', '#000000'),
    ('tag-ssr', 'SSR', 'ssr', 'frontend', '#000000')
ON CONFLICT (id) DO NOTHING;

-- Insert challenges with full details
INSERT INTO challenges (id, title, description, difficulty, type, max_score, repo_template_url, requirements, tech_stack, estimated_hours, is_published, created_at, updated_at) VALUES
(
    'challenge-1',
    'CSS Flexbox Froggy',
    'Master CSS Flexbox by completing a series of layout challenges. You will learn how to use flex-direction, justify-content, align-items, flex-wrap, and more to position elements on a page.',
    'Easy',
    'project',
    10,
    'https://github.com/devarena/css-flexbox-starter',
    '["Use only CSS Flexbox properties", "No use of position: absolute", "Must be responsive", "Pass all layout tests"]',
    '["CSS", "Flexbox", "HTML"]',
    1,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-2',
    'React Counter',
    'Build a simple counter application using React hooks. Implement increment, decrement, and reset functionality while maintaining clean component structure.',
    'Easy',
    'project',
    15,
    'https://github.com/devarena/react-counter-starter',
    '["Use useState hook", "Implement increment/decrement/reset", "Add keyboard shortcuts", "Style with CSS modules"]',
    '["React", "Hooks", "JavaScript"]',
    1,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-3',
    'Infinite Scroll',
    'Implement infinite scroll functionality in a React application. Load more data as the user scrolls to the bottom of the page, with proper loading states and error handling.',
    'Medium',
    'feature',
    30,
    'https://github.com/devarena/infinite-scroll-starter',
    '["Use Intersection Observer API", "Handle loading states", "Implement error boundaries", "Add skeleton loading", "Optimize performance with virtualization"]',
    '["React", "Intersection Observer", "JavaScript", "CSS"]',
    3,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-4',
    'Custom Dropdown',
    'Create an accessible custom dropdown component. It should support keyboard navigation, screen readers, and follow WAI-ARIA guidelines.',
    'Medium',
    'project',
    25,
    'https://github.com/devarena/custom-dropdown-starter',
    '["Support keyboard navigation", "Implement ARIA attributes", "Handle focus management", "Support multi-select option", "Mobile friendly"]',
    '["HTML", "CSS", "JavaScript", "ARIA"]',
    2,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-5',
    'Node.js File Upload',
    'Build a file upload service with Node.js and Express. Support multiple file uploads, file validation, progress tracking, and storage to local filesystem or cloud.',
    'Medium',
    'project',
    40,
    'https://github.com/devarena/nodejs-file-upload-starter',
    '["Handle multipart form data", "Validate file types and sizes", "Implement progress tracking", "Add error handling", "Support multiple files", "Add cloud storage option"]',
    '["Node.js", "Express", "Multer", "AWS S3"]',
    4,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-6',
    'API Rate Limiter',
    'Design and implement an API rate limiter that can handle multiple strategies: fixed window, sliding window, token bucket. Must be distributed and work across multiple server instances.',
    'Hard',
    'project',
    80,
    'https://github.com/devarena/rate-limiter-starter',
    '["Implement fixed window algorithm", "Implement sliding window algorithm", "Implement token bucket algorithm", "Support distributed rate limiting with Redis", "Add configurable limits per endpoint", "Handle edge cases gracefully"]',
    '["Node.js", "Redis", "System Design", "Algorithms"]',
    8,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-7',
    'SQL Complex Join',
    'Write complex SQL queries involving multiple JOINs, subqueries, window functions, and CTEs to analyze an e-commerce database.',
    'Medium',
    'project',
    35,
    'https://github.com/devarena/sql-complex-join-starter',
    '["Use multiple JOIN types", "Implement window functions", "Use CTEs effectively", "Optimize query performance", "Handle NULL values correctly"]',
    '["SQL", "PostgreSQL", "Query Optimization"]',
    3,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-8',
    'Sentiment Analysis',
    'Build a sentiment analysis tool that can classify text as positive, negative, or neutral. Use NLP techniques and optionally integrate with pre-trained models.',
    'Easy',
    'project',
    20,
    'https://github.com/devarena/sentiment-analysis-starter',
    '["Preprocess text data", "Implement basic tokenization", "Use sentiment lexicon or ML model", "Handle edge cases", "Provide confidence scores"]',
    '["Python", "NLP", "NLTK", "scikit-learn"]',
    2,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-9',
    'Image Classification',
    'Train an image classification model using TensorFlow/Keras. Build a CNN that can classify images into multiple categories with high accuracy.',
    'Hard',
    'project',
    100,
    'https://github.com/devarena/image-classification-starter',
    '["Build CNN architecture", "Implement data augmentation", "Use transfer learning", "Achieve > 90% accuracy", "Add model evaluation metrics", "Deploy as API endpoint"]',
    '["Python", "TensorFlow", "Keras", "Computer Vision", "Docker"]',
    10,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-10',
    'Debug Login Flow',
    'Find and fix bugs in a broken authentication flow. The login system has multiple issues including security vulnerabilities and logic errors.',
    'Medium',
    'bugfix',
    45,
    'https://github.com/devarena/debug-login-starter',
    '["Fix authentication logic bugs", "Patch security vulnerabilities", "Fix session management issues", "Add proper error handling", "Write tests for edge cases"]',
    '["JavaScript", "Node.js", "Express", "JWT", "Security"]',
    4,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-11',
    'Go Concurrency',
    'Master Go concurrency patterns by implementing a worker pool, rate limiter, and fan-out/fan-in pattern using goroutines and channels.',
    'Hard',
    'project',
    90,
    'https://github.com/devarena/go-concurrency-starter',
    '["Implement worker pool pattern", "Build rate limiter with goroutines", "Implement fan-out/fan-in", "Handle graceful shutdown", "Avoid race conditions", "Write comprehensive tests"]',
    '["Go", "Goroutines", "Channels", "Concurrency"]',
    8,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-12',
    'Task Manager',
    'Build a task manager application with Redux for state management. Implement CRUD operations, filtering, sorting, and drag-and-drop reordering.',
    'Medium',
    'project',
    30,
    'https://github.com/devarena/task-manager-starter',
    '["Use Redux Toolkit", "Implement CRUD operations", "Add filtering and sorting", "Implement drag-and-drop", "Persist state to localStorage", "Add undo/redo functionality"]',
    '["React", "Redux", "Redux Toolkit", "DnD"]',
    4,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-13',
    'Three.js Cube',
    'Create an interactive 3D scene with Three.js. Build a rotating cube with textures, lighting, and user interaction capabilities.',
    'Hard',
    'project',
    50,
    'https://github.com/devarena/threejs-cube-starter',
    '["Set up Three.js scene", "Add PBR materials and textures", "Implement lighting system", "Add orbit controls", "Create animations", "Optimize for performance"]',
    '["Three.js", "WebGL", "JavaScript", "3D Graphics"]',
    6,
    TRUE,
    NOW(),
    NOW()
),
(
    'challenge-14',
    'Next.js Blog',
    'Build a full-featured blog with Next.js using both SSG and SSR. Implement MDX support, dynamic routes, SEO optimization, and a CMS integration.',
    'Medium',
    'project',
    40,
    'https://github.com/devarena/nextjs-blog-starter',
    '["Use both SSG and SSR appropriately", "Implement MDX for content", "Add dynamic routing", "Optimize images with next/image", "Implement SEO best practices", "Add CMS integration"]',
    '["Next.js", "React", "MDX", "Tailwind CSS", "SEO"]',
    5,
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    difficulty = EXCLUDED.difficulty,
    type = EXCLUDED.type,
    max_score = EXCLUDED.max_score,
    requirements = EXCLUDED.requirements,
    tech_stack = EXCLUDED.tech_stack,
    estimated_hours = EXCLUDED.estimated_hours,
    is_published = EXCLUDED.is_published,
    updated_at = NOW();

-- Link challenges to tags
INSERT INTO challenge_tags (challenge_id, tag_id) VALUES
    -- CSS Flexbox Froggy
    ('challenge-1', 'tag-css'),
    ('challenge-1', 'tag-layout'),
    -- React Counter
    ('challenge-2', 'tag-react'),
    ('challenge-2', 'tag-hooks'),
    -- Infinite Scroll
    ('challenge-3', 'tag-react'),
    ('challenge-3', 'tag-dom'),
    -- Custom Dropdown
    ('challenge-4', 'tag-html'),
    ('challenge-4', 'tag-accessibility'),
    -- Node.js File Upload
    ('challenge-5', 'tag-nodejs'),
    ('challenge-5', 'tag-express'),
    -- API Rate Limiter
    ('challenge-6', 'tag-system-design'),
    ('challenge-6', 'tag-algorithms'),
    -- SQL Complex Join
    ('challenge-7', 'tag-sql'),
    ('challenge-7', 'tag-databases'),
    -- Sentiment Analysis
    ('challenge-8', 'tag-nlp'),
    ('challenge-8', 'tag-python'),
    -- Image Classification
    ('challenge-9', 'tag-tensorflow'),
    ('challenge-9', 'tag-cv'),
    -- Debug Login Flow
    ('challenge-10', 'tag-debug'),
    ('challenge-10', 'tag-auth'),
    -- Go Concurrency
    ('challenge-11', 'tag-golang'),
    ('challenge-11', 'tag-goroutines'),
    -- Task Manager
    ('challenge-12', 'tag-redux'),
    ('challenge-12', 'tag-state'),
    -- Three.js Cube
    ('challenge-13', 'tag-3d'),
    ('challenge-13', 'tag-canvas'),
    -- Next.js Blog
    ('challenge-14', 'tag-ssg'),
    ('challenge-14', 'tag-ssr')
ON CONFLICT (challenge_id, tag_id) DO NOTHING;

-- Verify insertion
SELECT 
    c.id, 
    c.title, 
    c.difficulty, 
    c.type, 
    c.max_score, 
    c.estimated_hours,
    c.is_published,
    array_agg(t.name) as tags
FROM challenges c
LEFT JOIN challenge_tags ct ON c.id = ct.challenge_id
LEFT JOIN tags t ON ct.tag_id = t.id
GROUP BY c.id
ORDER BY c.id;
