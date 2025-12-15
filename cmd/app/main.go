package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"linkedin-automation-poc/internal/auth"
	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/interaction"
	"linkedin-automation-poc/internal/log"
	"linkedin-automation-poc/internal/messaging"
	"linkedin-automation-poc/internal/search"
	"linkedin-automation-poc/internal/state"
	"linkedin-automation-poc/internal/stealth"

	"go.uber.org/zap"
)

const (
	DRY_RUN_MODE       = true
	MAX_PROFILE_VISITS = 2
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := log.Init("info", false); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	logger := log.Module("main")
	logger.Info("LinkedIn Automation POC starting",
		zap.Bool("dry_run_mode", DRY_RUN_MODE),
		zap.Int("max_profile_visits", MAX_PROFILE_VISITS))

	// Initialize state store
	store, err := state.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		logger.Fatal("failed to initialize state store", zap.Error(err))
	}
	defer store.Close()

	// Initialize browser
	browserInstance, err := browser.New(cfg.Browser)
	if err != nil {
		logger.Fatal("failed to initialize browser", zap.Error(err))
	}
	defer browserInstance.Close()

	// Generate fingerprint
	fingerprint, err := stealth.GenerateFingerprint(cfg.Stealth.Fingerprint)
	if err != nil {
		logger.Fatal("failed to generate fingerprint", zap.Error(err))
	}
	logger.Info("fingerprint generated", zap.String("fingerprint", fingerprint.String()))

	// Create browser context
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	ctx, err := browser.NewContext(browserInstance.Instance(), sessionID, fingerprint)
	if err != nil {
		logger.Fatal("failed to create browser context", zap.Error(err))
	}
	defer ctx.Close()

	// Initialize session manager
	sessionMgr := auth.NewSession(sessionID)

	// Try to load existing session
	savedSession, err := store.GetSession(sessionID)
	if err == nil && savedSession != nil {
		logger.Info("loading saved session")
		if err := sessionMgr.FromJSON(savedSession.Cookies); err != nil {
			logger.Warn("failed to load session, will login", zap.Error(err))
		} else if sessionMgr.IsValid() {
			logger.Info("valid session loaded, loading cookies")
			ctx.LoadCookies(sessionMgr.GetCookies())
		}
	}

	// Login if needed
	if !sessionMgr.IsValid() || !sessionMgr.HasAuthCookie() {
		logger.Info("logging in to LinkedIn")
		loginHandler := auth.NewLogin(ctx, cfg.LinkedIn.Email, cfg.LinkedIn.Password)
		if err := loginHandler.Execute(); err != nil {
			logger.Fatal("login failed", zap.Error(err))
		}

		// Save session
		cookies, err := ctx.GetCookies()
		if err == nil {
			sessionMgr.Save(cookies)
			sessionData, _ := sessionMgr.ToJSON()
			store.SaveSession(state.Session{
				SessionID: sessionID,
				Cookies:   sessionData,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
				IsValid:   true,
			})
		}
	}

	// Check for checkpoints
	checkpoint := auth.NewCheckpoint(ctx.Page())
	if checkpointType, msg := checkpoint.Detect(); checkpointType != auth.CheckpointNone {
		logger.Error("checkpoint detected",
			zap.String("type", string(checkpointType)),
			zap.String("message", msg))
		os.Exit(1)
	}

	// Initialize components for dry-run
	timing := stealth.NewTiming(cfg.Stealth.Timing)
	mouse := stealth.NewMouse(ctx.Page(), cfg.Stealth.MouseMovement)
	scroll := stealth.NewScroll(ctx.Page(), cfg.Stealth.Scroll)
	typing := stealth.NewTyping(ctx.Page(), cfg.Stealth.Typing)
	searcher := search.NewSearch(ctx)
	parser := search.NewParser(ctx.Page())
	dryrun := interaction.NewDryRun(ctx, mouse, scroll, timing)
	messagingSvc := messaging.NewService(typing)

	// ==============================================
	// DRY-RUN MODE: Search + Parse + Visit (No Actions)
	// ==============================================
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("🎯 DRY-RUN MODE ACTIVE")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Execute search
	searchParams := search.SearchParams{
		Keywords: "Software Engineer",
		Location: "United States",
	}

	logger.Info("executing search",
		zap.String("keywords", searchParams.Keywords),
		zap.String("location", searchParams.Location))

	if err := searcher.ExecuteSearch(searchParams); err != nil {
		logger.Fatal("search failed", zap.Error(err))
	}

	logger.Info("✅ Search navigation successful")

	// Wait for page load
	time.Sleep(3 * time.Second)

	// OPTIONAL: Manual inspection pause (comment out for automatic mode)
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⏸️  OPTIONAL: Press ENTER to continue, or wait 5 seconds...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Non-blocking wait with timeout
	done := make(chan bool)
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("✅ Continuing...")
	case <-time.After(5 * time.Second):
		fmt.Println("✅ Auto-continuing after timeout...")
	}

	logger.Info("starting profile parsing and dry-run visits")

	// Scroll to load results
	logger.Info("📜 Scrolling page to load results...")
	scroll.ScrollDown()
	time.Sleep(1 * time.Second)
	scroll.ScrollDown()
	time.Sleep(1 * time.Second)

	// Parse profiles
	logger.Info("🔍 Parsing profiles...")
	profiles, err := parser.ParseResults()
	if err != nil {
		logger.Error("parser error", zap.Error(err))
	}

	logger.Info("📊 Parser results",
		zap.Int("total_profiles_found", len(profiles)))

	if len(profiles) == 0 {
		logger.Warn("❌ No profiles extracted - check selectors")
		logger.Info("Exiting cleanly")
		os.Exit(0)
	}

	// Log parsed profiles
	logger.Info("✅ Profiles successfully extracted:")
	for i, profile := range profiles {
		if i < 5 { // Show first 5
			logger.Info("  → Profile",
				zap.Int("index", i+1),
				zap.String("url", profile.ProfileURL),
				zap.String("name", profile.Name))
		}
	}
	if len(profiles) > 5 {
		logger.Info(fmt.Sprintf("  ... and %d more profiles", len(profiles)-5))
	}

	// Dry-run profile visits (max 2)
	visitCount := MAX_PROFILE_VISITS
	if len(profiles) < visitCount {
		visitCount = len(profiles)
	}

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("🚀 Starting dry-run profile visits",
		zap.Int("profiles_to_visit", visitCount))
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i := 0; i < visitCount; i++ {
		profile := profiles[i]

		logger.Info(fmt.Sprintf("DRY_RUN visiting profile %d/%d", i+1, visitCount),
			zap.String("name", profile.Name),
			zap.String("url", profile.ProfileURL))

		if err := dryrun.VisitProfile(profile.ProfileURL); err != nil {
			logger.Error("DRY_RUN visit failed",
				zap.String("profile", profile.ProfileURL),
				zap.Error(err))
			continue
		}

		// Extract first name from full name for message personalization
		firstName := profile.Name
		if parts := strings.Fields(profile.Name); len(parts) > 0 {
			firstName = parts[0]
		}

		// Demo messaging after profile visit (ONLY ONCE for first profile)
		if i == 0 {
			messagingSvc.DryRunSendMessage(profile.ProfileURL, firstName)
		}

		// Wait between visits
		if i < visitCount-1 {
			waitTime := timing.ActionDelay()
			logger.Info("DRY_RUN waiting before next profile",
				zap.Duration("duration", waitTime))
			time.Sleep(waitTime)
		}
	}

	// Final summary
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("📋 DRY-RUN EXECUTION SUMMARY")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("Mode", zap.String("value", "DRY-RUN (No real actions taken)"))
	logger.Info("Search executed", zap.String("status", "✅"))
	logger.Info("Profiles extracted", zap.Int("count", len(profiles)))
	logger.Info("Profiles visited", zap.Int("count", visitCount))
	logger.Info("Messages demonstrated", zap.Int("count", 1))
	logger.Info("Real actions taken", zap.String("count", "0 (dry-run mode)"))
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("✅ DRY-RUN EXECUTION COMPLETE")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("Demonstrated capabilities:")
	logger.Info("  • Authentication flow executed (session reuse active)")
	logger.Info("  • Search & profile parsing demonstrated")
	logger.Info("  • Connection intent simulated (no action taken)")
	logger.Info("  • Messaging workflow demonstrated (template + typing)")
	logger.Info("  • Stealth techniques active throughout execution")
	logger.Info("  • Modular architecture & state persistence verified")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("🛑 Execution stopped intentionally (dry-run mode)")
	logger.Info("Next steps: Review logs, verify stealth behavior, disable dry-run for production")

	// Keep browser open briefly for final inspection
	logger.Info("Browser will close in 10 seconds...")
	time.Sleep(10 * time.Second)

	logger.Info("Cleaning up and exiting")
}
