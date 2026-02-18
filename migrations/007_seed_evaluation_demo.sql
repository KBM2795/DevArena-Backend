-- Migration 007: Seed Dummy Data for Evaluation Pipeline
-- Inserts sample data to test the full 3-step evaluation flow
-- 
-- Prerequisites: 
--   - Users table has at least one user
--   - Challenges table has challenge-1 through challenge-14
--   - Migrations 005 and 006 have been applied

-- ============================================================================
-- 1. SEED PROMPT VERSIONS
-- ============================================================================

INSERT INTO prompt_versions (id, name, step, model, prompt_text, is_active) VALUES
(
    'prompt-code-review-v1',
    'code_review_v1',
    'code_review',
    'gpt-4o',
    'You are an expert code reviewer for DevArena. Analyze the submitted code and return a structured JSON evaluation.

You will receive:
- challenge context (title, description, requirements, constraints, rubric)
- submission files (source code)
- test summary (tests passed/failed)

Return ONLY valid JSON in this exact format:
{
  "scores": {
    "code_quality": <int>,
    "constraints": <int>,
    "architecture": <int>
  },
  "strengths": ["..."],
  "issues": ["..."],
  "improvements": ["..."]
}

RULES:
- Scores must be integers
- Scores must NOT exceed the rubric max for each category
- No negative scores
- All fields are required
- Do NOT calculate final score
- Do NOT modify functionality score',
    TRUE
),
(
    'prompt-report-v1',
    'report_v1',
    'report',
    'gpt-4o',
    'You are a professional coding mentor for DevArena. Generate a detailed, encouraging feedback report.

You will receive:
- challenge info (title, difficulty)
- score breakdown (final score, category scores)
- strengths, issues, and improvements from the code review

Return ONLY valid JSON in this exact format:
{
  "summary": "...",
  "detailed_feedback": {
    "functionality": "...",
    "code_quality": "...",
    "constraints": "...",
    "architecture": "..."
  },
  "dos": ["..."],
  "donts": ["..."],
  "next_steps": ["..."]
}

RULES:
- Be professional and encouraging
- Do NOT rescore or reinterpret raw code
- Reference specific strengths and issues from the review
- Keep suggestions actionable',
    TRUE
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. SEED DUMMY SUBMISSIONS
-- NOTE: Replace 'user-XXXXXX' with an actual user ID from your users table.
--       Run: SELECT id, email FROM users LIMIT 5;  to find valid IDs.
-- ============================================================================

-- You can update this user ID after checking your users table
DO $$
DECLARE
    v_user_id VARCHAR(255);
BEGIN
    -- Grab the first user from the table
    SELECT id INTO v_user_id FROM users LIMIT 1;

    -- If no user exists, create a dummy one
    IF v_user_id IS NULL THEN
        v_user_id := 'user-dummy-001';
        INSERT INTO users (id, clerk_user_id, email, username, display_name, onboarding_completed)
        VALUES (v_user_id, 'clerk_dummy_001', 'dummy@devarena.dev', 'dummydev', 'Dummy Developer', TRUE)
        ON CONFLICT (id) DO NOTHING;
    END IF;

    -- Insert dummy submissions
    INSERT INTO submissions (id, user_id, challenge_id, repo_url, branch, commit_hash, evaluation_status, evaluation_started_at, evaluation_completed_at)
    VALUES
    (
        'sub-demo-001',
        v_user_id,
        'challenge-1',
        'https://github.com/dummydev/challenge-1-solution',
        'main',
        'a1b2c3d4e5f6789012345678901234567890abcd',
        'completed',
        NOW() - INTERVAL '5 minutes',
        NOW() - INTERVAL '2 minutes'
    ),
    (
        'sub-demo-002',
        v_user_id,
        'challenge-2',
        'https://github.com/dummydev/challenge-2-solution',
        'main',
        'b2c3d4e5f6789012345678901234567890abcdef',
        'completed',
        NOW() - INTERVAL '10 minutes',
        NOW() - INTERVAL '7 minutes'
    ),
    (
        'sub-demo-003',
        v_user_id,
        'challenge-3',
        'https://github.com/dummydev/challenge-3-solution',
        'main',
        'c3d4e5f6789012345678901234567890abcdef01',
        'failed',
        NOW() - INTERVAL '15 minutes',
        NOW() - INTERVAL '14 minutes'
    )
    ON CONFLICT (id) DO NOTHING;

END $$;

-- ============================================================================
-- 3. SEED SUBMISSION TEST RESULTS (Step 1 — Docker Output)
-- ============================================================================

INSERT INTO submission_test_results (id, submission_id, build_success, tests_total, tests_passed, tests_failed, functionality_score, max_functionality_score, execution_time_ms, memory_usage_mb, raw_output) VALUES
(
    'test-result-001',
    'sub-demo-001',
    TRUE,
    10,
    9,
    1,
    45,          -- (9/10) * 50 = 45
    50,          -- from rubric: functionality = 50
    3420,
    128,
    '✓ Navbar should use Flexbox
✓ Nav Links should be horizontal
✓ Hero should be centered
✓ Footer should use Flexbox
✓ Cards should wrap on mobile
✓ Sidebar should stick on desktop
✓ Grid fallback works without CSS Grid
✓ No position:absolute used
✓ Responsive at 768px breakpoint
✗ Responsive at 480px breakpoint — expected padding 8px, got 16px

Tests: 9/10 passed
Time: 3.42s'
),
(
    'test-result-002',
    'sub-demo-002',
    TRUE,
    8,
    8,
    0,
    50,          -- (8/8) * 50 = 50 (perfect!)
    50,
    2150,
    96,
    '✓ Counter renders initial value 0
✓ Increment button increases count
✓ Decrement button decreases count
✓ Reset button returns to 0
✓ Keyboard shortcut ↑ increments
✓ Keyboard shortcut ↓ decrements
✓ Keyboard shortcut R resets
✓ Counter does not go below 0 when min bound set

Tests: 8/8 passed
Time: 2.15s'
),
(
    'test-result-003',
    'sub-demo-003',
    FALSE,
    0,
    0,
    0,
    0,
    50,
    1200,
    64,
    'ERROR: Build failed
> tsc && vite build

src/hooks/useInfiniteScroll.ts:15:3 - error TS2322: Type ''string'' is not assignable to type ''number''.
src/components/ItemList.tsx:8:5 - error TS2307: Cannot find module ''./LoadingSkeleton''

Build exited with code 1'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 4. SEED AI REVIEWS (Step 2 — AI Code Review)
-- Only for successful builds (sub-demo-001 and sub-demo-002)
-- ============================================================================

INSERT INTO ai_reviews (id, submission_id, code_quality_score, constraint_score, architecture_score, strengths_json, issues_json, improvements_json, prompt_version_id, raw_response, reviewed_at) VALUES
(
    'review-001',
    'sub-demo-001',
    20,          -- out of 25
    12,          -- out of 15
    8,           -- out of 10
    '["Clear separation of navbar, hero, and footer sections", "Proper use of flexbox properties throughout", "Clean and readable CSS with logical grouping", "Good use of semantic HTML elements"]',
    '["Media queries could be more structured — mobile-first approach recommended", "Repeated CSS rules for flex centering across multiple components", "Missing hover states on navigation links", "480px breakpoint padding is incorrect"]',
    '["Use CSS custom properties (variables) to reduce duplication", "Adopt mobile-first media query strategy", "Add transition effects for interactive elements", "Consider using a CSS reset for cross-browser consistency"]',
    'prompt-code-review-v1',
    '{"scores":{"code_quality":20,"constraints":12,"architecture":8},"strengths":["Clear separation of navbar, hero, and footer sections","Proper use of flexbox properties throughout","Clean and readable CSS with logical grouping","Good use of semantic HTML elements"],"issues":["Media queries could be more structured — mobile-first approach recommended","Repeated CSS rules for flex centering across multiple components","Missing hover states on navigation links","480px breakpoint padding is incorrect"],"improvements":["Use CSS custom properties (variables) to reduce duplication","Adopt mobile-first media query strategy","Add transition effects for interactive elements","Consider using a CSS reset for cross-browser consistency"]}',
    NOW() - INTERVAL '3 minutes'
),
(
    'review-002',
    'sub-demo-002',
    22,          -- out of 25
    14,          -- out of 15
    9,           -- out of 10
    '["Excellent use of useState hook with clean state logic", "All keyboard shortcuts implemented correctly", "Component is well-structured and focused", "CSS modules used properly with no style leakage", "Edge case handled: counter minimum bound"]',
    '["Could extract keyboard handler into a custom hook", "No TypeScript types defined for component props"]',
    '["Create a useKeyboardShortcuts custom hook for reusability", "Add PropTypes or TypeScript interfaces", "Consider adding animation on value change for better UX", "Add aria-label to buttons for accessibility"]',
    'prompt-code-review-v1',
    '{"scores":{"code_quality":22,"constraints":14,"architecture":9},"strengths":["Excellent use of useState hook with clean state logic","All keyboard shortcuts implemented correctly","Component is well-structured and focused","CSS modules used properly with no style leakage","Edge case handled: counter minimum bound"],"issues":["Could extract keyboard handler into a custom hook","No TypeScript types defined for component props"],"improvements":["Create a useKeyboardShortcuts custom hook for reusability","Add PropTypes or TypeScript interfaces","Consider adding animation on value change for better UX","Add aria-label to buttons for accessibility"]}',
    NOW() - INTERVAL '8 minutes'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 5. SEED SUBMISSION SCORES (Computed in Go — final math)
-- final_score = functionality + code_quality + constraints + architecture
-- ============================================================================

INSERT INTO submission_scores (id, submission_id, functionality_score, code_quality_score, constraint_score, architecture_score, final_score, max_score) VALUES
(
    'score-001',
    'sub-demo-001',
    45,          -- from test results
    20,          -- from AI review
    12,          -- from AI review
    8,           -- from AI review
    85,          -- 45 + 20 + 12 + 8 = 85
    100
),
(
    'score-002',
    'sub-demo-002',
    50,          -- from test results (perfect)
    22,          -- from AI review
    14,          -- from AI review
    9,           -- from AI review
    95,          -- 50 + 22 + 14 + 9 = 95
    100
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 6. SEED AI REPORTS (Step 3 — Professional Feedback)
-- ============================================================================

INSERT INTO ai_reports (id, submission_id, summary_md, detailed_feedback_md, dos_json, donts_json, next_steps_json, prompt_version_id, raw_response) VALUES
(
    'report-001',
    'sub-demo-001',
    '## Great Work on the Responsive Navbar! 🎉

Your solution demonstrates a **solid understanding** of CSS Flexbox fundamentals. You scored **85/100**, with particularly strong marks in architecture and code organization. The navbar, hero section, and footer are all well-structured using proper flexbox properties.

One test failed due to incorrect padding at the 480px breakpoint — a small fix that would push your functionality score even higher.',

    '{"functionality": "You passed 9 out of 10 layout tests. The only failure was at the 480px mobile breakpoint where padding was 16px instead of the expected 8px. This is likely a specificity issue in your media queries.", "code_quality": "Your CSS is clean and well-organized with logical grouping of related properties. However, there is noticeable duplication in flex centering rules that could be extracted into utility classes.", "constraints": "You correctly avoided position:absolute throughout and used only flexbox for layout. The grid fallback also works properly. Minor deduction for unstructured media queries.", "architecture": "Good separation of concerns between navbar, hero, and footer. Semantic HTML usage is appropriate. The component structure supports future extensibility."}',

    '["Continue using semantic HTML elements for structure", "Keep CSS organized with logical property grouping", "Maintain the no-absolute-positioning discipline", "Use flexbox gap property — you applied it well"]',
    '["Avoid repeating flex centering rules — extract to a utility class", "Don''t use fixed pixel values for spacing at small breakpoints", "Avoid mixing mobile-first and desktop-first media queries"]',
    '["Fix the 480px breakpoint padding to pass the remaining test", "Refactor repeated CSS into utility classes", "Practice mobile-first responsive design approach", "Try the Custom Dropdown challenge next to practice accessible components"]',
    'prompt-report-v1',
    NULL
),
(
    'report-002',
    'sub-demo-002',
    '## Excellent React Counter Implementation! 🏆

Your solution scored an impressive **95/100** — one of the highest scores on this challenge. Every test passed, all keyboard shortcuts work flawlessly, and your component structure is clean and focused.

The small deductions were for missing TypeScript types and a keyboard handler that could be extracted into a reusable hook.',

    '{"functionality": "Perfect score! All 8 tests passed including increment, decrement, reset, keyboard shortcuts, and the minimum bound edge case. Your implementation handles every requirement correctly.", "code_quality": "Very clean code with good separation of concerns. State logic is simple and focused. The only improvements would be adding TypeScript interfaces and extracting the keyboard handler.", "constraints": "Excellent constraint adherence. You used useState (not class state), styled with CSS modules, and implemented all required keyboard shortcuts. Only minor deduction for not defining explicit TypeScript types.", "architecture": "Well-structured single-component design that is appropriate for this challenge scope. The CSS module approach prevents style leakage effectively."}',

    '["Continue using useState for simple state — you used it perfectly", "Keep components focused and single-responsibility", "Maintain CSS module usage for style isolation", "Good edge case handling with minimum bounds"]',
    '["Avoid defining event handlers inline — extract them", "Don''t skip TypeScript types even for simple components"]',
    '["Extract keyboard shortcuts into a useKeyboardShortcuts custom hook", "Add subtle animation on counter value change", "Add aria-labels to buttons for screen reader accessibility", "Try the Infinite Scroll challenge next to practice advanced hooks"]',
    'prompt-report-v1',
    NULL
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 7. VERIFICATION — Check all seeded data
-- ============================================================================

-- Prompt versions
SELECT id, name, step, model, is_active FROM prompt_versions;

-- Submissions with evaluation status
SELECT id, challenge_id, evaluation_status, evaluation_started_at, evaluation_completed_at FROM submissions WHERE id LIKE 'sub-demo%';

-- Test results
SELECT id, submission_id, build_success, tests_passed, tests_total, functionality_score FROM submission_test_results;

-- AI reviews
SELECT id, submission_id, code_quality_score, constraint_score, architecture_score FROM ai_reviews WHERE id LIKE 'review%';

-- Final scores
SELECT id, submission_id, functionality_score, code_quality_score, constraint_score, architecture_score, final_score FROM submission_scores;

-- Reports
SELECT id, submission_id, LEFT(summary_md, 60) AS summary_preview FROM ai_reports;
