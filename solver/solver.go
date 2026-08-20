package solver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Result struct {
	Token          string
	Sitekey        string
	Cookies        string
	Elapsed        time.Duration
	URL            string
	UserAgent      string
	ChallengeSolved bool
	FinalURL       string
	CFClearance    string
	Err            error
}

type Config struct {
	Proxy          string
	Headless       bool
	Timeout        time.Duration
	UserAgent      string
	ViewportWidth  int
	ViewportHeight int
	DisableImages  bool
	DisableJS      bool
}

func DefaultConfig() Config {
	return Config{
		Headless:       false,
		Timeout:        90 * time.Second,
		ViewportWidth:  1920,
		ViewportHeight: 1080,
		DisableImages:  false,
		DisableJS:      false,
	}
}

func launchBrowser(cfg Config) (*rod.Browser, error) {
	l := launcher.New().
		Headless(cfg.Headless).
		NoSandbox(true).
		Set("disable-gpu", "true").
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-dev-shm-usage", "true").
		Set("window-size", fmt.Sprintf("%d,%d", cfg.ViewportWidth, cfg.ViewportHeight)).
		Set("disable-infobars", "true").
		Set("no-first-run", "true").
		Set("disable-extensions", "true").
		Set("disable-background-networking", "true").
		Set("disable-sync", "true").
		Set("disable-default-apps", "true").
		Set("disable-component-extensions-with-background-pages", "true").
		Set("disable-popup-blocking", "true").
		Set("disable-hang-monitor", "true").
		Set("disable-prompt-on-repost", "true").
		Set("disable-breakpad", "true").
		Set("disable-client-side-phishing-detection", "true").
		Set("disable-component-update", "true").
		Set("disable-domain-reliability", "true").
		Set("disable-features", "IsolateOrigins,site-per-process,TranslateUI").
		Set("enable-features", "NetworkService,NetworkServiceInProcess").
		Set("metrics-recording-only", "true").
		Set("no-default-browser-check", "true").
		Set("password-store", "basic").
		Set("use-mock-keychain", "true").
		Set("export-tagged-pdf", "false").
		Set("disable-search-engine-choice-screen", "true")

	if cfg.Proxy != "" {
		l = l.Proxy(cfg.Proxy)
	}

	if cfg.UserAgent != "" {
		l = l.Set("user-agent", cfg.UserAgent)
	}

	u, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}

	return browser, nil
}

const stealthScript = `
// Remove webdriver flag
Object.defineProperty(navigator, 'webdriver', {get: () => undefined});

// Add chrome runtime
window.chrome = { runtime: {}, loadTimes: function(){}, csi: function(){} };

// Set languages
Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']});

// Override plugins to look like a real browser
Object.defineProperty(navigator, 'plugins', {
	get: () => {
		const plugins = [
			{name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format'},
			{name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: ''},
			{name: 'Native Client', filename: 'internal-nacl-plugin', description: ''}
		];
		plugins.item = (i) => plugins[i];
		plugins.namedItem = (name) => plugins.find(p => p.name === name) || null;
		plugins.refresh = () => {};
		return Object.setPrototypeOf(plugins, PluginArray.prototype);
	}
});

// Override permissions query
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.query = (parameters) => (
	parameters.name === 'notifications' ?
		Promise.resolve({ state: Notification.permission }) :
		originalQuery(parameters)
);

// Hide automation-related properties
const getParameter = WebGLRenderingContext.prototype.getParameter;
WebGLRenderingContext.prototype.getParameter = function(parameter) {
	if (parameter === 37445) return 'Google Inc. (Intel)';
	if (parameter === 37446) return 'ANGLE (Intel, Mesa Intel(R) UHD Graphics 630 (CFL GT2), OpenGL 4.6)';
	return getParameter.call(this, parameter);
};

// Fix iframe contentWindow detection
const originalAttachShadow = Element.prototype.attachShadow;
Element.prototype.attachShadow = function() {
	return originalAttachShadow.call(this, ...arguments);
};

// Override toString for modified functions
const nativeToString = Function.prototype.toString;
Function.prototype.toString = function() {
	if (this === navigator.permissions.query) return 'function query() { [native code] }';
	if (this === Function.prototype.toString) return 'function toString() { [native code] }';
	return nativeToString.call(this);
};

// Set proper screen properties
Object.defineProperty(screen, 'width', {get: () => 1920});
Object.defineProperty(screen, 'height', {get: () => 1080});
Object.defineProperty(screen, 'availWidth', {get: () => 1920});
Object.defineProperty(screen, 'availHeight', {get: () => 1040});
Object.defineProperty(screen, 'colorDepth', {get: () => 24});
Object.defineProperty(screen, 'pixelDepth', {get: () => 24});

// Set connection properties
Object.defineProperty(navigator, 'connection', {
	get: () => ({
		effectiveType: '4g',
		rtt: 50,
		downlink: 10,
		saveData: false,
		addEventListener: () => {},
		removeEventListener: () => {},
	})
});

// Hardware concurrency
Object.defineProperty(navigator, 'hardwareConcurrency', {get: () => 8});

// Device memory
Object.defineProperty(navigator, 'deviceMemory', {get: () => 8});

// Platform
Object.defineProperty(navigator, 'platform', {get: () => 'Win32'});

// Vendor
Object.defineProperty(navigator, 'vendor', {get: () => 'Google Inc.'});

// Max touch points
Object.defineProperty(navigator, 'maxTouchPoints', {get: () => 0});

// Override getBattery to look realistic
if (navigator.getBattery) {
	navigator.getBattery = () => Promise.resolve({
		charging: true,
		chargingTime: 0,
		dischargingTime: Infinity,
		level: 1,
		addEventListener: () => {},
		removeEventListener: () => {},
	});
}
`

const turnstileInteractScript = `
// Try to find and click the turnstile checkbox inside iframes
function tryClickTurnstile() {
	var iframes = document.querySelectorAll('iframe[src*="challenges.cloudflare.com"]');
	for (var i = 0; i < iframes.length; i++) {
		try {
			var iframeDoc = iframes[i].contentDocument;
			if (iframeDoc) {
				var checkbox = iframeDoc.querySelector('input[type="checkbox"]');
				if (checkbox && !checkbox.checked) {
					checkbox.click();
					return true;
				}
				var label = iframeDoc.querySelector('label');
				if (label) {
					label.click();
					return true;
				}
			}
		} catch(e) {
			// Cross-origin, try clicking the iframe itself
			var rect = iframes[i].getBoundingClientRect();
			var centerX = rect.left + rect.width / 2;
			var centerY = rect.top + rect.height / 2;
			var el = document.elementFromPoint(centerX, centerY);
			if (el) {
				el.click();
				return true;
			}
		}
	}
	return false;
}
tryClickTurnstile();
`

func waitForToken(ctx context.Context, pg *rod.Page, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		token, err := pg.Eval(`() => {
			// Method 1: Standard turnstile hidden input
			var input = document.querySelector('input[name="cf-turnstile-response"]');
			if (input && input.value && input.value.length > 10) return input.value;

			// Method 2: Any input with cf-turnstile in the id
			var inputs = document.querySelectorAll('input[id*="cf-turnstile"]');
			for (var i = 0; i < inputs.length; i++) {
				if (inputs[i].value && inputs[i].value.length > 10) return inputs[i].value;
			}

			// Method 3: Any input with turnstile in the name
			var allInputs = document.querySelectorAll('input[name*="turnstile"]');
			for (var j = 0; j < allInputs.length; j++) {
				if (allInputs[j].value && allInputs[j].value.length > 10) return allInputs[j].value;
			}

			// Method 4: Check for turnstile token in window object
			if (window.turnstileToken) return window.turnstileToken;

			// Method 5: Check for cf_turnstile_response in form data
			var formInputs = document.querySelectorAll('form input');
			for (var k = 0; k < formInputs.length; k++) {
				var name = formInputs[k].name || formInputs[k].id || '';
				if (name.indexOf('turnstile') !== -1 || name.indexOf('cf-chl') !== -1) {
					if (formInputs[k].value && formInputs[k].value.length > 10) return formInputs[k].value;
				}
			}

			return '';
		}`)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if token != nil && token.Value.Str() != "" {
			return token.Value.Str(), nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for turnstile token")
}

func isManagedChallengePage(pg *rod.Page) bool {
	result, err := pg.Eval(`() => {
		var title = document.title || '';
		var h2 = document.querySelector('h2');
		var h2Text = h2 ? h2.textContent : '';
		var hasTurnstileScript = !!document.querySelector('script[src*="challenges.cloudflare.com/turnstile"]');
		var hasCfChlOpt = !!window._cf_chl_opt;
		var hasCfTurnstileResponse = !!document.querySelector('input[name="cf-turnstile-response"]') ||
				!!document.querySelector('input[id*="cf-chl-widget"]');
		var bodyText = document.body ? document.body.innerText : '';

		return {
			title: title,
			h2Text: h2Text,
			hasTurnstileScript: hasTurnstileScript,
			hasCfChlOpt: hasCfChlOpt,
			hasCfTurnstileResponse: hasCfTurnstileResponse,
			isChallenge: title.indexOf('Just a moment') !== -1 ||
					h2Text.indexOf('security') !== -1 ||
					h2Text.indexOf('verification') !== -1 ||
					h2Text.indexOf('Checking') !== -1 ||
					bodyText.indexOf('Verify you are human') !== -1 ||
					hasCfChlOpt
		};
	}`)
	if err != nil {
		return false
	}
	if result == nil {
		return false
	}
	isChallenge, _ := result.Value.Get("isChallenge").Bool()
	return isChallenge
}

func waitForChallengeCompletion(ctx context.Context, pg *rod.Page, timeout time.Duration) (*Result, error) {
	deadline := time.Now().Add(timeout)
	var lastToken string

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check if we've been redirected past the challenge
		result, err := pg.Eval(`() => {
			var currentURL = window.location.href;
			var title = document.title || '';

			// Get token from any turnstile input
			var token = '';
			var inputs = document.querySelectorAll('input[name*="turnstile"], input[id*="cf-chl-widget"]');
			for (var i = 0; i < inputs.length; i++) {
				if (inputs[i].value && inputs[i].value.length > 10) {
					token = inputs[i].value;
					break;
				}
			}

			// Check for success indicators
			var waitingText = document.querySelector('#wGhx6');
			var isWaiting = waitingText && waitingText.style.display !== 'none';

			// Check if the challenge page is gone (redirected to actual content)
			var isChallengePage = title === 'Just a moment...' ||
					!!document.querySelector('input[name="cf-turnstile-response"]') ||
					!!window._cf_chl_opt;

			// Check for cf_clearance cookie indicator
			var hasCfClearance = document.cookie.indexOf('cf_clearance') !== -1;

			return {
				url: currentURL,
				title: title,
				token: token,
				isWaiting: isWaiting,
				isChallengePage: isChallengePage,
				hasCfClearance: hasCfClearance
			};
		}`)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if result == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		url, _ := result.Value.Get("url").String()
		title, _ := result.Value.Get("title").String()
		token, _ := result.Value.Get("token").String()
		isWaiting, _ := result.Value.Get("isWaiting").Bool()
		isChallengePage, _ := result.Value.Get("isChallengePage").Bool()
		hasCfClearance, _ := result.Value.Get("hasCfClearance").Bool()

		if token != "" {
			lastToken = token
		}

		// If we have a token and we're either: waiting for site response, or no longer on challenge page
		if lastToken != "" && (isWaiting || !isChallengePage || hasCfClearance) {
			return &Result{
				Token:          lastToken,
				ChallengeSolved: true,
				FinalURL:       url,
			}, nil
		}

		// If the page has changed (redirected), challenge is likely done
		if !isChallengePage && lastToken != "" {
			return &Result{
				Token:          lastToken,
				ChallengeSolved: true,
				FinalURL:       url,
			}, nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Return whatever token we found even if challenge didn't fully complete
	if lastToken != "" {
		url := ""
		if result != nil {
			url, _ = result.Value.Get("url").String()
		}
		return &Result{
			Token:          lastToken,
			ChallengeSolved: false,
			FinalURL:       url,
		}, nil
	}

	return nil, fmt.Errorf("timeout waiting for challenge completion")
}

func getCookies(pg *rod.Page) string {
	cookies, err := proto.NetworkGetAllCookies{}.Call(pg)
	if err != nil {
		return ""
	}
	var parts []string
	for _, c := range cookies.Cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

func getCFClearance(pg *rod.Page) string {
	cookies, err := proto.NetworkGetAllCookies{}.Call(pg)
	if err != nil {
		return ""
	}
	for _, c := range cookies.Cookies {
		if c.Name == "cf_clearance" {
			return c.Value
		}
	}
	return ""
}

func Solve(ctx context.Context, url string, cfg Config) (*Result, error) {
	start := time.Now()

	browser, err := launchBrowser(cfg)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}
	defer browser.MustClose()

	browser.MustIgnoreCertErrors(true)

	pg := browser.MustPage()
	pg.MustEvalOnNewDocument(stealthScript)

	// Set extra headers to look more like a real browser
	proto.NetworkSetExtraHTTPHeaders{
		Headers: proto.NetworkHeaders{
			"Accept-Language": proto.String("en-US,en;q=0.9"),
			"sec-ch-ua":          proto.String(`"Google Chrome";v="138", "Chromium";v="138", "Not_A Brand";v="24"`),
			"sec-ch-ua-mobile":  proto.String("?0"),
			"sec-ch-ua-platform": proto.String(`"Windows"`),
		},
	}.Call(pg)

	err = pg.Navigate(url)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	err = pg.WaitLoad()
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	// Wait a moment for the page to fully render
	time.Sleep(2 * time.Second)

	// Detect if this is a managed challenge page
	if isManagedChallengePage(pg) {
		// For managed challenges, try to interact with the turnstile widget
		// The turnstile is in explicit render mode, need to wait for it to load
		time.Sleep(2 * time.Second)

		// Try clicking the turnstile checkbox
		pg.Eval(turnstileInteractScript)

		// Wait for challenge completion
		challengeResult, err := waitForChallengeCompletion(ctx, pg, cfg.Timeout-10*time.Second)
		if err != nil {
			return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
		}

		cookieStr := getCookies(pg)
		cfClearance := getCFClearance(pg)

		ua := ""
		res, err := proto.RuntimeEvaluate{Expression: "navigator.userAgent"}.Call(pg)
		if err == nil && res.Result != nil {
			ua = res.Result.Value.String()
		}

		finalURL := challengeResult.FinalURL
		if finalURL == "" {
			finalURL = url
		}

		challengeResult.Cookies = cookieStr
		challengeResult.CFClearance = cfClearance
		challengeResult.UserAgent = ua
		challengeResult.URL = url
		challengeResult.Elapsed = time.Since(start)

		return challengeResult, nil
	}

	// Standard turnstile page (not managed challenge)
	token, err := waitForToken(ctx, pg, cfg.Timeout-5*time.Second)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	cookieStr := getCookies(pg)
	cfClearance := getCFClearance(pg)

	ua := ""
	res, err := proto.RuntimeEvaluate{Expression: "navigator.userAgent"}.Call(pg)
	if err == nil && res.Result != nil {
		ua = res.Result.Value.String()
	}

	return &Result{
		Token:          token,
		Cookies:        cookieStr,
		CFClearance:    cfClearance,
		Elapsed:        time.Since(start),
		URL:            url,
		UserAgent:      ua,
		ChallengeSolved: token != "",
	}, nil
}

func SolveWithWidget(ctx context.Context, url, sitekey string, cfg Config) (*Result, error) {
	start := time.Now()

	browser, err := launchBrowser(cfg)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}
	defer browser.MustClose()

	browser.MustIgnoreCertErrors(true)

	pg := browser.MustPage()
	pg.MustEvalOnNewDocument(stealthScript)

	widgetHTML := fmt.Sprintf(`<!DOCTYPE html><html><head><title>Turnstile</title>
<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
</head><body>
<div class="cf-turnstile" data-sitekey="%s" data-callback="onTurnstileSuccess"></div>
<script>
function onTurnstileSuccess(token) {
	window.turnstileToken = token;
	var input = document.querySelector('input[name="cf-turnstile-response"]');
	if (input) input.value = token;
}
</script>
</body></html>`, sitekey)

	err = pg.Navigate("data:text/html," + widgetHTML)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	err = pg.WaitLoad()
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	// Wait for turnstile to render
	time.Sleep(2 * time.Second)

	token, err := waitForToken(ctx, pg, cfg.Timeout-5*time.Second)
	if err != nil {
		return &Result{Err: err, Elapsed: time.Since(start), URL: url}, err
	}

	cookieStr := getCookies(pg)

	return &Result{
		Token:           token,
		Sitekey:         sitekey,
		Cookies:         cookieStr,
		Elapsed:         time.Since(start),
		URL:             url,
		ChallengeSolved: token != "",
	}, nil
}

type Pool struct {
	workers int
	config  Config
	results chan *Result
	tasks   chan SolveTask
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

type SolveTask struct {
	URL     string
	Sitekey string
	ID      int
}

func NewPool(ctx context.Context, workers int, cfg Config) *Pool {
	pctx, pcancel := context.WithCancel(ctx)
	return &Pool{
		workers: workers,
		config:  cfg,
		results: make(chan *Result, workers*2),
		tasks:   make(chan SolveTask, workers*10),
		ctx:     pctx,
		cancel:  pcancel,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for task := range p.tasks {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		if task.Sitekey != "" {
			result, err := SolveWithWidget(p.ctx, task.URL, task.Sitekey, p.config)
			if err != nil {
				result.Err = fmt.Errorf("worker %d: %w", id, err)
			}
			result.URL = task.URL
			p.results <- result
		} else {
			result, err := Solve(p.ctx, task.URL, p.config)
			if err != nil {
				result.Err = fmt.Errorf("worker %d: %w", id, err)
			}
			result.URL = task.URL
			p.results <- result
		}
	}
}

func (p *Pool) Submit(task SolveTask) {
	p.tasks <- task
}

func (p *Pool) Results() <-chan *Result {
	return p.results
}

func (p *Pool) Close() {
	close(p.tasks)
	p.wg.Wait()
}

func (p *Pool) SolveAll(urls []string) []*Result {
	p.Start()
	for i, url := range urls {
		p.Submit(SolveTask{URL: url, ID: i})
	}
	p.Close()
	close(p.results)

	var results []*Result
	for r := range p.results {
		results = append(results, r)
	}
	return results
}
