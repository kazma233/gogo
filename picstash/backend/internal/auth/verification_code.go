package auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type VerificationCodeService struct {
	db *sql.DB
}

func NewVerificationCodeService(db *sql.DB) *VerificationCodeService {
	return &VerificationCodeService{db: db}
}

func (s *VerificationCodeService) GenerateAndSave(email, ipAddress string, expiresIn int) (string, error) {
	code := GenerateCode()

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Minute)

	_, err := s.db.Exec(`INSERT INTO verification_codes (email, code, expires_at, ip_address) VALUES (?, ?, ?, ?)`, email, code, expiresAt, ipAddress)
	if err != nil {
		return "", fmt.Errorf("保存验证码失败: %w", err)
	}

	slog.Info("验证码已生成", "email", email, "expires_in", expiresIn)

	return code, nil
}

func (s *VerificationCodeService) Verify(email, code string) (bool, error) {
	var expiresAt time.Time
	var usedAt *time.Time

	err := s.db.QueryRow(`
		SELECT expires_at, used_at
		FROM verification_codes
		WHERE email = ? AND code = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, email, code).Scan(&expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("查询验证码失败: %w", err)
	}

	if time.Now().After(expiresAt) {
		return false, nil
	}

	if usedAt != nil {
		return false, nil
	}

	_, err = s.db.Exec(`UPDATE verification_codes SET used_at = ? WHERE email = ? AND code = ?`, time.Now(), email, code)
	if err != nil {
		return false, fmt.Errorf("标记验证码已使用失败: %w", err)
	}

	return true, nil
}

func (s *VerificationCodeService) CanSendCode(email string) (bool, error) {
	var lastSentAt time.Time

	err := s.db.QueryRow(`
		SELECT created_at
		FROM verification_codes
		WHERE email = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, email).Scan(&lastSentAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("查询验证码失败: %w", err)
	}

	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	if lastSentAt.After(oneMinuteAgo) {
		return false, nil
	}

	return true, nil
}

func (s *VerificationCodeService) CleanupExpired() error {
	result, err := s.db.Exec(`DELETE FROM verification_codes WHERE expires_at < ?`, time.Now())
	if err != nil {
		return fmt.Errorf("清理过期验证码失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		slog.Info("清理过期验证码", "count", rows)
	}

	return nil
}
