package pkg

func DefaultSensBans() []string {
	return []string{
		"password",
		"passwd",
		"pwd",
		"pass",
		"pin",
		"pincode",
		"pin_code",
		"otp",
		"one_time_password",
		"mfa",
		"totp",
		"2fa",
		"secret",
		"token",
		"api_key",
		"apikey",
		"api-key",
		"api_token",
		"apitoken",
		"api-token",
		"access_token",
		"access_key",
		"auth_token",
		"auth_key",
		"credentials",
		"credit_card",
		"creditcard",
		"card_number",
		"cardnumber",
		"cvv",
		"cvc",
		"ssn",
		"social_security",
		"private_key",
		"privatekey",
		"secret_key",
		"secretkey",
		"encryption_key",
		"bearer",
		"authorization",
		"session_id",
		"sessionid",
		"cookie",
		"jwt",
		"oauth_token",
		"refresh_token",
		"client_secret",
		"client_id",
		"aws_access_key",
		"aws_secret_key",
		"aws_session_token",
	}
}

func DefaultTokensPatterns() []string {
	return []string{
		// GitHub tokens
		`gh[pousr]_[0-9a-zA-Z]{36}`,
		// GitLab tokens
		`glpat-[0-9a-zA-Z\-_]{20,}`,
		// AWS Access Key
		`AKIA[0-9A-Z]{16}`,
		`ABIA[0-9A-Z]{16}`,
		`ACCA[0-9A-Z]{16}`,
		`ASIA[0-9A-Z]{16}`,
		// OpenAI API key
		`sk-[0-9a-zA-Z]{48}`,
		// Stripe keys
		`sk_live_[0-9a-zA-Z]{24,}`,
		`pk_live_[0-9a-zA-Z]{24,}`,
		`rk_live_[0-9a-zA-Z]{24,}`,
		// Square tokens
		`sq0csp-[0-9a-zA-Z\-_]{43}`,
		`sq0atp-[0-9a-zA-Z\-_]{22}`,
		// Slack tokens
		`xoxb-[0-9]{10,}-[0-9a-zA-Z]{24,}`,
		`xoxp-[0-9]{10,}-[0-9]{10,}-[0-9a-zA-Z]{24,}`,
		`xoxa-[0-9]{10,}-[0-9a-zA-Z]{24,}`,
		`xoxr-[0-9]{10,}-[0-9a-zA-Z]{24,}`,
		// SendGrid API key
		`SG\.[0-9a-zA-Z\-_]{22}\.[0-9a-zA-Z\-_]{43}`,
		// JWT token
		`\beyJ[0-9A-Za-z_-]{2,}\.[0-9A-Za-z_-]{10,}\.[0-9A-Za-z_-]{10,}\b`,
		// Generic Bearer token
		`[Bb]earer\s+[0-9a-zA-Z\-_.~+/]+=*`,
		// Generic hex secrets (32, 40, 64 chars)
		`(?i)(?:secret|token|key|password|passwd|pwd|api_key|apikey|access_token|auth_token|credentials|private_key|secret_key|encryption_key|client_secret)\s*[:=]\s*["\x60']?[0-9a-fA-F]{32,64}["\x60']?`,
		// Generic base64 secrets
		`(?i)(?:secret|token|key|password|passwd|pwd|api_key|apikey|access_token|auth_token|credentials|private_key|secret_key|encryption_key|client_secret)\s*[:=]\s*["\x60']?[A-Za-z0-9+/]{40,}={0,2}["\x60']?`,
		// Heroku API key
		`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
		// Google API key
		`AIza[0-9A-Za-z\-_]{35}`,
		// Google OAuth token
		`ya29\.[0-9A-Za-z\-_]+`,
		// Twilio API key
		`SK[0-9a-fA-F]{32}`,
		// Mailgun API key
		`key-[0-9a-zA-Z]{32}`,
		// NPM token
		`npm_[0-9a-zA-Z]{36}`,
		// PyPI token
		`pypi-[0-9a-zA-Z\-_]{100,}`,
		// Telegram Bot token
		`[0-9]{8,10}:[0-9A-Za-z_-]{35}`,
		// Discord Bot token
		`[MN][A-Za-z\d]{23,}\.[\w-]{6}\.[\w-]{27}`,
		// Azure storage key
		`(?i)DefaultEndpointsProtocol=https;AccountName=[^;]+;AccountKey=[A-Za-z0-9+/=]{88};`,
		// RSA private key
		`-----BEGIN\s(?:RSA\s)?PRIVATE\sKEY-----`,
		// SSH private key
		`-----BEGIN\sOPENSSH\sPRIVATE\sKEY-----`,
		// PGP private key
		`-----BEGIN\sPGP\sPRIVATE\sKEY\sBLOCK-----`,
	}
}
