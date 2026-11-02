package composition

import (
	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/modules/attempts"
	attemptscontrollers "engflex-api/internal/modules/attempts/controllers"
	attemptsservices "engflex-api/internal/modules/attempts/services"
	"engflex-api/internal/modules/bookmarks"
	bookmarkscontrollers "engflex-api/internal/modules/bookmarks/controllers"
	bookmarksservices "engflex-api/internal/modules/bookmarks/services"
	"engflex-api/internal/modules/conversations"
	conversationscontrollers "engflex-api/internal/modules/conversations/controllers"
	conversationsservices "engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/modules/exams"
	examscontrollers "engflex-api/internal/modules/exams/controllers"
	examsservices "engflex-api/internal/modules/exams/services"
	"engflex-api/internal/modules/lessons"
	lessonscontrollers "engflex-api/internal/modules/lessons/controllers"
	lessonsservices "engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/modules/payments"
	paymentscontrollers "engflex-api/internal/modules/payments/controllers"
	paymentsservices "engflex-api/internal/modules/payments/services"
	"engflex-api/internal/modules/profiles"
	profilescontrollers "engflex-api/internal/modules/profiles/controllers"
	profilesservices "engflex-api/internal/modules/profiles/services"
	"engflex-api/internal/modules/scenarios"
	scenarioscontrollers "engflex-api/internal/modules/scenarios/controllers"
	scenariosservices "engflex-api/internal/modules/scenarios/services"
	"engflex-api/internal/modules/stories"
	storiescontrollers "engflex-api/internal/modules/stories/controllers"
	storiesservices "engflex-api/internal/modules/stories/services"
	"engflex-api/internal/modules/users"
	userscontrollers "engflex-api/internal/modules/users/controllers"
	usersservices "engflex-api/internal/modules/users/services"
	"engflex-api/internal/modules/videos"
	videoscontrollers "engflex-api/internal/modules/videos/controllers"
	videosservices "engflex-api/internal/modules/videos/services"
	"engflex-api/internal/modules/vocabulary"
	vocabularycontrollers "engflex-api/internal/modules/vocabulary/controllers"
	vocabularyservices "engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/voice"
	"engflex-api/internal/voice/pronunciation"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterModules(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config, sessionLimiter gin.HandlerFunc) {
	attemptRepository := repositories.NewAttemptRepository(db)
	bookmarkRepository := repositories.NewBookmarkRepository(db)
	conversationRepository := repositories.NewConversationRepository(db)
	conversationTurnRepository := repositories.NewConversationTurnRepository(db)
	examCategoryRepository := repositories.NewExamCategoryRepository(db)
	examTestRepository := repositories.NewExamTestRepository(db)
	feedbackRepository := repositories.NewFeedbackRepository(db)
	lessonActivityRepository := repositories.NewLessonActivityRepository(db)
	lessonRepository := repositories.NewLessonRepository(db)
	lessonSectionRepository := repositories.NewLessonSectionRepository(db)
	paymentTransactionRepository := repositories.NewPaymentTransactionRepository(db)
	personaRepository := repositories.NewPersonaRepository(db)
	planRepository := repositories.NewPlanRepository(db)
	premiumFeatureRepository := repositories.NewPremiumFeatureRepository(db)
	scenarioRepository := repositories.NewScenarioRepository(db)
	scenarioTopicRepository := repositories.NewScenarioTopicRepository(db)
	storyRepository := repositories.NewStoryRepository(db)
	subscriptionRepository := repositories.NewSubscriptionRepository(db)
	transactionManager := repositories.NewTransactionManager(db)
	userDirectoryRepository := repositories.NewUserDirectoryRepository(db)
	userProfileRepository := repositories.NewUserProfileRepository(db)
	userVocabularyRepository := repositories.NewUserVocabularyRepository(db)
	videoCategoryRepository := repositories.NewVideoCategoryRepository(db)
	videoExerciseRepository := repositories.NewVideoExerciseRepository(db)
	videoTranscriptRepository := repositories.NewVideoTranscriptRepository(db)
	vocabularyCategoryRepository := repositories.NewVocabularyCategoryRepository(db)
	vocabularyDeckItemRepository := repositories.NewVocabularyDeckItemRepository(db)
	vocabularyDeckRepository := repositories.NewVocabularyDeckRepository(db)
	vocabularyItemRepository := repositories.NewVocabularyItemRepository(db)

	voiceClient := voice.NewHTTPVoiceClient(cfg.Voice)
	pronounceClient := voice.NewHTTPPronounceEngine(cfg.Voice.PronounceURL, cfg.Voice.PronounceTimeout)
	llmClient := llm.NewHTTPLLMClient(cfg.LLM)

	premiumSvc := paymentsservices.NewSubscriptionService(subscriptionRepository, planRepository)
	scenarioPolicy := scenariosservices.NewScenarioService(scenarioRepository, scenarioTopicRepository, premiumSvc)
	{

		svc := attemptsservices.NewProgressService(
			attemptRepository,
			conversationRepository,
			userProfileRepository,
		)
		ctl := attemptscontrollers.NewProgressController(svc)

		attempts.RegisterRoutes(rg, ctl)
	}
	{

		repo := bookmarkRepository
		svc := bookmarksservices.NewBookmarkService(repo)
		ctl := bookmarkscontrollers.NewBookmarkController(svc)

		bookmarks.RegisterRoutes(rg, ctl)
	}
	{
		cfg := cfg.Voice

		repo := conversationRepository
		turns := conversationTurnRepository
		feedbacks := feedbackRepository
		engine := voiceClient
		pronounce := pronounceClient
		svc := conversationsservices.NewConversationService(repo, turns, engine, pronounce, cfg.MaxDurationSec, feedbacks,
			scenarioRepository, personaRepository, scenarioPolicy)
		ctl := conversationscontrollers.NewConversationController(svc, cfg.MaxDurationSec)

		cb := conversationscontrollers.NewVoiceCallbackController(svc)
		tcb := conversationscontrollers.NewTurnCallbackController(svc)

		conversations.RegisterRoutes(rg, ctl, cb, tcb, sessionLimiter, cfg.InternalSecret)
	}
	{

		categories := examCategoryRepository
		tests := examTestRepository
		attempts := attemptRepository
		ctl := examscontrollers.NewExamController(examsservices.NewCatalogService(categories, tests), examsservices.NewAttemptService(attempts, tests, transactionManager))

		exams.RegisterRoutes(rg, ctl)
	}
	{
		media := cfg.Media

		lessonSvc := lessonsservices.NewLessonService(
			lessonSectionRepository,
			lessonRepository,
			premiumSvc,
		)
		lessonCtl := lessonscontrollers.NewLessonController(lessonSvc, media)

		activitySvc := lessonsservices.NewLessonActivityService(
			lessonActivityRepository,
			pronunciation.NewScorer(pronounceClient),
			llmClient,
			attemptRepository,
			transactionManager,
			lessonSvc,
		)
		activityCtl := lessonscontrollers.NewLessonActivityController(activitySvc, media)

		lessons.RegisterRoutes(rg, lessonCtl, activityCtl, sessionLimiter)
	}
	{
		cfg := cfg.Payment

		subRepo := subscriptionRepository
		txnRepo := paymentTransactionRepository
		planRepo := planRepository
		featureRepo := premiumFeatureRepository

		subSvc := premiumSvc
		planSvc := paymentsservices.NewPlanService(planRepo)
		featureSvc := paymentsservices.NewPremiumFeatureService(featureRepo)
		vnpaySvc := paymentsservices.NewVNPayService(paymentsservices.VNPayConfig{
			TmnCode:    cfg.VNPayTmnCode,
			HashSecret: cfg.VNPayHashSecret,
			PayURL:     cfg.VNPayPayURL,
			ReturnURL:  cfg.VNPayReturnURL,
		}, transactionManager, subRepo, txnRepo, planRepo)

		paymentCtl := paymentscontrollers.NewPaymentController(planSvc, subSvc, vnpaySvc)
		planCtl := paymentscontrollers.NewPlanController(planSvc)
		featureCtl := paymentscontrollers.NewPremiumFeatureController(featureSvc)
		subAdminCtl := paymentscontrollers.NewSubscriptionAdminController(subSvc)

		payments.RegisterRoutes(rg, paymentCtl, planCtl, featureCtl, subAdminCtl)
	}
	{

		svc := profilesservices.NewProfileService(userProfileRepository)
		ctl := profilescontrollers.NewProfileController(svc)

		profiles.RegisterRoutes(rg, ctl)
	}
	{

		svc := scenarioPolicy
		ctl := scenarioscontrollers.NewScenarioController(svc)

		personaSvc := scenariosservices.NewPersonaService(personaRepository)
		personaCtl := scenarioscontrollers.NewPersonaController(personaSvc)

		scenarios.RegisterRoutes(rg, ctl, personaCtl)
	}
	{

		repo := storyRepository
		svc := storiesservices.NewStoryService(repo)
		ctl := storiescontrollers.NewStoryController(svc)

		stories.RegisterRoutes(rg, ctl)
	}
	{

		subSvc := premiumSvc
		attemptRepo := attemptRepository
		svc := usersservices.NewUserService(userDirectoryRepository, attemptRepo)
		ctl := userscontrollers.NewUserController(svc, subSvc)

		users.RegisterRoutes(rg, ctl)
	}
	{

		exerciseSvc := videosservices.NewVideoExerciseService(videoExerciseRepository, premiumSvc)
		exerciseCtl := videoscontrollers.NewVideoExerciseController(exerciseSvc)

		attemptSvc := videosservices.NewVideoAttemptService(attemptRepository, videoExerciseRepository, exerciseSvc)
		attemptCtl := videoscontrollers.NewVideoAttemptController(attemptSvc)

		transcriptSvc := videosservices.NewVideoTranscriptService(videoTranscriptRepository, exerciseSvc)
		transcriptCtl := videoscontrollers.NewVideoTranscriptController(transcriptSvc)

		categorySvc := videosservices.NewVideoCategoryService(videoCategoryRepository)
		categoryCtl := videoscontrollers.NewVideoCategoryController(categorySvc)

		videos.RegisterRoutes(rg, exerciseCtl, attemptCtl, transcriptCtl, categoryCtl)
	}
	{

		categorySvc := vocabularyservices.NewVocabularyCategoryService(vocabularyCategoryRepository)
		categoryCtl := vocabularycontrollers.NewVocabularyCategoryController(categorySvc)

		deckSvc := vocabularyservices.NewVocabularyDeckService(vocabularyDeckRepository, premiumSvc)
		deckCtl := vocabularycontrollers.NewVocabularyDeckController(deckSvc)

		itemSvc := vocabularyservices.NewVocabularyDeckItemService(vocabularyDeckItemRepository, vocabularyDeckRepository, vocabularyItemRepository, premiumSvc)
		itemCtl := vocabularycontrollers.NewVocabularyDeckItemController(itemSvc)

		itemRepo := vocabularyItemRepository
		dictSvc := vocabularyservices.NewVocabularyItemService(itemRepo, premiumSvc)
		dictCtl := vocabularycontrollers.NewVocabularyItemController(dictSvc)

		stateSvc := vocabularyservices.NewUserVocabularyService(userVocabularyRepository, dictSvc)
		stateCtl := vocabularycontrollers.NewUserVocabularyController(stateSvc)

		vocabulary.RegisterRoutes(rg, categoryCtl, deckCtl, itemCtl, dictCtl, stateCtl)
	}
}
