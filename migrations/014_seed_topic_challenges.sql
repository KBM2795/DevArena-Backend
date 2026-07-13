-- ============================================================================
-- Seed: Topic-only Challenges (Community Prompts / Creator Challenges)
-- ============================================================================

INSERT INTO challenges (
    id, title, description, difficulty, type, max_score, 
    repo_template_url, requirements, tech_stack, estimated_hours, 
    is_published, created_at, updated_at, description_md, rubric
) VALUES 
-- 1. Collaborative Canvas (r/place-style App)
(
    '2',
    'Collaborative Canvas (r/place-style App)',
    'Build a real-time collaborative pixel drawing canvas where users place colored tiles on a shared grid.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Implement a shared grid canvas where multiple users can place colored pixels simultaneously", "Introduce a rate-limit/cooldown period per user between pixel placement", "Synchronize grid state in real-time across all clients", "Include features for zooming and panning across the canvas"]',
    '["WebSockets", "HTML5 Canvas", "React/Vue/Svelte", "Node.js", "Redis/Database"]',
    12,
    TRUE,
    NOW(),
    NOW(),
    $$# Collaborative Canvas (r/place-style App)

Build a real-time collaborative pixel drawing canvas where multiple users can draw together on a shared grid, inspired by Reddit's r/place.

## Overview
This challenge focuses on building a highly interactive, real-time shared drawing board. Users can click any pixel on a large grid (e.g., 500x500 or 1000x1000 pixels) and paint it with a color. However, to keep it fun and prevent spam, each user should have a cooldown timer (e.g., 30 seconds) before they can place another tile.

## Features to Build
1. **Shared Interactive Grid**: A canvas UI allowing users to hover over pixels, select colors from a palette, and submit their selection.
2. **Real-time Synchronization**: Every pixel update must be instantly visible to all other connected clients.
3. **Zoom & Pan controls**: Let users navigate a large canvas layout comfortably on both desktop and mobile devices.
4. **Cooldown Mechanism**: Restrict users from painting infinitely. Show a visual timer/countdown between tile placements.

## Core Technologies
- **WebSockets / Socket.io / Server-Sent Events** for live synchronization.
- **HTML5 Canvas API** or a performance-optimized SVG/CSS Grid for rendering the pixels.
- **In-Memory Caching (Redis/Memory-Store)** to quickly read/write the grid state and track rate limits.
$$,
    '{}'::jsonb
),

-- 2. Browser Homepage / New Tab
(
    '3',
    'Custom Browser Homepage / New Tab Page',
    'Create a personalized custom browser start page or dashboard widget system.',
    'Easy',
    'topic',
    100,
    NULL,
    '["Build a responsive dashboard layout optimized for new tabs", "Provide custom quick-links/bookmarks and keyboard shortcuts", "Include useful widgets: to-do list, weather widget, and custom drawing canvas", "Support customizable themes or background wallpaper selection"]',
    '["HTML5", "CSS Grid/Flexbox", "Local Storage", "Vanilla JS / React", "Weather API"]',
    4,
    TRUE,
    NOW(),
    NOW(),
    $$# Custom Browser Homepage / New Tab Page

Design and build a personalized homepage start screen or browser extension replacement dashboard.

## Overview
Instead of a blank screen when opening a new tab, users want a dashboard that keeps them productive. This challenge encourages building widgets, customizing layouts, and saving state locally.

## Features to Build
1. **Personalization Hub**: Quick access shortcuts with customizable keyboard bindings.
2. **Widgets**: A standard To-Do checklist, clock, local weather forecasting widget, and a mini drawing canvas/notepad.
3. **Theme & Background Manager**: Allow users to upload custom wallpapers or select gradient theme styles.
4. **State Persistence**: Preserve widgets state and layout settings across browser sessions using localStorage or cookies.
$$,
    '{}'::jsonb
),

-- 3. Live Data Visualization Dashboard
(
    '4',
    'Live Data Visualization Dashboard',
    'Build an interactive dashboard that fetches and renders real-time streaming data.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Fetch and process real-time streaming data from a public API", "Render data dynamically using line/bar charts, maps, or sparklines", "Support search and sorting, filtering options, and custom alert thresholds", "Handle loading, error, and offline states gracefully"]',
    '["React/Vue", "Chart.js / Recharts / D3.js", "WebSocket / Polling APIs", "Tailwind CSS"]',
    8,
    TRUE,
    NOW(),
    NOW(),
    $$# Live Data Visualization Dashboard

Build an interactive dashboard that connects to live data streams (e.g., financial stock prices, cryptocurrency trackers, weather data, or real-time news trends) and visualizes them beautifully.

## Overview
Data is most useful when it is responsive and visual. This challenge focuses on fetching data, keeping it refreshed, and presenting it using clean visualization charts.

## Features to Build
1. **Live Charting**: Render trends dynamically using line, bar, or area charts that update as new values come in.
2. **Interactive Filters**: Allow users to sort, search, and filter records by categories, timeframes, or values.
3. **Alert Constraints**: Let users define rules (e.g., alert me if Bitcoin crosses $80,000) and display active warnings.
4. **Responsive Layout**: A modern grid dashboard that dynamically shifts elements based on screen resolutions.
$$,
    '{}'::jsonb
),

-- 4. Web Game (2D Platformer / Puzzle Game)
(
    '5',
    'Browser-Based 2D Game',
    'Develop a simple interactive 2D browser game with user inputs, animations, and sound effects.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Implement game loops, player collision detection, and user input movement controls", "Include sound effects, custom background music, and score tracking", "Build at least 3 distinct levels or increasing difficulty stages", "Provide a leaderboard or simple high score board"]',
    '["JavaScript Canvas API", "HTML5 Audio", "CSS Animations", "Game Engine / Physics"]',
    10,
    TRUE,
    NOW(),
    NOW(),
    $$# Browser-Based 2D Game

Develop an interactive browser game using vanilla JavaScript or canvas libraries. It could be a side-scrolling platformer, a puzzle game, or a simple top-down adventure.

## Overview
This challenge exercises your core logic and game development skills. Focus on user input response, collision accuracy, animations, state management (start, play, game over), and audio integration.

## Features to Build
1. **Responsive Game Loop**: A standard requestAnimationFrame-driven loop separating update and render cycles.
2. **Movement & Physics**: Horizontal/vertical movement, jumping mechanics, gravity physics, and block/boundary collision detection.
3. **Audio & VFX**: Play custom sounds on specific events (jump, collect coin, game over) and animate sprite sheets.
4. **Levels & States**: Maintain menu screen, instructions modal, game state, and high-scores.
$$,
    '{}'::jsonb
),

-- 5. Study Helper AI
(
    '6',
    'AI-Powered Study Helper',
    'Create an application that processes uploaded notes and text to generate summary materials and flashcards.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Integrate an LLM API (OpenAI, Gemini, Claude, or local) to process note uploads", "Generate structured study summaries, quiz questions, and flashcards", "Design an interactive UI to test users on the generated flashcards", "Preserve past summaries and study progress in a database"]',
    '["React/Next.js", "AI/LLM API Integration", "PostgreSQL / MongoDB", "Tailwind CSS"]',
    8,
    TRUE,
    NOW(),
    NOW(),
    $$# AI-Powered Study Helper

Build an intelligent companion app that accepts text or PDF study notes, summarizes them, and tests the user using dynamically generated flashcards and quizzes.

## Overview
Leverage modern large language models to build a productivity companion. The app takes raw inputs and creates an interactive review flow.

## Features to Build
1. **Content Upload**: Paste text directly or upload document files.
2. **AI Material Generator**: Call an LLM endpoint with prompt templates to generate summaries, multiple-choice quizzes, and review flashcards.
3. **Interactive Study Mode**: A flashcard flipper UI, along with a quiz game that scores responses.
4. **Study History**: Track score histories and completed topics.
$$,
    '{}'::jsonb
),

-- 6. Personalized Meal Planner & Recipe Finder
(
    '7',
    'AI Meal Planner & Recipe Finder',
    'Build an AI-powered meal planner that generates weekly recipes based on inventory and dietary constraints.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Provide inputs for dietary restrictions, preferences, and current pantry ingredients", "Utilize AI to generate recipes fitting the user constraints", "Generate a shopping list of missing ingredients automatically", "Enable exporting or scheduling email digests of the planned recipes"]',
    '["Node.js / Python", "React", "OpenAI / LLM API", "Nodemailer / SendGrid", "Database"]',
    6,
    TRUE,
    NOW(),
    NOW(),
    $$# AI Meal Planner & Recipe Finder

Build an intelligent application that crafts custom meal plans and helps users find recipes using whatever ingredients they have in their fridge.

## Overview
Reduce food waste and automate dinner plans. Users specify their allergies, goals (e.g. low-carb), and inventory, and the AI designs recipes and compiles a grocery checklist.

## Features to Build
1. **Pantry Inventory Manager**: Add, update, and remove raw ingredients currently on-hand.
2. **Constraint Forms**: Configure dietary boundaries (e.g., vegan, gluten-free, nut allergies) and time limits.
3. **Meal Schedule Generator**: Present daily breakfast, lunch, and dinner cards with step-by-step cooking directions.
4. **Export & Sharing**: Export recipes as PDF or send them to the user's inbox.
$$,
    '{}'::jsonb
),

-- 7. Handwritten Text Recognition
(
    '8',
    'Handwritten Text Recognition (OCR)',
    'Develop an interactive canvas to draw letters and numbers and transcribe them into digital text.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Implement an interactive canvas for users to write characters with mouse/touch inputs", "Integrate OCR models or neural networks to recognize characters", "Provide a real-time transcription log of recognized sentences", "Support exporting the transcribed texts to file formats"]',
    '["HTML5 Canvas", "Python / Node.js Backend", "TensorFlow.js / Tesseract.js", "React"]',
    10,
    TRUE,
    NOW(),
    NOW(),
    $$# Handwritten Text Recognition (OCR)

Build an OCR tool where users can draw/write letters, numbers, or full words directly on-screen, and watch an AI model transcribe it into structured digital text in real-time.

## Overview
This challenge covers input drawing canvases and basic machine learning text recognition APIs or client-side libraries.

## Features to Build
1. **Interactive Notebook Canvas**: A fluid sketchpad supporting brush adjustments and erase options.
2. **Handwriting Predictor**: Process canvas snapshots through a library like Tesseract.js, or send image matrices to a customized ML API.
3. **Live Transcription Pane**: Display recognized characters as the user writes, letting them edit or copy the result.
4. **Export Options**: Export the resulting text file or save sketches as PNG drawings.
$$,
    '{}'::jsonb
),

-- 8. Hobby-Specific AI Assistant
(
    '9',
    'Hobby-Specific AI Assistant / Niche Agent',
    'Create an AI tool or agent that addresses a niche, real-world utility scenario (e.g., audio normalization).',
    'Hard',
    'topic',
    200,
    NULL,
    '["Define a specific niche domain challenge (audio compression, calendar agent, code assistant, etc.)", "Integrate specialized AI APIs or libraries targeting the utility action", "Create a highly tuned UI suited for the specific task metrics", "Support exporting files, actions, or logs generated by the agent"]',
    '["Next.js / Python", "AI API Tool-Calling", "Niche Domain Library (e.g. Web Audio API)", "Tailwind"]',
    8,
    TRUE,
    NOW(),
    NOW(),
    $$# Hobby-Specific AI Assistant / Niche Agent

Create a specialized AI assistant geared towards solving a specific hobbyist or professional task—such as a smart audio normalizer, a personalized fitness scheduler, or a code debugger.

## Overview
Instead of a generic chatbot, this challenge focuses on target-specific AI integrations that manipulate external APIs, parse audio/video streams, or run complex scheduling tools.

## Features to Build
1. **Domain Input Console**: Customized widgets corresponding to the hobby (e.g., audio file uploads, calendars, audio players).
2. **AI Action Pipelines**: Connect prompts to execution models (like JSON tool-calling / function-calling) that process inputs.
3. **Interactive Control Dashboard**: Display metrics, changes, and logs generated by the assistant.
4. **Workflow Export**: Download the optimized audio track, updated schedule, or output templates.
$$,
    '{}'::jsonb
),

-- 9. n8n/Make Automation Flows
(
    '10',
    'n8n / Make Automation Pipeline Flow',
    'Design an end-to-end automated data pipeline that runs tasks across multiple platforms.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Configure an automation sequence in a platform like n8n or Make", "Trigger workflows using specific webhooks, schedules, or events", "Process datasets with AI/transformation nodes", "Integrate at least 3 third-party systems or notification channels"]',
    '["n8n / Make.com", "API Integrations", "Webhooks", "JSON Transformation", "AI APIs"]',
    6,
    TRUE,
    NOW(),
    NOW(),
    $$# n8n / Make Automation Pipeline Flow

Design, implement, and document an end-to-end automated workflow that fetches data, runs integrations, and sends alerts.

## Overview
No-code/low-code tools are extremely powerful for automation. This challenge tests your capability to coordinate APIs and format JSON objects across multiple third-party systems.

## Example Flow to Build
- **Trigger**: Monitor a platform (e.g., a Job Board API, a sub-reddit, or incoming emails).
- **Process**: Parse text content, pass it to an LLM node to categorize and rate relevance, and filter out low-value matches.
- **Action**: Compile summaries and dispatch notifications to Slack, Discord, or an email digest.
$$,
    '{}'::jsonb
),

-- 10. Personal Task Automation
(
    '11',
    'Personal Task Automation Flow',
    'Build an automation system to handle daily routines, like logging expenses or booking events.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Create an automated capture utility (e.g. Chatbot, Email trigger, Webhook link)", "Parse information input streams dynamically", "Format and insert data tables into cloud spreadsheets or calendars", "Configure automated email digests or alerts"]',
    '["n8n / Make", "Google Sheets API / Google Calendar", "Telegram / Discord Bot API", "Mailers"]',
    6,
    TRUE,
    NOW(),
    NOW(),
    $$# Personal Task Automation Flow

Create a system that automates repeating daily routines—such as logging expenditures via chat messages or parsing invoices from email attachments directly into a database.

## Overview
This challenge focuses on saving time on repetitive actions. Leverage Telegram bots, Discord webhooks, or automation triggers to feed databases or sheets automatically.

## Features to Build
1. **Input Interface**: A chat bot or email parser endpoint.
2. **Information Processing Node**: Extract items, numbers, and dates out of incoming inputs.
3. **Sheet/DB Integration**: Append parsed datasets directly into spreadsheets or a database.
4. **Daily Summaries**: Broadcast a daily digest of actions taken to a personal chat channel.
$$,
    '{}'::jsonb
),

-- 11. Cross-App Integration Challenge
(
    '12',
    'Cross-App Integration Pipeline',
    'Build a workflow connecting multiple SaaS applications, prioritizing custom filters and transformations.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Connect multiple independent web applications into a single sync pipeline", "Perform advanced structural JSON mapping and transformations", "Handle rate limit throttling and connection retry logic", "Establish monitoring/logging dashboards to catch workflow breaks"]',
    '["n8n / Make / Custom Scripts", "OAuth 2.0 Integration", "Webhooks", "JSON Transformations"]',
    6,
    TRUE,
    NOW(),
    NOW(),
    $$# Cross-App Integration Pipeline

Connect multiple independent SaaS apps into a single automated pipeline. Focus on formatting data accurately, catching errors, and mapping keys between systems.

## Overview
This challenge focuses on OAuth configurations, API mapping rules, and error recovery setups when synchronizing data across apps.

## Features to Build
1. **App Sync**: Synchronize items (e.g., synchronization of tasks between Trello and GitHub Issues).
2. **Custom Mapper**: Map custom fields and convert date formats.
3. **Fallback Actions**: Define catch-all routes if an external service drops offline.
4. **Audit Logs**: Maintain a dashboard tracking successful updates and logs of failed sync cycles.
$$,
    '{}'::jsonb
),

-- 12. Home Automation / Bots
(
    '13',
    'Home Automation & Intelligent Bots',
    'Build a system that reads environmental inputs or feeds, and controls smart device setups.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Integrate with IoT mockups, sensors, or smart home APIs", "Deploy an LLM-driven voice or chat command interface", "Implement rules-based conditional trigger engines", "Maintain security credentials and authorization safeguards"]',
    '["Node-RED / n8n", "MQTT / WebSockets", "Smart Home APIs (Home Assistant / Local)", "LLM Calling"]',
    8,
    TRUE,
    NOW(),
    NOW(),
    $$# Home Automation & Intelligent Bots

Design a conversational chatbot or automation panel that parses environmental inputs (e.g., weather updates, calendar schedules) and triggers actions on smart devices.

## Overview
Develop custom command flows, set up trigger criteria, and implement natural language controls to interact with connected devices or simulate IoT dashboards.

## Features to Build
1. **Sensor Logs**: Periodically log temperature, weather indicators, or system usage statistics.
2. **Intelligent Command Router**: Feed text requests to an AI engine that maps user intentions to smart device functions.
3. **Conditional Rules**: Set up standard triggers (e.g. if temperature is > 25°C and person is home, turn on AC).
4. **Event Logs**: Provide a real-time terminal widget showing triggered events.
$$,
    '{}'::jsonb
),

-- 13. Mobile Recipe Finder
(
    '14',
    'Mobile Recipe Finder App',
    'Build a cross-platform mobile application that fetches recipes based on keywords, food tags, or ingredients.',
    'Easy',
    'topic',
    100,
    NULL,
    '["Implement a search interface with filters for dietary bounds and ingredients", "Fetch dynamic recipe data from a public food API", "Create a favorites collection that persists local device storage", "Build a grocery shopping list checklist widget based on select recipes"]',
    '["Flutter / React Native", "Public Food API (Spoonacular, etc.)", "Local Storage (Hive, SQLite, SharedPreferences)"]',
    6,
    TRUE,
    NOW(),
    NOW(),
    $$# Mobile Recipe Finder App

Build a responsive cross-platform mobile application that lets users query recipes, organize their menus, and export grocery lists.

## Overview
This challenge tests your mobile layout capabilities, API communication methods, search debounce structures, and local state management.

## Features to Build
1. **Dietary Filter Panel**: Easy toggle buttons for dietary bounds (vegan, keto, gluten-free, etc.).
2. **Search Bar**: Quick-search API query fields with debounce to prevent excessive request firing.
3. **Bookmark Manager**: Save favorite recipe cards to local disk storage for offline lookup.
4. **Interactive Grocery List**: Compile all necessary ingredients from saved recipes into a clean checklist.
$$,
    '{}'::jsonb
),

-- 14. Multiplayer Board Game
(
    '15',
    'Multiplayer Board Game App',
    'Create a mobile board game where players sync game state in real-time over the network or local Bluetooth.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Develop a multiplayer board game (e.g. Chess, Tic-Tac-Write, Robo Rally, etc.)", "Implement network synchronization of player positions and turn states", "Handle disconnection events and game recovery flows", "Create a clean lobby system for joining and starting matches"]',
    '["React Native / Flutter", "WebSockets / Firebase Realtime DB", "State Management (Redux/Zustand)"]',
    12,
    TRUE,
    NOW(),
    NOW(),
    $$# Multiplayer Board Game App

Build a synchronized multiplayer mobile game where friends can connect, choose game lobbies, and play a turn-based board game together.

## Overview
Coordinate game logic and states across multiple physical devices. Focus on game turns, synchronization accuracy, and handling sudden disconnections.

## Features to Build
1. **Lobby Matchmaker**: Create room codes or scan local connections to join matches.
2. **Turn-based Engine**: Ensure strict validation rules so players can only move during their active turns.
3. **Game State Synchronization**: Send move actions to a coordinate server and update all connected boards instantly.
4. **Reconnection handler**: Gracefully restore game states if a player temporarily loses signal.
$$,
    '{}'::jsonb
),

-- 15. Fitness or Habit Tracker
(
    '16',
    'Mobile Fitness & Habit Tracker',
    'Develop a habit-forming mobile application with notifications and daily goal progress visualizations.',
    'Medium',
    'topic',
    150,
    NULL,
    '["Build an interactive habit logging interface supporting daily streaks", "Integrate device native APIs (pedometer data or local push notifications)", "Render goal completion status using visual rings, graphs, or widgets", "Save user habit tracking histories to persistent device storage"]',
    '["Flutter / React Native", "Device APIs (Pedometer, Notifications)", "SQLite / Hive / Room DB"]',
    8,
    TRUE,
    NOW(),
    NOW(),
    $$# Mobile Fitness & Habit Tracker

Create a clean mobile application that helps users build consistent routines, track health stats (like step count or water intake), and view visual analytics.

## Overview
This challenge focuses on notification scheduling, reading native mobile device sensors (pedometer), and presenting data analytics on-screen.

## Features to Build
1. **Habit Routine Manager**: Add custom habits with specific target frequencies (e.g., drink water 5x daily).
2. **Local Reminders**: Trigger local push notifications at custom times reminding users to log their habits.
3. **Native Sensors**: Read device step counters or location details if permitted.
4. **Interactive Dashboard**: Build rings or progress bars displaying weekly logs.
$$,
    '{}'::jsonb
),

-- 16. Augmented Reality (AR) App
(
    '17',
    'Mobile Augmented Reality (AR) App',
    'Develop a mobile AR app that places interactive 3D assets onto real-world flat surfaces.',
    'Hard',
    'topic',
    200,
    NULL,
    '["Implement AR camera views utilizing plane detection frameworks", "Let users place, rotate, and scale 3D objects on flat surfaces", "Create touch gestures to interact with placed 3D objects", "Design a menu library of multiple 3D assets to choose from"]',
    '["React Native / Flutter", "ARKit / ARCore", "Unity / Three.js / Sceneform", "3D asset files"]',
    10,
    TRUE,
    NOW(),
    NOW(),
    $$# Mobile Augmented Reality (AR) App

Develop an AR application that lets users select virtual 3D models and position them onto real-world surfaces like tables or floors.

## Overview
This challenge introduces spatial computing, interactive 3D assets, plane detection configurations, and mobile AR camera integrations.

## Features to Build
1. **AR Canvas**: An camera-feed view overlaying interactive nodes.
2. **Plane Detector**: Track and highlight flat floors, walls, or surfaces in real-time.
3. **Object Manipulator**: Select models from a sidebar library and use tap gestures to rotate, scale, or place them.
4. **Physics Interaction**: Let objects drop onto surfaces or react to physical taps.
$$,
    '{}'::jsonb
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    difficulty = EXCLUDED.difficulty,
    type = EXCLUDED.type,
    max_score = EXCLUDED.max_score,
    repo_template_url = EXCLUDED.repo_template_url,
    requirements = EXCLUDED.requirements,
    tech_stack = EXCLUDED.tech_stack,
    estimated_hours = EXCLUDED.estimated_hours,
    is_published = EXCLUDED.is_published,
    description_md = EXCLUDED.description_md,
    rubric = EXCLUDED.rubric,
    updated_at = NOW();

-- Insert new tags required by the topic challenges (ON CONFLICT DO NOTHING)
INSERT INTO tags (id, name, slug, category, color) VALUES
    ('tag-websockets', 'WebSockets', 'websockets', 'backend', '#009688'),
    ('tag-charts', 'Charts & Visuals', 'charts', 'frontend', '#ff9800'),
    ('tag-gamedev', 'Game Dev', 'gamedev', 'frontend', '#e91e63'),
    ('tag-ai', 'Artificial Intelligence', 'ai', 'ai', '#9c27b0'),
    ('tag-automation', 'Automation', 'automation', 'backend', '#607d8b'),
    ('tag-iot', 'Internet of Things', 'iot', 'fundamentals', '#795548'),
    ('tag-mobile', 'Mobile Development', 'mobile', 'frontend', '#3f51b5'),
    ('tag-ar', 'Augmented Reality', 'ar', 'frontend', '#00bcd4')
ON CONFLICT (id) DO NOTHING;

-- Link topic challenges to their respective tags
INSERT INTO challenge_tags (challenge_id, tag_id) VALUES
    -- 1. Collaborative Canvas
    ('2', 'tag-canvas'),
    ('2', 'tag-websockets'),
    -- 2. Browser Homepage
    ('3', 'tag-html'),
    ('3', 'tag-css'),
    -- 3. Live Data Visualization
    ('4', 'tag-charts'),
    ('4', 'tag-react'),
    -- 4. Browser-Based Game
    ('5', 'tag-canvas'),
    ('5', 'tag-gamedev'),
    -- 5. Study Helper AI
    ('6', 'tag-ai'),
    ('6', 'tag-react'),
    -- 6. AI Meal Planner
    ('7', 'tag-ai'),
    ('7', 'tag-python'),
    -- 7. Handwritten OCR
    ('8', 'tag-canvas'),
    ('8', 'tag-tensorflow'),
    ('8', 'tag-ai'),
    -- 8. Hobby AI Assistant
    ('9', 'tag-ai'),
    -- 9. n8n Flow
    ('10', 'tag-automation'),
    ('10', 'tag-system-design'),
    -- 10. Personal Automation
    ('11', 'tag-automation'),
    -- 11. Cross-App Integration
    ('12', 'tag-automation'),
    -- 12. Home Automation
    ('13', 'tag-iot'),
    ('13', 'tag-automation'),
    -- 13. Mobile Recipe Finder
    ('14', 'tag-mobile'),
    ('14', 'tag-databases'),
    -- 14. Multiplayer Board Game
    ('15', 'tag-mobile'),
    ('15', 'tag-websockets'),
    -- 15. Fitness or Habit Tracker
    ('16', 'tag-mobile'),
    -- 16. AR App
    ('17', 'tag-mobile'),
    ('17', 'tag-3d'),
    ('17', 'tag-ar')
ON CONFLICT (challenge_id, tag_id) DO NOTHING;
