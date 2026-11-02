package repositories

import (
	"context"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

func (r *attemptRepository) GetOpenLessonGraph(ctx context.Context, userID, lessonID string) (*models.Attempt, error) {
	if _, err := parseID(lessonID, "lessonId"); err != nil {
		return nil, err
	}
	var attempt models.Attempt
	if err := r.db.WithContext(ctx).Joins("Lesson").Where("attempts.user_id = ? AND attempts.lesson_id = ? AND attempts.type = ? AND attempts.status = ?", userID, lessonID, enums.AttemptTypeLesson, enums.AttemptStatusInProgress).First(&attempt).Error; err != nil {
		return nil, err
	}
	if err := r.attachLessonActivities(ctx, []*models.Attempt{&attempt}); err != nil {
		return nil, err
	}
	return &attempt, nil
}
func (r *attemptRepository) ListLessonProgress(ctx context.Context, userID string) ([]*models.Attempt, error) {
	var attempts []*models.Attempt
	if err := r.db.WithContext(ctx).Select("DISTINCT ON (attempts.lesson_id) attempts.*").Joins("Lesson").Where("attempts.user_id = ? AND attempts.type = ? AND attempts.lesson_id IS NOT NULL", userID, enums.AttemptTypeLesson).Order("attempts.lesson_id, CASE WHEN attempts.status = 'completed' THEN 0 ELSE 1 END, attempts.created_at DESC, attempts.id DESC").Find(&attempts).Error; err != nil {
		return nil, err
	}
	if err := r.attachLessonActivities(ctx, attempts); err != nil {
		return nil, err
	}
	return attempts, nil
}
func (r *attemptRepository) attachLessonActivities(ctx context.Context, attempts []*models.Attempt) error {
	if len(attempts) == 0 {
		return nil
	}
	ids := make([]string, 0, len(attempts))
	lessons := make(map[string][]*models.Lesson, len(attempts))
	for _, attempt := range attempts {
		if attempt.Lesson != nil {
			ids = append(ids, attempt.Lesson.ID)
			lessons[attempt.Lesson.ID] = append(lessons[attempt.Lesson.ID], attempt.Lesson)
			attempt.Lesson.Activities = []*models.LessonActivity{}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var activities []*models.LessonActivity
	if err := r.db.WithContext(ctx).Where("lesson_id IN ?", ids).Order("part_number ASC, id ASC").Find(&activities).Error; err != nil {
		return err
	}
	for _, activity := range activities {
		for _, lesson := range lessons[activity.LessonID] {
			lesson.Activities = append(lesson.Activities, activity)
		}
	}
	return nil
}
