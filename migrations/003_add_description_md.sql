-- Migration: Add description_md column to challenges table
-- This column stores rich Markdown content for each challenge's detailed description

-- Add the description_md column
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS description_md TEXT;

-- Backfill existing challenges with sample Markdown content
UPDATE challenges SET description_md = 
'## Overview
Master CSS Flexbox by completing a series of layout challenges. You will learn how to use flex-direction, justify-content, align-items, flex-wrap, and more to position elements on a page.

## What You''ll Learn
- Use `flex-direction` to control layout direction
- Master `justify-content` and `align-items` for alignment
- Apply `flex-wrap` for responsive layouts
- Understand `flex-grow`, `flex-shrink`, and `flex-basis`

## Requirements
1. Use **only CSS Flexbox** properties for layouts
2. No `position: absolute` or `position: fixed` allowed
3. All layouts must be responsive (mobile to desktop)
4. Pass all automated layout tests

## Getting Started
```bash
git clone https://github.com/devarena/css-flexbox-starter
cd css-flexbox-starter
npm install
npm run dev
```

> ⚠️ **Important**: You must build upon the provided template. Submissions that do not follow the folder structure will be automatically rejected by the AI grader.'
WHERE id = 'challenge-1';

UPDATE challenges SET description_md = 
'## Overview
Build a simple counter application using React hooks. Implement increment, decrement, and reset functionality while maintaining clean component structure.

## What You''ll Learn
- Master the `useState` hook for state management
- Handle user events in React
- Implement keyboard shortcuts for accessibility
- Style components with CSS modules

## Requirements
1. Use the `useState` hook for counter state
2. Implement increment, decrement, and reset buttons
3. Add keyboard shortcuts (↑ for increment, ↓ for decrement, R for reset)
4. Style the counter with CSS modules

## Getting Started
```bash
git clone https://github.com/devarena/react-counter-starter
cd react-counter-starter
npm install
npm run dev
```

## Bonus Challenges
- Add step size configuration
- Implement min/max bounds
- Add animation on value change'
WHERE id = 'challenge-2';

UPDATE challenges SET description_md = 
'## Overview
Implement infinite scroll functionality in a React application. Load more data as the user scrolls to the bottom of the page, with proper loading states and error handling.

## What You''ll Learn
- Use the Intersection Observer API for scroll detection
- Manage loading, error, and empty states
- Implement skeleton loading for better UX
- Optimize performance with virtualization techniques

## Requirements
1. Use **Intersection Observer API** (no scroll event listeners)
2. Show skeleton loading while fetching data
3. Handle network errors gracefully with retry option
4. Display appropriate message for empty results

## API Endpoint
```javascript
// Fetch paginated data
const response = await fetch(`/api/items?page=${page}&limit=20`);
const { data, hasMore } = await response.json();
```

## Getting Started
```bash
git clone https://github.com/devarena/infinite-scroll-starter
cd infinite-scroll-starter
npm install
npm run dev
```

> 💡 **Tip**: Consider using `react-virtual` or `@tanstack/react-virtual` for better performance with large lists.'
WHERE id = 'challenge-3';

UPDATE challenges SET description_md = 
'## Overview
Create an accessible custom dropdown component. It should support keyboard navigation, screen readers, and follow WAI-ARIA guidelines.

## What You''ll Learn
- Implement ARIA attributes for accessibility
- Handle keyboard navigation (Arrow keys, Enter, Escape)
- Manage focus correctly for dropdown interactions
- Build mobile-friendly touch interactions

## Requirements
1. Full keyboard navigation support
2. Proper ARIA attributes (`role`, `aria-expanded`, `aria-activedescendant`)
3. Focus management (trap focus when open, restore on close)
4. Optional multi-select mode
5. Mobile-friendly with touch support

## Accessibility Checklist
```
✅ Dropdown toggle has role="combobox"
✅ Options list has role="listbox"
✅ Each option has role="option"
✅ Arrow keys navigate options
✅ Enter/Space selects option
✅ Escape closes dropdown
✅ Screen reader announces selection
```

## Getting Started
```bash
git clone https://github.com/devarena/custom-dropdown-starter
cd custom-dropdown-starter
npm install
npm run dev
```'
WHERE id = 'challenge-4';

UPDATE challenges SET description_md = 
'## Overview
Build a file upload service with Node.js and Express. Support multiple file uploads, file validation, progress tracking, and storage to local filesystem or cloud.

## What You''ll Learn
- Handle multipart form data with Multer
- Validate file types and sizes securely
- Implement upload progress tracking
- Integrate with cloud storage (AWS S3)

## Requirements
1. Accept multipart form data uploads
2. Validate file types (images, PDFs only) and size limits (10MB max)
3. Provide upload progress via Server-Sent Events
4. Support both local storage and AWS S3
5. Return uploaded file URLs in response

## API Endpoints
```javascript
POST /api/upload          // Single file upload
POST /api/upload/multiple // Multiple files upload
GET  /api/upload/progress // SSE progress stream
```

## Getting Started
```bash
git clone https://github.com/devarena/nodejs-file-upload-starter
cd nodejs-file-upload-starter
npm install
cp .env.example .env
npm run dev
```

> ⚠️ **Security Note**: Always validate file types on the server. Never trust the `Content-Type` header alone—check the file signature (magic bytes).'
WHERE id = 'challenge-5';

UPDATE challenges SET description_md = 
'## Overview
Design and implement an API rate limiter that can handle multiple strategies: fixed window, sliding window, token bucket. Must be distributed and work across multiple server instances.

## What You''ll Learn
- Implement different rate limiting algorithms
- Use Redis for distributed rate limiting
- Configure limits per endpoint and user
- Handle edge cases gracefully

## Algorithms to Implement

### 1. Fixed Window
Count requests in fixed time windows (e.g., 100 requests per minute).

### 2. Sliding Window
More accurate than fixed window—prevents burst at window boundaries.

### 3. Token Bucket
Allows burst traffic up to bucket capacity, refills at steady rate.

## Requirements
1. Implement all three algorithms
2. Support distributed rate limiting with Redis
3. Configure different limits per endpoint
4. Return proper `429 Too Many Requests` with `Retry-After` header
5. Include rate limit headers in responses

## Response Headers
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640995200
```

## Getting Started
```bash
git clone https://github.com/devarena/rate-limiter-starter
cd rate-limiter-starter
docker-compose up -d redis
npm install
npm run dev
```'
WHERE id = 'challenge-6';

UPDATE challenges SET description_md = 
'## Overview
Write complex SQL queries involving multiple JOINs, subqueries, window functions, and CTEs to analyze an e-commerce database.

## What You''ll Learn
- Master different JOIN types (INNER, LEFT, RIGHT, FULL)
- Use window functions (ROW_NUMBER, RANK, LAG, LEAD)
- Write Common Table Expressions (CTEs)
- Optimize query performance

## Database Schema
```sql
users (id, name, email, created_at)
products (id, name, price, category_id)
categories (id, name, parent_id)
orders (id, user_id, total, status, created_at)
order_items (id, order_id, product_id, quantity, price)
```

## Requirements
1. Use multiple JOIN types correctly
2. Implement window functions for rankings and running totals
3. Use CTEs to simplify complex queries
4. Optimize queries with proper indexes
5. Handle NULL values correctly

## Sample Queries to Implement
- Top 10 customers by lifetime value
- Month-over-month revenue growth
- Product affinity analysis (frequently bought together)
- Category sales hierarchy report

## Getting Started
```bash
git clone https://github.com/devarena/sql-complex-join-starter
cd sql-complex-join-starter
docker-compose up -d postgres
npm run seed
npm run dev
```'
WHERE id = 'challenge-7';

UPDATE challenges SET description_md = 
'## Overview
Build a sentiment analysis tool that can classify text as positive, negative, or neutral. Use NLP techniques and optionally integrate with pre-trained models.

## What You''ll Learn
- Preprocess text data (tokenization, stemming, lemmatization)
- Use sentiment lexicons (VADER, TextBlob)
- Train custom ML models with scikit-learn
- Provide confidence scores for predictions

## Requirements
1. Clean and preprocess text input
2. Implement basic tokenization
3. Use either lexicon-based or ML-based approach
4. Handle edge cases (sarcasm, negations, emojis)
5. Return sentiment label with confidence score

## API Response Format
```json
{
  "text": "I love this product!",
  "sentiment": "positive",
  "confidence": 0.92,
  "scores": {
    "positive": 0.92,
    "neutral": 0.05,
    "negative": 0.03
  }
}
```

## Getting Started
```bash
git clone https://github.com/devarena/sentiment-analysis-starter
cd sentiment-analysis-starter
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python app.py
```

> 💡 **Tip**: For production, consider using pre-trained transformers like BERT or DistilBERT from Hugging Face.'
WHERE id = 'challenge-8';

UPDATE challenges SET description_md = 
'## Overview
Train an image classification model using TensorFlow/Keras. Build a CNN that can classify images into multiple categories with high accuracy.

## What You''ll Learn
- Build CNN architecture from scratch
- Apply data augmentation for better generalization
- Use transfer learning with pre-trained models
- Evaluate models with proper metrics
- Deploy as an API endpoint

## Requirements
1. Build a CNN with at least 3 convolutional layers
2. Implement data augmentation (rotation, flip, zoom)
3. Use transfer learning (ResNet50 or MobileNetV2)
4. Achieve **>90% accuracy** on test set
5. Add model evaluation metrics (precision, recall, F1)
6. Deploy as REST API using FastAPI or Flask

## Model Architecture
```python
model = Sequential([
    Conv2D(32, (3, 3), activation="relu", input_shape=(224, 224, 3)),
    MaxPooling2D((2, 2)),
    Conv2D(64, (3, 3), activation="relu"),
    MaxPooling2D((2, 2)),
    Conv2D(128, (3, 3), activation="relu"),
    MaxPooling2D((2, 2)),
    Flatten(),
    Dense(128, activation="relu"),
    Dropout(0.5),
    Dense(num_classes, activation="softmax")
])
```

## Getting Started
```bash
git clone https://github.com/devarena/image-classification-starter
cd image-classification-starter
pip install -r requirements.txt
python train.py
```

> ⚠️ **GPU Recommended**: Training CNNs is much faster with GPU. Use Google Colab if you don''t have local GPU.'
WHERE id = 'challenge-9';

UPDATE challenges SET description_md = 
'## Overview
Find and fix bugs in a broken authentication flow. The login system has multiple issues including security vulnerabilities and logic errors.

## What You''ll Learn
- Identify common authentication bugs
- Fix security vulnerabilities (XSS, CSRF, injection)
- Implement proper session management
- Write tests for edge cases

## Known Issues to Fix
The codebase contains **at least 8 bugs**:
- [ ] SQL injection vulnerability
- [ ] Missing password hashing
- [ ] Session fixation vulnerability
- [ ] Improper error messages (information leakage)
- [ ] Missing rate limiting on login
- [ ] JWT token never expires
- [ ] Password reset token reuse
- [ ] CSRF protection missing

## Requirements
1. Fix all authentication logic bugs
2. Patch security vulnerabilities
3. Implement proper session management
4. Add comprehensive error handling
5. Write unit tests for all edge cases

## Getting Started
```bash
git clone https://github.com/devarena/debug-login-starter
cd debug-login-starter
npm install
npm run dev
npm test  # Run tests to find bugs
```

> 🔒 **Security First**: Never expose stack traces or database errors to users. Use generic error messages for failed login attempts.'
WHERE id = 'challenge-10';

UPDATE challenges SET description_md = 
'## Overview
Master Go concurrency patterns by implementing a worker pool, rate limiter, and fan-out/fan-in pattern using goroutines and channels.

## What You''ll Learn
- Implement worker pool pattern for parallel processing
- Build a concurrent rate limiter
- Use fan-out/fan-in for parallel → sequential pipelines
- Handle graceful shutdown
- Avoid race conditions

## Patterns to Implement

### 1. Worker Pool
```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
}

func (wp *WorkerPool) Start(ctx context.Context)
func (wp *WorkerPool) Submit(job Job) error
func (wp *WorkerPool) Shutdown()
```

### 2. Rate Limiter
```go
type RateLimiter struct {
    rate     int           // requests per second
    burst    int           // max burst size
    tokens   chan struct{}
}

func (rl *RateLimiter) Allow() bool
func (rl *RateLimiter) Wait(ctx context.Context) error
```

### 3. Fan-Out/Fan-In
```go
func FanOut[T, R any](input <-chan T, workers int, process func(T) R) <-chan R
func FanIn[T any](channels ...<-chan T) <-chan T
```

## Requirements
1. All patterns must be context-aware for cancellation
2. Handle graceful shutdown with cleanup
3. Avoid race conditions (verified with `-race` flag)
4. Write comprehensive tests with `go test`

## Getting Started
```bash
git clone https://github.com/devarena/go-concurrency-starter
cd go-concurrency-starter
go mod tidy
go test -race ./...
```'
WHERE id = 'challenge-11';

UPDATE challenges SET description_md = 
'## Overview
Build a task manager application with Redux for state management. Implement CRUD operations, filtering, sorting, and drag-and-drop reordering.

## What You''ll Learn
- Use Redux Toolkit for state management
- Implement CRUD with optimistic updates
- Build drag-and-drop with `@dnd-kit`
- Add undo/redo functionality
- Persist state to localStorage

## Requirements
1. Use **Redux Toolkit** (not vanilla Redux)
2. Full CRUD operations (Create, Read, Update, Delete)
3. Filter by status, priority, and tags
4. Sort by due date, priority, or created date
5. Drag-and-drop task reordering
6. Persist to localStorage with rehydration
7. Undo/redo for all actions

## Redux Slice Structure
```javascript
const tasksSlice = createSlice({
  name: "tasks",
  initialState: {
    items: [],
    filter: "all",
    sortBy: "createdAt",
    history: [],
    future: []
  },
  reducers: {
    addTask, updateTask, deleteTask,
    reorderTasks, setFilter, setSortBy,
    undo, redo
  }
});
```

## Getting Started
```bash
git clone https://github.com/devarena/task-manager-starter
cd task-manager-starter
npm install
npm run dev
```'
WHERE id = 'challenge-12';

UPDATE challenges SET description_md = 
'## Overview
Create an interactive 3D scene with Three.js. Build a rotating cube with textures, lighting, and user interaction capabilities.

## What You''ll Learn
- Set up Three.js scene, camera, and renderer
- Apply PBR materials and textures
- Implement lighting (ambient, directional, point)
- Add orbit controls for camera movement
- Animate objects with requestAnimationFrame
- Optimize for performance

## Requirements
1. Set up complete Three.js scene
2. Create cube with PBR materials (albedo, normal, roughness maps)
3. Add 3-point lighting setup
4. Implement OrbitControls for camera
5. Animate cube rotation
6. Handle window resize responsively
7. Achieve 60 FPS on mid-range hardware

## Scene Setup
```javascript
const scene = new THREE.Scene();
const camera = new THREE.PerspectiveCamera(75, aspect, 0.1, 1000);
const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
```

## Getting Started
```bash
git clone https://github.com/devarena/threejs-cube-starter
cd threejs-cube-starter
npm install
npm run dev
```

> 💡 **Performance Tip**: Limit `devicePixelRatio` to 2 for better performance on high-DPI displays.'
WHERE id = 'challenge-13';

UPDATE challenges SET description_md = 
'## Overview
Build a full-featured blog with Next.js using both SSG and SSR. Implement MDX support, dynamic routes, SEO optimization, and a CMS integration.

## What You''ll Learn
- Choose between SSG and SSR appropriately
- Use MDX for rich content with components
- Implement dynamic routing (`[slug]`)
- Optimize images with `next/image`
- Add comprehensive SEO metadata
- Integrate with headless CMS

## Requirements
1. Use SSG for blog listing, SSR for preview mode
2. Support MDX with custom components
3. Dynamic routes for blog posts
4. Optimize all images with `next/image`
5. Full SEO (Open Graph, Twitter cards, JSON-LD)
6. Integrate with CMS (Contentful, Sanity, or Notion)

## Route Structure
```
/                     → Home (SSG)
/blog                 → Blog listing (SSG with revalidate)
/blog/[slug]          → Blog post (SSG with fallback)
/api/preview          → Preview mode toggle
/api/revalidate       → On-demand revalidation
```

## MDX Components
```jsx
// Custom components available in MDX
<Callout type="info">Important note here</Callout>
<CodeBlock language="javascript" filename="example.js">
  const x = 1;
</CodeBlock>
<YouTubeEmbed id="dQw4w9WgXcQ" />
```

## Getting Started
```bash
git clone https://github.com/devarena/nextjs-blog-starter
cd nextjs-blog-starter
npm install
cp .env.example .env.local
npm run dev
```'
WHERE id = 'challenge-14';
