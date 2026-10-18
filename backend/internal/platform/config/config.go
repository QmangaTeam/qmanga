package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GoogleClientID        string
	GoogleClientSecret    string
	Environment           string
	HTTPPort              string
	DatabaseURL           string
	S3Endpoint            string
	S3AccessKey           string
	S3SecretKey           string
	S3BucketMedia         string
	S3BucketSource        string
	S3UseSSL              bool
	SessionSecret         string
	CSRFSecret            string
	AppOrigin             string
	MailTransport         string
	SMTPAddress           string
	SMTPFrom              string
	SMTPUsername          string
	SMTPPassword          string
	OCIEmailCompartmentID string
	TrustedProxyHost      string
}

func Load() (Config, error) {
	environment := strings.TrimSpace(os.Getenv("APP_ENV"))
	portString := strings.TrimSpace(os.Getenv("PORT"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	s3Endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	s3AccessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY"))
	s3SecretKey := strings.TrimSpace(os.Getenv("S3_SECRET_KEY"))
	s3BucketMedia := strings.TrimSpace(os.Getenv("S3_BUCKET_MEDIA"))
	s3BucketSource := strings.TrimSpace(os.Getenv("S3_BUCKET_SOURCE"))
	s3UseSSLString := strings.TrimSpace(os.Getenv("S3_USE_SSL"))
	sessionSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	csrfSecret := strings.TrimSpace(os.Getenv("CSRF_SECRET"))
	appOrigin := strings.TrimSpace(os.Getenv("APP_ORIGIN"))
	mailTransport := strings.TrimSpace(os.Getenv("MAIL_TRANSPORT"))
	if mailTransport == "" {
		mailTransport = "smtp"
	}
	smtpAddress := strings.TrimSpace(os.Getenv("SMTP_ADDRESS"))
	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	ociEmailCompartmentID := strings.TrimSpace(os.Getenv("OCI_EMAIL_COMPARTMENT_ID"))
	trustedProxyHost := strings.TrimSpace(os.Getenv("TRUSTED_PROXY_HOST"))
	if smtpAddress == "" && environment != "production" && mailTransport == "smtp" {
		smtpAddress = "mailpit:1025"
	}
	if smtpFrom == "" && environment != "production" {
		smtpFrom = "no-reply@qmanga.local"
	}

	var validationErrors []string
	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	if (googleClientID == "") != (googleClientSecret == "") {
		validationErrors = append(validationErrors, "GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must both be configured or both absent")
	}
	if googleClientID != "" && (!strings.HasSuffix(googleClientID, ".apps.googleusercontent.com") || strings.ContainsAny(googleClientID, " /\\\r\n")) {
		validationErrors = append(validationErrors, "GOOGLE_CLIENT_ID must identify a Google Web OAuth client")
	}
	if googleClientSecret != "" && (len(googleClientSecret) < 16 || isPlaceholder(googleClientSecret)) {
		validationErrors = append(validationErrors, "GOOGLE_CLIENT_SECRET must be a configured server secret")
	}

	switch mailTransport {
	case "smtp":
		if _, _, parseError := net.SplitHostPort(smtpAddress); parseError != nil {
			validationErrors = append(validationErrors, "SMTP_ADDRESS must be host:port")
		}
	case "oci":
		if !strings.HasPrefix(ociEmailCompartmentID, "ocid1.") {
			validationErrors = append(validationErrors, "OCI_EMAIL_COMPARTMENT_ID must be an OCI compartment OCID")
		}
	default:
		validationErrors = append(validationErrors, "MAIL_TRANSPORT must be 'smtp' or 'oci'")
	}
	if parsed, parseError := mail.ParseAddress(smtpFrom); parseError != nil || parsed.Address != smtpFrom {
		validationErrors = append(validationErrors, "SMTP_FROM must be an email address")
	}

	if environment != "development" && environment != "test" && environment != "production" {
		validationErrors = append(validationErrors, "APP_ENV must be explicitly set to 'development', 'test', or 'production'")
	}
	if environment == "production" && trustedProxyHost == "" {
		validationErrors = append(validationErrors, "TRUSTED_PROXY_HOST is required in production")
	}

	if portString == "" {
		portString = "8080"
	}
	portNumber, parsePortErr := strconv.Atoi(portString)
	if parsePortErr != nil || portNumber < 1 || portNumber > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("PORT must be a valid integer between 1 and 65535, got %q", portString))
	}

	useSSL := false
	if s3UseSSLString != "" {
		parsedBool, parseBoolErr := strconv.ParseBool(s3UseSSLString)
		if parseBoolErr != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("S3_USE_SSL must be a valid boolean, got %q", s3UseSSLString))
		} else {
			useSSL = parsedBool
		}
	}

	if databaseURL == "" {
		validationErrors = append(validationErrors, "missing required DATABASE_URL")
	} else {
		parsedURL, parseURLErr := url.Parse(databaseURL)
		if parseURLErr != nil || (parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql") {
			validationErrors = append(validationErrors, "DATABASE_URL must be a valid PostgreSQL connection URL")
		} else {
			if parsedURL.User == nil {
				validationErrors = append(validationErrors, "DATABASE_URL must include username and password credentials")
			} else {
				passwordValue, hasPassword := parsedURL.User.Password()
				if !hasPassword || strings.TrimSpace(passwordValue) == "" {
					validationErrors = append(validationErrors, "DATABASE_URL must include a non-empty password")
				} else if isPlaceholder(passwordValue) {
					validationErrors = append(validationErrors, "DATABASE_URL contains an unconfigured placeholder password")
				}
			}
			if strings.TrimSpace(parsedURL.Host) == "" {
				validationErrors = append(validationErrors, "DATABASE_URL must include a host")
			}
		}
	}

	if s3Endpoint == "" {
		validationErrors = append(validationErrors, "missing required S3_ENDPOINT")
	}

	if s3AccessKey == "" {
		validationErrors = append(validationErrors, "missing required S3_ACCESS_KEY")
	} else if isPlaceholder(s3AccessKey) {
		validationErrors = append(validationErrors, "S3_ACCESS_KEY contains an unconfigured placeholder")
	}

	if s3SecretKey == "" {
		validationErrors = append(validationErrors, "missing required S3_SECRET_KEY")
	} else if isPlaceholder(s3SecretKey) {
		validationErrors = append(validationErrors, "S3_SECRET_KEY contains an unconfigured placeholder")
	}

	if s3BucketMedia == "" {
		s3BucketMedia = "qmanga-media"
	}
	if s3BucketSource == "" {
		s3BucketSource = "qmanga-source"
	}

	if len(sessionSecret) < 32 {
		validationErrors = append(validationErrors, "SESSION_SECRET must be at least 32 characters")
	} else if isPlaceholder(sessionSecret) {
		validationErrors = append(validationErrors, "SESSION_SECRET contains an unconfigured placeholder")
	}

	if len(csrfSecret) < 32 {
		validationErrors = append(validationErrors, "CSRF_SECRET must be at least 32 characters")
	} else if isPlaceholder(csrfSecret) {
		validationErrors = append(validationErrors, "CSRF_SECRET contains an unconfigured placeholder")
	}

	if appOrigin == "" {
		if environment == "production" {
			validationErrors = append(validationErrors, "missing required APP_ORIGIN in production")
		} else {
			appOrigin = "http://localhost:8080"
		}
	}

	if appOrigin != "" {
		originURL, parseOriginErr := url.ParseRequestURI(appOrigin)
		if parseOriginErr != nil || originURL.Host == "" || (originURL.Scheme != "http" && originURL.Scheme != "https") || originURL.User != nil || originURL.Path != "" || originURL.RawQuery != "" || originURL.Fragment != "" {
			validationErrors = append(validationErrors, fmt.Sprintf("APP_ORIGIN must be a valid URL with host, got %q", appOrigin))
		} else if environment == "production" && originURL.Scheme != "https" {
			validationErrors = append(validationErrors, fmt.Sprintf("APP_ORIGIN must use HTTPS in production, got %q", appOrigin))
		}
	}

	configInstance := Config{
		GoogleClientID:        googleClientID,
		GoogleClientSecret:    googleClientSecret,
		Environment:           environment,
		HTTPPort:              portString,
		DatabaseURL:           databaseURL,
		S3Endpoint:            s3Endpoint,
		S3AccessKey:           s3AccessKey,
		S3SecretKey:           s3SecretKey,
		S3BucketMedia:         s3BucketMedia,
		S3BucketSource:        s3BucketSource,
		S3UseSSL:              useSSL,
		SessionSecret:         sessionSecret,
		CSRFSecret:            csrfSecret,
		AppOrigin:             appOrigin,
		MailTransport:         mailTransport,
		SMTPAddress:           smtpAddress,
		SMTPFrom:              smtpFrom,
		SMTPUsername:          os.Getenv("SMTP_USERNAME"),
		SMTPPassword:          os.Getenv("SMTP_PASSWORD"),
		OCIEmailCompartmentID: ociEmailCompartmentID,
		TrustedProxyHost:      trustedProxyHost,
	}

	if len(validationErrors) > 0 {
		return configInstance, fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return configInstance, nil
}

func isPlaceholder(value string) bool {
	lowered := strings.ToLower(value)
	return strings.Contains(lowered, "replace_with_") ||
		strings.Contains(lowered, "placeholder") ||
		strings.Contains(lowered, "change_in_production")
}
