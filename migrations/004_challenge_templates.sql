-- Migration: Add challenge_templates table
-- Stores template configuration for file locking, validation rules, and UI tree

CREATE TABLE IF NOT EXISTS challenge_templates (
    id VARCHAR(255) PRIMARY KEY,
    challenge_id VARCHAR(255) NOT NULL UNIQUE REFERENCES challenges(id) ON DELETE CASCADE,
    repo_template_url TEXT NOT NULL,
    entry_file VARCHAR(255) DEFAULT 'src/index.tsx',
    allowed_edit_paths JSONB DEFAULT '[]',
    readonly_paths JSONB DEFAULT '[]',
    forbidden_packages JSONB DEFAULT '[]',
    template_tree JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_challenge_templates_challenge_id ON challenge_templates(challenge_id);

-- Seed template data for existing challenges
INSERT INTO challenge_templates (id, challenge_id, repo_template_url, entry_file, allowed_edit_paths, readonly_paths, forbidden_packages, template_tree) VALUES
(
    'template-1',
    'challenge-1',
    'https://github.com/devarena/css-flexbox-starter',
    'src/styles.css',
    '["src/styles.css", "src/components"]',
    '["package.json", "tsconfig.json", "src/index.tsx"]',
    '[]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "components", "type": "folder", "editable": true, "children": [
                {"name": "FlexContainer.tsx", "type": "file", "editable": true}
            ]},
            {"name": "index.tsx", "type": "file", "editable": false},
            {"name": "styles.css", "type": "file", "editable": true}
        ]},
        {"name": "package.json", "type": "file", "editable": false},
        {"name": "README.md", "type": "file", "editable": false}
    ]'
),
(
    'template-2',
    'challenge-2',
    'https://github.com/devarena/react-counter-starter',
    'src/components/Counter.tsx',
    '["src/components", "src/styles"]',
    '["package.json", "tsconfig.json", "src/index.tsx", "src/App.tsx"]',
    '["lodash", "underscore"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "components", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "Counter.tsx", "type": "file", "editable": true},
                {"name": "Counter.module.css", "type": "file", "editable": true}
            ]},
            {"name": "App.tsx", "type": "file", "editable": false},
            {"name": "index.tsx", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false},
        {"name": "README.md", "type": "file", "editable": false}
    ]'
),
(
    'template-3',
    'challenge-3',
    'https://github.com/devarena/infinite-scroll-starter',
    'src/hooks/useInfiniteScroll.ts',
    '["src/hooks", "src/components"]',
    '["package.json", "src/index.tsx", "src/api"]',
    '["react-infinite-scroll-component", "react-virtualized"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "hooks", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "useInfiniteScroll.ts", "type": "file", "editable": true}
            ]},
            {"name": "components", "type": "folder", "editable": true, "children": [
                {"name": "ItemList.tsx", "type": "file", "editable": true},
                {"name": "LoadingSkeleton.tsx", "type": "file", "editable": true}
            ]},
            {"name": "api", "type": "folder", "editable": false, "children": [
                {"name": "fetchItems.ts", "type": "file", "editable": false}
            ]},
            {"name": "index.tsx", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-4',
    'challenge-4',
    'https://github.com/devarena/custom-dropdown-starter',
    'src/components/Dropdown.tsx',
    '["src/components", "src/styles"]',
    '["package.json", "src/index.tsx"]',
    '["react-select", "@radix-ui/react-select", "downshift"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "components", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "Dropdown.tsx", "type": "file", "editable": true},
                {"name": "DropdownOption.tsx", "type": "file", "editable": true}
            ]},
            {"name": "styles", "type": "folder", "editable": true, "children": [
                {"name": "dropdown.css", "type": "file", "editable": true}
            ]},
            {"name": "index.tsx", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-5',
    'challenge-5',
    'https://github.com/devarena/nodejs-file-upload-starter',
    'src/routes/upload.ts',
    '["src/routes", "src/middleware", "src/utils"]',
    '["package.json", "src/index.ts", "src/config"]',
    '[]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "routes", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "upload.ts", "type": "file", "editable": true}
            ]},
            {"name": "middleware", "type": "folder", "editable": true, "children": [
                {"name": "fileValidator.ts", "type": "file", "editable": true}
            ]},
            {"name": "utils", "type": "folder", "editable": true, "children": [
                {"name": "storage.ts", "type": "file", "editable": true}
            ]},
            {"name": "config", "type": "folder", "editable": false, "children": [
                {"name": "index.ts", "type": "file", "editable": false}
            ]},
            {"name": "index.ts", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-6',
    'challenge-6',
    'https://github.com/devarena/rate-limiter-starter',
    'src/limiters/index.ts',
    '["src/limiters", "src/middleware"]',
    '["package.json", "src/index.ts", "src/redis"]',
    '[]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "limiters", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "fixedWindow.ts", "type": "file", "editable": true},
                {"name": "slidingWindow.ts", "type": "file", "editable": true},
                {"name": "tokenBucket.ts", "type": "file", "editable": true},
                {"name": "index.ts", "type": "file", "editable": true}
            ]},
            {"name": "middleware", "type": "folder", "editable": true, "children": [
                {"name": "rateLimiter.ts", "type": "file", "editable": true}
            ]},
            {"name": "redis", "type": "folder", "editable": false, "children": [
                {"name": "client.ts", "type": "file", "editable": false}
            ]},
            {"name": "index.ts", "type": "file", "editable": false}
        ]},
        {"name": "docker-compose.yml", "type": "file", "editable": false},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-7',
    'challenge-7',
    'https://github.com/devarena/sql-complex-join-starter',
    'queries/solutions.sql',
    '["queries"]',
    '["schema", "seeds", "docker-compose.yml"]',
    '[]',
    '[
        {"name": "queries", "type": "folder", "isOpen": true, "editable": true, "children": [
            {"name": "solutions.sql", "type": "file", "editable": true},
            {"name": "tests.sql", "type": "file", "editable": false}
        ]},
        {"name": "schema", "type": "folder", "editable": false, "children": [
            {"name": "tables.sql", "type": "file", "editable": false}
        ]},
        {"name": "seeds", "type": "folder", "editable": false, "children": [
            {"name": "data.sql", "type": "file", "editable": false}
        ]},
        {"name": "docker-compose.yml", "type": "file", "editable": false}
    ]'
),
(
    'template-8',
    'challenge-8',
    'https://github.com/devarena/sentiment-analysis-starter',
    'src/analyzer.py',
    '["src"]',
    '["requirements.txt", "tests", "data"]',
    '["transformers", "torch"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": true, "children": [
            {"name": "analyzer.py", "type": "file", "editable": true},
            {"name": "preprocessor.py", "type": "file", "editable": true},
            {"name": "lexicon.py", "type": "file", "editable": true}
        ]},
        {"name": "tests", "type": "folder", "editable": false, "children": [
            {"name": "test_analyzer.py", "type": "file", "editable": false}
        ]},
        {"name": "data", "type": "folder", "editable": false, "children": [
            {"name": "sentiment_lexicon.json", "type": "file", "editable": false}
        ]},
        {"name": "requirements.txt", "type": "file", "editable": false},
        {"name": "app.py", "type": "file", "editable": false}
    ]'
),
(
    'template-9',
    'challenge-9',
    'https://github.com/devarena/image-classification-starter',
    'src/model.py',
    '["src"]',
    '["requirements.txt", "data", "tests"]',
    '[]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": true, "children": [
            {"name": "model.py", "type": "file", "editable": true},
            {"name": "train.py", "type": "file", "editable": true},
            {"name": "data_loader.py", "type": "file", "editable": true},
            {"name": "augmentation.py", "type": "file", "editable": true}
        ]},
        {"name": "data", "type": "folder", "editable": false, "children": [
            {"name": "train", "type": "folder", "editable": false, "children": []},
            {"name": "test", "type": "folder", "editable": false, "children": []}
        ]},
        {"name": "requirements.txt", "type": "file", "editable": false},
        {"name": "Dockerfile", "type": "file", "editable": false}
    ]'
),
(
    'template-10',
    'challenge-10',
    'https://github.com/devarena/debug-login-starter',
    'src/auth/login.ts',
    '["src/auth", "src/middleware", "src/utils"]',
    '["package.json", "src/index.ts", "tests"]',
    '[]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "auth", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "login.ts", "type": "file", "editable": true},
                {"name": "session.ts", "type": "file", "editable": true},
                {"name": "password.ts", "type": "file", "editable": true}
            ]},
            {"name": "middleware", "type": "folder", "editable": true, "children": [
                {"name": "auth.ts", "type": "file", "editable": true}
            ]},
            {"name": "utils", "type": "folder", "editable": true, "children": [
                {"name": "crypto.ts", "type": "file", "editable": true}
            ]},
            {"name": "index.ts", "type": "file", "editable": false}
        ]},
        {"name": "tests", "type": "folder", "editable": false, "children": [
            {"name": "auth.test.ts", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-11',
    'challenge-11',
    'https://github.com/devarena/go-concurrency-starter',
    'pkg/patterns/worker_pool.go',
    '["pkg/patterns"]',
    '["go.mod", "go.sum", "cmd", "internal/tests"]',
    '[]',
    '[
        {"name": "pkg", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "patterns", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "worker_pool.go", "type": "file", "editable": true},
                {"name": "rate_limiter.go", "type": "file", "editable": true},
                {"name": "fan_out_in.go", "type": "file", "editable": true}
            ]}
        ]},
        {"name": "cmd", "type": "folder", "editable": false, "children": [
            {"name": "main.go", "type": "file", "editable": false}
        ]},
        {"name": "internal", "type": "folder", "editable": false, "children": [
            {"name": "tests", "type": "folder", "editable": false, "children": [
                {"name": "patterns_test.go", "type": "file", "editable": false}
            ]}
        ]},
        {"name": "go.mod", "type": "file", "editable": false}
    ]'
),
(
    'template-12',
    'challenge-12',
    'https://github.com/devarena/task-manager-starter',
    'src/store/tasksSlice.ts',
    '["src/store", "src/components", "src/hooks"]',
    '["package.json", "src/index.tsx", "src/App.tsx"]',
    '["redux", "react-redux"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "store", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "tasksSlice.ts", "type": "file", "editable": true},
                {"name": "store.ts", "type": "file", "editable": true}
            ]},
            {"name": "components", "type": "folder", "editable": true, "children": [
                {"name": "TaskList.tsx", "type": "file", "editable": true},
                {"name": "TaskItem.tsx", "type": "file", "editable": true},
                {"name": "TaskForm.tsx", "type": "file", "editable": true}
            ]},
            {"name": "hooks", "type": "folder", "editable": true, "children": [
                {"name": "useLocalStorage.ts", "type": "file", "editable": true}
            ]},
            {"name": "App.tsx", "type": "file", "editable": false},
            {"name": "index.tsx", "type": "file", "editable": false}
        ]},
        {"name": "package.json", "type": "file", "editable": false}
    ]'
),
(
    'template-13',
    'challenge-13',
    'https://github.com/devarena/threejs-cube-starter',
    'src/scene.ts',
    '["src"]',
    '["package.json", "public", "vite.config.ts"]',
    '["@react-three/fiber", "react-three-fiber"]',
    '[
        {"name": "src", "type": "folder", "isOpen": true, "editable": true, "children": [
            {"name": "scene.ts", "type": "file", "editable": true},
            {"name": "cube.ts", "type": "file", "editable": true},
            {"name": "lights.ts", "type": "file", "editable": true},
            {"name": "controls.ts", "type": "file", "editable": true},
            {"name": "main.ts", "type": "file", "editable": true}
        ]},
        {"name": "public", "type": "folder", "editable": false, "children": [
            {"name": "textures", "type": "folder", "editable": false, "children": []}
        ]},
        {"name": "package.json", "type": "file", "editable": false},
        {"name": "vite.config.ts", "type": "file", "editable": false}
    ]'
),
(
    'template-14',
    'challenge-14',
    'https://github.com/devarena/nextjs-blog-starter',
    'app/blog/[slug]/page.tsx',
    '["app/blog", "components", "lib"]',
    '["package.json", "next.config.js", "app/layout.tsx"]',
    '[]',
    '[
        {"name": "app", "type": "folder", "isOpen": true, "editable": false, "children": [
            {"name": "blog", "type": "folder", "editable": true, "isOpen": true, "children": [
                {"name": "[slug]", "type": "folder", "editable": true, "children": [
                    {"name": "page.tsx", "type": "file", "editable": true}
                ]},
                {"name": "page.tsx", "type": "file", "editable": true}
            ]},
            {"name": "layout.tsx", "type": "file", "editable": false},
            {"name": "page.tsx", "type": "file", "editable": false}
        ]},
        {"name": "components", "type": "folder", "editable": true, "children": [
            {"name": "BlogPost.tsx", "type": "file", "editable": true},
            {"name": "MDXComponents.tsx", "type": "file", "editable": true}
        ]},
        {"name": "lib", "type": "folder", "editable": true, "children": [
            {"name": "mdx.ts", "type": "file", "editable": true}
        ]},
        {"name": "package.json", "type": "file", "editable": false},
        {"name": "next.config.js", "type": "file", "editable": false}
    ]'
)
ON CONFLICT (id) DO UPDATE SET
    repo_template_url = EXCLUDED.repo_template_url,
    entry_file = EXCLUDED.entry_file,
    allowed_edit_paths = EXCLUDED.allowed_edit_paths,
    readonly_paths = EXCLUDED.readonly_paths,
    forbidden_packages = EXCLUDED.forbidden_packages,
    template_tree = EXCLUDED.template_tree,
    updated_at = NOW();
