-- +goose Up
-- Seed sample content for the admin console's vocabulary and videos tabs,
-- which otherwise start completely empty (unlike lessons/scenarios, which
-- already had seed data from earlier migrations). A few rows are left as
-- draft on purpose so the publish/unpublish action has something to do.

-- Vocabulary categories -------------------------------------------------------
INSERT INTO vocabulary_categories (id, name, description) VALUES
 ('55555555-0000-4000-8000-000000000001','Backend & Database','Server-side and data-storage terms.'),
 ('55555555-0000-4000-8000-000000000002','Frontend & UI','Client-side and interface terms.'),
 ('55555555-0000-4000-8000-000000000003','DevOps & Cloud','Deployment, infra and operations terms.'),
 ('55555555-0000-4000-8000-000000000004','AI & Machine Learning','Model training and inference terms.'),
 ('55555555-0000-4000-8000-000000000005','Distributed Systems','Multi-node and consistency terms.')
ON CONFLICT (id) DO NOTHING;

-- Dictionary entries (shared vocabulary_items, admin-curated) ----------------
INSERT INTO vocabulary_items (id, term, definition, cefr, domain, status) VALUES
 ('66666666-0000-4000-8000-000000000001','idempotent','An operation that can be applied multiple times without changing the result beyond the first application.','C1','backend_db','published'),
 ('66666666-0000-4000-8000-000000000002','connection pooling','Reusing a set of open database connections instead of opening a new one per request.','B1','backend_db','published'),
 ('66666666-0000-4000-8000-000000000003','cache invalidation','Removing or marking stale a cached value so the next read fetches fresh data.','B2','backend_db','published'),
 ('66666666-0000-4000-8000-000000000004','sharding','Splitting a dataset across multiple databases or nodes to spread load.','C1','distributed_systems','published'),
 ('66666666-0000-4000-8000-000000000005','eventual consistency','A guarantee that replicas converge to the same value given no new writes, but not immediately.','C1','distributed_systems','draft'),
 ('66666666-0000-4000-8000-000000000006','container orchestration','Automated deployment, scaling and management of containerized applications.','B2','devops_cloud','published'),
 ('66666666-0000-4000-8000-000000000007','blue-green deployment','A release strategy that runs two production environments and switches traffic between them.','B2','devops_cloud','draft'),
 ('66666666-0000-4000-8000-000000000008','load balancer','A component that distributes incoming requests across multiple servers.','A2','devops_cloud','published'),
 ('66666666-0000-4000-8000-000000000009','responsive design','Layout that adapts to different screen sizes.','B1','frontend_ui','published'),
 ('66666666-0000-4000-8000-00000000000a','hydration','Attaching client-side interactivity to server-rendered HTML.','B2','frontend_ui','published'),
 ('66666666-0000-4000-8000-00000000000b','gradient descent','An optimization algorithm that iteratively moves parameters toward lower loss.','C1','ai_ml','published'),
 ('66666666-0000-4000-8000-00000000000c','overfitting','A model that fits training data too closely and generalizes poorly.','B2','ai_ml','draft')
ON CONFLICT (id) DO NOTHING;

-- System flashcard decks -------------------------------------------------------
INSERT INTO vocabulary_decks (id, user_id, category_id, name, description, thumbnail_url, level, is_default, status) VALUES
 ('77777777-0000-4000-8000-000000000001','system','55555555-0000-4000-8000-000000000001','Backend Essentials','Core server-side vocabulary for daily standups and code review.','','B1',true,'published'),
 ('77777777-0000-4000-8000-000000000002','system','55555555-0000-4000-8000-000000000003','Cloud & DevOps Starter','Deployment and infra terms for onboarding engineers.','','B2',true,'published'),
 ('77777777-0000-4000-8000-000000000003','system','55555555-0000-4000-8000-000000000004','AI/ML Fundamentals','Vocabulary for reading ML papers and design docs.','','C1',true,'draft')
ON CONFLICT (id) DO NOTHING;

-- Cards inside the system decks ------------------------------------------------
INSERT INTO vocabulary_deck_items (id, deck_id, phrase, normalized_phrase, meaning, example_sentence, note) VALUES
 ('88888888-0000-4000-8000-000000000001','77777777-0000-4000-8000-000000000001','idempotent','idempotent','can be applied repeatedly with the same result','Retrying an idempotent request is always safe.',''),
 ('88888888-0000-4000-8000-000000000002','77777777-0000-4000-8000-000000000001','connection pooling','connection pooling','reusing open DB connections instead of opening new ones','We enabled connection pooling to cut latency under load.',''),
 ('88888888-0000-4000-8000-000000000003','77777777-0000-4000-8000-000000000002','load balancer','load balancer','distributes traffic across servers','The load balancer took the unhealthy instance out of rotation.',''),
 ('88888888-0000-4000-8000-000000000004','77777777-0000-4000-8000-000000000002','blue-green deployment','blue-green deployment','release strategy swapping two live environments','We rolled back instantly thanks to the blue-green deployment.',''),
 ('88888888-0000-4000-8000-000000000005','77777777-0000-4000-8000-000000000003','gradient descent','gradient descent','optimization algorithm minimizing loss step by step','Gradient descent converged after 200 epochs.','')
ON CONFLICT (id) DO NOTHING;

-- Video categories --------------------------------------------------------------
INSERT INTO video_categories (id, slug, name) VALUES
 ('99999999-0000-4000-8000-000000000001','trailer','Trailer'),
 ('99999999-0000-4000-8000-000000000002','podcast','Podcast Clip'),
 ('99999999-0000-4000-8000-000000000003','interview','Interview'),
 ('99999999-0000-4000-8000-000000000004','tutorial','Tutorial')
ON CONFLICT (id) DO NOTHING;

-- Video exercises ----------------------------------------------------------------
INSERT INTO video_exercises (id, category_id, title, description, video_url, thumbnail_url, cefr_level, duration, status) VALUES
 ('aaaaaaaa-0000-4000-8000-000000000001','99999999-0000-4000-8000-000000000001','Startup Pitch Trailer','A 90-second teaser for a product demo day.','https://example.com/videos/startup-pitch-trailer.mp4','','B1',95,'published'),
 ('aaaaaaaa-0000-4000-8000-000000000002','99999999-0000-4000-8000-000000000002','Tech Podcast: Microservices','A clip from a podcast discussing microservice migration.','https://example.com/videos/podcast-microservices.mp4','','B2',420,'published'),
 ('aaaaaaaa-0000-4000-8000-000000000003','99999999-0000-4000-8000-000000000003','Engineering Interview Walkthrough','A mock system-design interview walkthrough.','https://example.com/videos/interview-walkthrough.mp4','','B2',300,'draft'),
 ('aaaaaaaa-0000-4000-8000-000000000004','99999999-0000-4000-8000-000000000004','Docker Basics Tutorial','An introduction to containers and images.','https://example.com/videos/docker-basics.mp4','','A2',180,'published'),
 ('aaaaaaaa-0000-4000-8000-000000000005','99999999-0000-4000-8000-000000000004','Kubernetes Deep Dive','An advanced tour of pods, services and deployments.','https://example.com/videos/kubernetes-deep-dive.mp4','','C1',600,'draft')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM video_exercises WHERE id IN (
  'aaaaaaaa-0000-4000-8000-000000000001','aaaaaaaa-0000-4000-8000-000000000002',
  'aaaaaaaa-0000-4000-8000-000000000003','aaaaaaaa-0000-4000-8000-000000000004',
  'aaaaaaaa-0000-4000-8000-000000000005');
DELETE FROM video_categories WHERE id IN (
  '99999999-0000-4000-8000-000000000001','99999999-0000-4000-8000-000000000002',
  '99999999-0000-4000-8000-000000000003','99999999-0000-4000-8000-000000000004');
DELETE FROM vocabulary_deck_items WHERE id IN (
  '88888888-0000-4000-8000-000000000001','88888888-0000-4000-8000-000000000002',
  '88888888-0000-4000-8000-000000000003','88888888-0000-4000-8000-000000000004',
  '88888888-0000-4000-8000-000000000005');
DELETE FROM vocabulary_decks WHERE id IN (
  '77777777-0000-4000-8000-000000000001','77777777-0000-4000-8000-000000000002',
  '77777777-0000-4000-8000-000000000003');
DELETE FROM vocabulary_items WHERE id IN (
  '66666666-0000-4000-8000-000000000001','66666666-0000-4000-8000-000000000002',
  '66666666-0000-4000-8000-000000000003','66666666-0000-4000-8000-000000000004',
  '66666666-0000-4000-8000-000000000005','66666666-0000-4000-8000-000000000006',
  '66666666-0000-4000-8000-000000000007','66666666-0000-4000-8000-000000000008',
  '66666666-0000-4000-8000-000000000009','66666666-0000-4000-8000-00000000000a',
  '66666666-0000-4000-8000-00000000000b','66666666-0000-4000-8000-00000000000c');
DELETE FROM vocabulary_categories WHERE id IN (
  '55555555-0000-4000-8000-000000000001','55555555-0000-4000-8000-000000000002',
  '55555555-0000-4000-8000-000000000003','55555555-0000-4000-8000-000000000004',
  '55555555-0000-4000-8000-000000000005');
